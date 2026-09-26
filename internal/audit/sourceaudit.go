// Package audit binds Skillfile draft sources to the existing assurance
// gates through the source-audit-v1 record (skillfile-sources §4).
//
// The machine-local source audit binds one frozen package identity, its
// context hash, the digest of the trusted machine policy it was audited
// under, the digest of the complete persisted evidence report, the audit
// timestamp, and the resulting decision. It supplements the registry
// attestation recorded in the install marker; it never replaces it and it
// is never itself a registry attestation:
//
//   - local inputs have no network registry identity, so no audit-record-v1
//     is forged for local content and no source-audit object authorizes a
//     network attestation;
//   - a missing, unreadable, malformed, stale, or mismatching report fails
//     instead of becoming an unattested success;
//   - the stored object is a tamper-evident record, not authority: every
//     binding is recomputed under the current trusted machine policy and
//     the live audit decision is enforced before any cache or compiler work.
package audit

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/relux-works/curator/internal/stateread"

	"github.com/relux-works/curator/internal/capabilities"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/protocoljson"
)

// SourceAuditSchemaVersion is the only source-audit wire version accepted.
const SourceAuditSchemaVersion = 1

// Source-audit package identity arms (source-types-v1). The arms are
// disjoint exactly like the lock identities they bind.
const (
	SourcePackageLocalSnapshot = "local-snapshot"
	SourcePackageNetworkGit    = "network-git"
	SourcePackageConfiguredGit = "configured-git"
)

// DefaultSourceAuditMaxAge bounds how long a stored binding stays fresh.
// An older binding is stale, never current: the mutating path re-runs the
// gates and re-issues, the read-only path refuses. The week matches the
// default registry snapshot rotation so both evidence families age out
// together.
const DefaultSourceAuditMaxAge = 7 * 24 * time.Hour

// sourceAuditClockSkew tolerates clock disagreement between the writer and
// the reader of a binding. Timestamps beyond now+skew are rejected as
// forged rather than fresh.
const sourceAuditClockSkew = 5 * time.Minute

var (
	sourceAuditSHA256RE = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
	sourceAuditHex40RE  = regexp.MustCompile(`^[0-9a-f]{40}$`)
	sourceAuditHex64RE  = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

// SourceCommit is a locked Git object bound by a source-audit package.
type SourceCommit struct {
	ObjectFormat string `json:"object_format"`
	Hex          string `json:"hex"`
}

// SourcePackage is the frozen package identity a source-audit object binds.
// Exactly one arm is populated; the commit pointer stays nil on the local
// arm so the marshalled shape matches the source-types-v1 arms exactly.
type SourcePackage struct {
	Kind       string        `json:"kind"`
	Snapshot   string        `json:"snapshot,omitempty"`
	Repository string        `json:"repository,omitempty"`
	Source     string        `json:"source,omitempty"`
	Commit     *SourceCommit `json:"commit,omitempty"`
	Directory  string        `json:"directory,omitempty"`
}

// sourcePackageArm is the closed wire shape of one package union arm,
// derived from source-types-v1.schema.json #/$defs/package (oneOf[0..2],
// each additionalProperties:false). allowed and required are identical for
// the three arms: any present foreign member — even null or empty — refuses,
// and every required member must be present, non-null, and of the declared
// JSON kind. Value patterns (digests, repository syntax, directory ".")
// stay in validate(); this table binds presence, closedness, and kind.
type sourcePackageArm struct {
	allowed  map[string]string
	required map[string]string
}

// sourcePackageArms is the schema-driven closed-shape table. A test reads
// the vendored accepted schemas and asserts this table matches their
// arms/required/additionalProperties, so schema drift fails the build
// instead of silently reopening a shape hole.
var sourcePackageArms = map[string]sourcePackageArm{
	SourcePackageLocalSnapshot: {
		allowed:  map[string]string{"kind": "string", "snapshot": "string"},
		required: map[string]string{"kind": "string", "snapshot": "string"},
	},
	SourcePackageNetworkGit: {
		allowed: map[string]string{
			"kind": "string", "repository": "string", "commit": "object", "directory": "string",
		},
		required: map[string]string{
			"kind": "string", "repository": "string", "commit": "object", "directory": "string",
		},
	},
	SourcePackageConfiguredGit: {
		allowed: map[string]string{
			"kind": "string", "source": "string", "commit": "object", "directory": "string",
		},
		required: map[string]string{
			"kind": "string", "source": "string", "commit": "object", "directory": "string",
		},
	},
}

// sourceAuditObjectMembers is the closed top-level shape of
// source-audit-v1.schema.json (required 7, additionalProperties:false).
var sourceAuditObjectMembers = map[string]string{
	"schema_version": "number", "package": "object", "content_sha256": "string",
	"policy_sha256": "string", "evidence_sha256": "string",
	"decision": "string", "created_at": "string",
}

// evidenceReportMembers is the closed top-level shape of the persisted
// evidence report. MarshalEvidenceReport always emits all 11 members, so
// all are required, non-null, and kind-checked; unknown members refuse.
var evidenceReportMembers = map[string]string{
	"schema_version": "number", "skill": "string", "package": "object",
	"content_sha256": "string", "findings": "array", "pinned": "boolean",
	"revoked": "boolean", "revocation": "string",
	"script_policy": "array", "assurance_policy": "array", "decision": "string",
}

// rawJSONKind classifies one JSON value by its wire kind without decoding
// into Go zero values (which erase null/empty distinctions).
func rawJSONKind(raw json.RawMessage) string {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return "unknown"
	}
	if string(trimmed) == "null" {
		return "null"
	}
	switch trimmed[0] {
	case '"':
		return "string"
	case '{':
		return "object"
	case '[':
		return "array"
	case 't', 'f':
		return "boolean"
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return "number"
	default:
		return "unknown"
	}
}

// validateSourceCommitRaw enforces the closed lockedCommit shape
// (object_format + hex, both non-null strings, no extras) on raw bytes.
func validateSourceCommitRaw(commitRaw json.RawMessage) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(commitRaw, &raw); err != nil {
		return fmt.Errorf("commit is malformed: %v", err)
	}
	for key := range raw {
		if key != "object_format" && key != "hex" {
			return fmt.Errorf("commit carries no %q member", key)
		}
	}
	for _, key := range []string{"object_format", "hex"} {
		encoded, ok := raw[key]
		if !ok {
			return fmt.Errorf("commit is incomplete: missing %q", key)
		}
		if rawJSONKind(encoded) == "null" {
			return fmt.Errorf("commit member %q is null", key)
		}
		if rawJSONKind(encoded) != "string" {
			return fmt.Errorf("commit member %q must be a string", key)
		}
	}
	return nil
}

// validateSourcePackageRaw enforces the package union by selected kind on
// raw bytes BEFORE decoding into Go types. Presence alone refuses: a
// foreign-arm member fails even when null or empty, because zero-value
// decoding would otherwise erase it.
func validateSourcePackageRaw(packageRaw json.RawMessage) error {
	if rawJSONKind(packageRaw) == "null" {
		return fmt.Errorf("package is null")
	}
	if rawJSONKind(packageRaw) != "object" {
		return fmt.Errorf("package must be an object")
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(packageRaw, &raw); err != nil {
		return fmt.Errorf("package is malformed: %v", err)
	}
	kindRaw, ok := raw["kind"]
	if !ok {
		return fmt.Errorf("package is incomplete: missing %q", "kind")
	}
	if rawJSONKind(kindRaw) == "null" {
		return fmt.Errorf("package member %q is null", "kind")
	}
	if rawJSONKind(kindRaw) != "string" {
		return fmt.Errorf("package member %q must be a string", "kind")
	}
	var kind string
	if err := json.Unmarshal(kindRaw, &kind); err != nil {
		return fmt.Errorf("package member %q is malformed: %v", "kind", err)
	}
	arm, ok := sourcePackageArms[kind]
	if !ok {
		return fmt.Errorf("unknown package kind %q", kind)
	}
	for key := range raw {
		if _, allowed := arm.allowed[key]; !allowed {
			return fmt.Errorf("package kind %q carries no %q member", kind, key)
		}
	}
	for key, want := range arm.required {
		encoded, ok := raw[key]
		if !ok {
			return fmt.Errorf("package kind %q is incomplete: missing %q", kind, key)
		}
		if rawJSONKind(encoded) == "null" {
			return fmt.Errorf("package member %q is null", key)
		}
		if got := rawJSONKind(encoded); got != want {
			return fmt.Errorf("package member %q must be %s", key, want)
		}
	}
	if commitRaw, ok := raw["commit"]; ok {
		if err := validateSourceCommitRaw(commitRaw); err != nil {
			return err
		}
	}
	return nil
}

// validateSourceAuditObjectRaw enforces the closed source-audit-v1 top-level
// shape on raw bytes: exactly the 7 required members, each non-null and of
// the declared kind, no unknown members.
func validateSourceAuditObjectRaw(raw map[string]json.RawMessage) error {
	for key := range raw {
		if _, ok := sourceAuditObjectMembers[key]; !ok {
			return fmt.Errorf("unknown member %q", key)
		}
	}
	for key, want := range sourceAuditObjectMembers {
		encoded, ok := raw[key]
		if !ok {
			return fmt.Errorf("missing %q", key)
		}
		if rawJSONKind(encoded) == "null" {
			return fmt.Errorf("member %q is null", key)
		}
		if got := rawJSONKind(encoded); got != want {
			return fmt.Errorf("member %q must be %s", key, want)
		}
	}
	return nil
}

// validateEvidenceReportRaw enforces the closed evidence-report shape on raw
// bytes: exactly the 11 required members, each non-null and of the declared
// kind, no unknown members, plus the closed package union by kind.
func validateEvidenceReportRaw(evidence []byte) (map[string]json.RawMessage, error) {
	// Same duplicate-key gate as the object path, on the original report
	// bytes before any lossy decoding. Covers top-level and nested
	// package/commit/report members recursively.
	if err := protocoljson.Validate(evidence); err != nil {
		return nil, fmt.Errorf("report is malformed: %v", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(evidence))
	var raw map[string]json.RawMessage
	if err := decoder.Decode(&raw); err != nil {
		return nil, fmt.Errorf("report is malformed: %v", err)
	}
	var extra json.RawMessage
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("report is malformed: trailing data after JSON document")
		}
		return nil, fmt.Errorf("report is malformed: trailing data: %v", err)
	}
	for key := range raw {
		if _, ok := evidenceReportMembers[key]; !ok {
			return nil, fmt.Errorf("report carries no %q member", key)
		}
	}
	for key, want := range evidenceReportMembers {
		encoded, ok := raw[key]
		if !ok {
			return nil, fmt.Errorf("report is incomplete: missing %q", key)
		}
		if rawJSONKind(encoded) == "null" {
			return nil, fmt.Errorf("report member %q is null", key)
		}
		if got := rawJSONKind(encoded); got != want {
			return nil, fmt.Errorf("report member %q must be %s", key, want)
		}
	}
	if err := validateSourcePackageRaw(raw["package"]); err != nil {
		return nil, err
	}
	return raw, nil
}

// validate enforces the disjoint source-types-v1 arms.
func (pkg SourcePackage) validate() error {
	switch pkg.Kind {
	case SourcePackageLocalSnapshot:
		if pkg.Snapshot == "" || !sourceAuditSHA256RE.MatchString(pkg.Snapshot) {
			return fmt.Errorf("local-snapshot package requires a sha256 snapshot digest")
		}
		if pkg.Repository != "" || pkg.Source != "" || pkg.Commit != nil || pkg.Directory != "" {
			return fmt.Errorf("local-snapshot package carries no Git fields")
		}
	case SourcePackageNetworkGit:
		if pkg.Repository == "" {
			return fmt.Errorf("network-git package requires a repository")
		}
		if pkg.Commit == nil {
			return fmt.Errorf("network-git package requires a commit")
		}
		if err := pkg.Commit.validate(); err != nil {
			return err
		}
		if pkg.Directory == "" {
			return fmt.Errorf("network-git package requires a directory")
		}
		if pkg.Snapshot != "" || pkg.Source != "" {
			return fmt.Errorf("network-git package carries no snapshot or configured source")
		}
	case SourcePackageConfiguredGit:
		if pkg.Source == "" {
			return fmt.Errorf("configured-git package requires a source")
		}
		if pkg.Commit == nil {
			return fmt.Errorf("configured-git package requires a commit")
		}
		if err := pkg.Commit.validate(); err != nil {
			return err
		}
		if pkg.Directory != "." {
			return fmt.Errorf("configured-git package directory must be \".\"")
		}
		if pkg.Snapshot != "" || pkg.Repository != "" {
			return fmt.Errorf("configured-git package carries no snapshot or repository")
		}
	default:
		return fmt.Errorf("unknown package kind %q", pkg.Kind)
	}
	return nil
}

// object projects the package into the CCJ-1 value domain. Only populated
// arm members appear, so the local shape carries no Git members.
func (pkg SourcePackage) object() map[string]any {
	document := map[string]any{"kind": pkg.Kind}
	if pkg.Snapshot != "" {
		document["snapshot"] = pkg.Snapshot
	}
	if pkg.Repository != "" {
		document["repository"] = pkg.Repository
	}
	if pkg.Source != "" {
		document["source"] = pkg.Source
	}
	if pkg.Commit != nil {
		document["commit"] = map[string]any{
			"hex":           pkg.Commit.Hex,
			"object_format": pkg.Commit.ObjectFormat,
		}
	}
	if pkg.Directory != "" {
		document["directory"] = pkg.Directory
	}
	return document
}

func (commit SourceCommit) validate() error {
	switch commit.ObjectFormat {
	case "sha1":
		if !sourceAuditHex40RE.MatchString(commit.Hex) {
			return fmt.Errorf("sha1 commit must be 40 lowercase hex")
		}
	case "sha256":
		if !sourceAuditHex64RE.MatchString(commit.Hex) {
			return fmt.Errorf("sha256 commit must be 64 lowercase hex")
		}
	default:
		return fmt.Errorf("commit object_format must be sha1 or sha256")
	}
	return nil
}

// SourceAudit is the source-audit-v1 wire object: the six bindings of
// skillfile-sources §4 (identity, context, policy, evidence, time,
// decision). It is a machine-local record, never a registry attestation.
type SourceAudit struct {
	SchemaVersion  int           `json:"schema_version"`
	Package        SourcePackage `json:"package"`
	ContentSHA256  string        `json:"content_sha256"`
	PolicySHA256   string        `json:"policy_sha256"`
	EvidenceSHA256 string        `json:"evidence_sha256"`
	Decision       string        `json:"decision"`
	CreatedAt      string        `json:"created_at"`
}

// createdTime parses the RFC3339 UTC timestamp. Spec timestamps end in Z.
func (object SourceAudit) createdTime() (time.Time, error) {
	if !strings.HasSuffix(object.CreatedAt, "Z") {
		return time.Time{}, fmt.Errorf("created_at must be an RFC3339 UTC timestamp")
	}
	parsed, err := time.Parse(time.RFC3339, object.CreatedAt)
	if err != nil {
		return time.Time{}, fmt.Errorf("created_at is malformed: %v", err)
	}
	return parsed, nil
}

// ParseSourceAudit decodes one wire object with closed-shape checks:
// unknown fields, missing bindings, wrong versions, non-enum decisions,
// malformed digests, non-UTC timestamps, and trailing data all fail. Exactly
// one complete JSON document is accepted: any trailing token after the first
// value refuses as malformed.
//
// The wire shape is validated on RAW bytes before decoding into Go types:
// the top-level object must carry exactly the 7 required members
// (non-null, declared kind, no extras) and the package union is validated
// by its selected kind arm with additionalProperties:false. Go zero values
// erase null/empty distinctions, so a foreign-arm member that is null or
// empty would otherwise disappear through decoding and projection.
func ParseSourceAudit(payload []byte) (SourceAudit, error) {
	// Core §1 duplicate-key rejection comes first: Go map/struct decoding
	// keeps only the last duplicate, so the closed raw shapes below would
	// otherwise admit a hidden foreign-arm member. Validate covers the full
	// document recursively (top level and nested) on the original bytes.
	if err := protocoljson.Validate(payload); err != nil {
		return SourceAudit{}, fmt.Errorf("source_audit_rejected: malformed: %v", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	var raw map[string]json.RawMessage
	if err := decoder.Decode(&raw); err != nil {
		return SourceAudit{}, fmt.Errorf("source_audit_rejected: malformed: %v", err)
	}
	var extra json.RawMessage
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return SourceAudit{}, fmt.Errorf("source_audit_rejected: malformed: trailing data after JSON document")
		}
		return SourceAudit{}, fmt.Errorf("source_audit_rejected: malformed: trailing data: %v", err)
	}
	if err := validateSourceAuditObjectRaw(raw); err != nil {
		return SourceAudit{}, fmt.Errorf("source_audit_rejected: malformed: %v", err)
	}
	if err := validateSourcePackageRaw(raw["package"]); err != nil {
		return SourceAudit{}, fmt.Errorf("source_audit_rejected: malformed: %v", err)
	}
	var object SourceAudit
	structDecoder := json.NewDecoder(bytes.NewReader(payload))
	structDecoder.DisallowUnknownFields()
	if err := structDecoder.Decode(&object); err != nil {
		return SourceAudit{}, fmt.Errorf("source_audit_rejected: malformed: %v", err)
	}
	if err := validateSourceAuditShape(object); err != nil {
		return SourceAudit{}, fmt.Errorf("source_audit_rejected: malformed: %w", err)
	}
	return object, nil
}

func validateSourceAuditShape(object SourceAudit) error {
	if object.SchemaVersion != SourceAuditSchemaVersion {
		return fmt.Errorf("schema_version must be %d", SourceAuditSchemaVersion)
	}
	if err := object.Package.validate(); err != nil {
		return fmt.Errorf("package: %w", err)
	}
	if !sourceAuditSHA256RE.MatchString(object.ContentSHA256) {
		return fmt.Errorf("content_sha256 is malformed")
	}
	if !sourceAuditSHA256RE.MatchString(object.PolicySHA256) {
		return fmt.Errorf("policy_sha256 is malformed")
	}
	if !sourceAuditSHA256RE.MatchString(object.EvidenceSHA256) {
		return fmt.Errorf("evidence_sha256 is malformed")
	}
	switch object.Decision {
	case DecisionAllow, DecisionWarn, DecisionBlock, DecisionRequirePin:
	default:
		return fmt.Errorf("decision %q is not recognized", object.Decision)
	}
	if _, err := object.createdTime(); err != nil {
		return err
	}
	return nil
}

// PolicyInputs are the trusted machine-policy inputs a source-audit policy
// binding covers. Audit mode/threshold, detector generations, revocations,
// and the effective script/assurance policy labels all enter the digest, so
// any policy change ages stored bindings out instead of silently
// re-authorizing old verdicts under new rules.
type PolicyInputs struct {
	AuditMode       string
	AuditFailOn     string
	Revocations     []string
	ScriptLabels    []string
	AssuranceLabels []string
}

// PolicyDigest binds the effective machine policy: SHA-256 over the CCJ-1
// of the policy inputs, in the schema digest shape.
func PolicyDigest(inputs PolicyInputs) (string, error) {
	revocations := append([]string(nil), inputs.Revocations...)
	document := map[string]any{
		"assurance": append([]string(nil), inputs.AssuranceLabels...),
		"audit": map[string]any{
			"fail_on":         inputs.AuditFailOn,
			"mode":            inputs.AuditMode,
			"prompt_version":  PromptVersion,
			"revocations":     revocations,
			"ruleset_version": RulesetVersion,
		},
		"script": append([]string(nil), inputs.ScriptLabels...),
	}
	if document["assurance"] == nil {
		document["assurance"] = []string{}
	}
	if document["script"] == nil {
		document["script"] = []string{}
	}
	canonical, err := protocoljson.MarshalCanonical(document)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(canonical)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

// EvidenceReport is the persisted complete audit report a source-audit
// object binds: findings, pins, revocations, and the effective
// script/assurance policy labels behind the verdict.
type EvidenceReport struct {
	SchemaVersion   int           `json:"schema_version"`
	Skill           string        `json:"skill"`
	Package         SourcePackage `json:"package"`
	ContentSHA256   string        `json:"content_sha256"`
	Findings        []Finding     `json:"findings"`
	Pinned          bool          `json:"pinned"`
	Revoked         bool          `json:"revoked"`
	Revocation      string        `json:"revocation,omitempty"`
	ScriptPolicy    []string      `json:"script_policy"`
	AssurancePolicy []string      `json:"assurance_policy"`
	Decision        string        `json:"decision"`
}

// MarshalEvidenceReport renders the canonical report bytes the evidence
// digest covers.
func MarshalEvidenceReport(report EvidenceReport) ([]byte, error) {
	findings := make([]any, 0, len(report.Findings))
	for _, finding := range report.Findings {
		findings = append(findings, map[string]any{
			"evidence":   finding.Evidence,
			"file":       finding.File,
			"id":         finding.ID,
			"severity":   finding.Severity,
			"verifiable": finding.Verifiable,
		})
	}
	return protocoljson.MarshalCanonical(map[string]any{
		"assurance_policy": append([]string(nil), report.AssurancePolicy...),
		"content_sha256":   report.ContentSHA256,
		"decision":         report.Decision,
		"findings":         findings,
		"package":          report.Package.object(),
		"pinned":           report.Pinned,
		"revocation":       report.Revocation,
		"revoked":          report.Revoked,
		"schema_version":   report.SchemaVersion,
		"script_policy":    append([]string(nil), report.ScriptPolicy...),
		"skill":            report.Skill,
	})
}

// EvidenceDigest is the SHA-256 of canonical report bytes, in schema shape.
func EvidenceDigest(canonical []byte) string {
	sum := sha256.Sum256(canonical)
	return "sha256:" + hex.EncodeToString(sum[:])
}

// SourceExpectation is the recomputed ground truth one stored object is
// checked against. Every field comes from live machine state — the lock,
// the trusted policy, the persisted report bytes, and a fresh audit run —
// never from the object under test. The Skill/Findings/Pinned/Revoked/
// Revocation/ScriptLabels/AssuranceLabels arms bind the persisted evidence
// report to the authoritative live verdict and policy labels; the stored
// bytes' own hash never authorizes them.
type SourceExpectation struct {
	Package         SourcePackage
	Content         string
	Policy          string
	Evidence        []byte
	Decision        string
	Now             time.Time
	MaxAge          time.Duration
	Skill           string
	Findings        []Finding
	Pinned          bool
	Revoked         bool
	Revocation      string
	ScriptLabels    []string
	AssuranceLabels []string
}

// sourceAuditError is the typed validation outcome for ValidateSourceAudit.
// Renewable and Cause are set only by the validator's own decision — never
// derived from diagnostic text — so untrusted report fields interpolated
// into the message cannot select renewal. Diagnostics may still quote
// report fields; classification reads only these fields via errors.As.
type sourceAuditError struct {
	msg       string
	Renewable bool
	Cause     string
}

func (e *sourceAuditError) Error() string { return e.msg }

// Renewal causes set by the validator. Only these two may renew.
const (
	auditCausePolicy = "policy"
	auditCauseStale  = "stale"
)

// rejectAuditf builds a non-renewable refusal. Every phase-1 binding uses
// it, so even a message quoting attacker-controlled report fields stays
// non-renewable.
func rejectAuditf(format string, args ...any) *sourceAuditError {
	return &sourceAuditError{msg: fmt.Sprintf(format, args...)}
}

// renewAuditf builds the only renewable outcomes: a trusted policy-label
// drift or an aged timestamp, decided by the validator after phase 1.
func renewAuditf(cause string, format string, args ...any) *sourceAuditError {
	return &sourceAuditError{msg: fmt.Sprintf(format, args...), Renewable: true, Cause: cause}
}

// ValidateSourceAudit checks the six §4 bindings in two strict phases. A
// stored object never overrides live state; the first mismatch fails.
//
// Phase 1 — NON-RENEWABLE bindings, evaluated completely first: identity,
// context, evidence integrity/completeness/trusted bindings, the stored
// decision against the live decision, and the forged-future time bound.
// Any failure here refuses on both paths and never overwrites.
//
// Phase 2 — RENEWABLE conditions, evaluated only after phase 1 passes: an
// aged timestamp or a trusted policy-label drift. Only these two may lead
// to re-issuance on the mutating path.
//
// The split is structural, not per-branch: adding a new renewable cause
// must extend phase 2, and no non-renewable check may move below it, so a
// renewable error can never hide another refusal. Evidence covers both the
// digest over the exact stored bytes and the completeness and trusted
// bindings of the decoded report: a stripped or wrong-skill report with a
// recomputed digest still fails because its contents do not match the live
// authoritative verdict and policy labels.
func ValidateSourceAudit(object SourceAudit, expected SourceExpectation) error {
	if err := validateSourceAuditShape(object); err != nil {
		return rejectAuditf("source_audit_rejected: malformed: %v", err)
	}
	if !packagesEqual(object.Package, expected.Package) {
		return rejectAuditf("source_audit_rejected: identity: object binds %+v, locked package is %+v", object.Package, expected.Package)
	}
	if !strings.EqualFold(object.ContentSHA256, expected.Content) {
		return rejectAuditf("source_audit_rejected: context: object binds %s, locked content is %s", object.ContentSHA256, expected.Content)
	}
	if EvidenceDigest(expected.Evidence) != strings.ToLower(object.EvidenceSHA256) {
		return rejectAuditf("source_audit_rejected: evidence: report digest mismatch")
	}
	// Closed report shape on RAW bytes before decoding into Go types: the
	// 11 required members must be present, non-null, and of the declared
	// kind with no unknown members, and the package union is validated by
	// its selected kind arm. Zero-value decoding would otherwise erase a
	// foreign-arm null/empty member or a null scalar.
	if _, err := validateEvidenceReportRaw(expected.Evidence); err != nil {
		return rejectAuditf("source_audit_rejected: evidence: %v", err)
	}
	var report EvidenceReport
	decoder := json.NewDecoder(bytes.NewReader(expected.Evidence))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&report); err != nil {
		return rejectAuditf("source_audit_rejected: evidence: report is malformed: %v", err)
	}
	var reportExtra json.RawMessage
	if err := decoder.Decode(&reportExtra); err != io.EOF {
		if err == nil {
			return rejectAuditf("source_audit_rejected: evidence: report is malformed: trailing data after JSON document")
		}
		return rejectAuditf("source_audit_rejected: evidence: report is malformed: trailing data: %v", err)
	}
	if report.SchemaVersion != 1 {
		return rejectAuditf("source_audit_rejected: evidence: report schema_version must be 1")
	}
	if report.Skill == "" || report.ContentSHA256 == "" || report.Decision == "" {
		return rejectAuditf("source_audit_rejected: evidence: report is malformed: missing skill, content, or decision")
	}
	if report.Findings == nil {
		return rejectAuditf("source_audit_rejected: evidence: report is incomplete: missing findings")
	}
	if report.ScriptPolicy == nil {
		return rejectAuditf("source_audit_rejected: evidence: report is incomplete: missing script_policy")
	}
	if report.AssurancePolicy == nil {
		return rejectAuditf("source_audit_rejected: evidence: report is incomplete: missing assurance_policy")
	}
	if err := report.Package.validate(); err != nil {
		return rejectAuditf("source_audit_rejected: evidence: report package is malformed: %v", err)
	}
	if !sourceAuditSHA256RE.MatchString(report.ContentSHA256) {
		return rejectAuditf("source_audit_rejected: evidence: report content_sha256 is malformed")
	}
	switch report.Decision {
	case DecisionAllow, DecisionWarn, DecisionBlock, DecisionRequirePin:
	default:
		return rejectAuditf("source_audit_rejected: evidence: report decision %q is not recognized", report.Decision)
	}
	for _, finding := range report.Findings {
		if finding.ID == "" || finding.Severity == "" || finding.Evidence == "" {
			return rejectAuditf("source_audit_rejected: evidence: report finding is malformed")
		}
		if _, known := severityRank[finding.Severity]; !known {
			return rejectAuditf("source_audit_rejected: evidence: report finding severity %q is not recognized", finding.Severity)
		}
	}
	if !strings.EqualFold(report.ContentSHA256, object.ContentSHA256) || !packagesEqual(report.Package, object.Package) {
		return rejectAuditf("source_audit_rejected: evidence: report binds another package")
	}
	if report.Decision != object.Decision {
		return rejectAuditf("source_audit_rejected: evidence: report decision %q differs from bound decision %q", report.Decision, object.Decision)
	}
	if expected.Skill != "" && report.Skill != expected.Skill {
		return rejectAuditf("source_audit_rejected: evidence: report skill %q differs from locked skill %q", report.Skill, expected.Skill)
	}
	if !findingsEqual(report.Findings, expected.Findings) {
		return rejectAuditf("source_audit_rejected: evidence: report findings differ from the live audit verdict")
	}
	if report.Pinned != expected.Pinned {
		return rejectAuditf("source_audit_rejected: evidence: report pin state differs from the live operator pins")
	}
	if report.Revoked != expected.Revoked || report.Revocation != expected.Revocation {
		return rejectAuditf("source_audit_rejected: evidence: report revocation differs from the live revocation state")
	}
	if !labelsEqual(report.ScriptPolicy, expected.ScriptLabels) {
		return rejectAuditf("source_audit_rejected: evidence: report script_policy differs from the effective script policy")
	}
	if !labelsEqual(report.AssurancePolicy, expected.AssuranceLabels) {
		return rejectAuditf("source_audit_rejected: evidence: report assurance_policy differs from the effective assurance policy")
	}
	// Phase 1 ends here: the stored decision must equal the live decision
	// before any renewable condition is considered. A forged decision
	// combined with a stale timestamp or a policy drift still refuses.
	if object.Decision != expected.Decision {
		return rejectAuditf("source_audit_rejected: decision: object claims %q, current policy decides %q", object.Decision, expected.Decision)
	}
	created, err := object.createdTime()
	if err != nil {
		return rejectAuditf("source_audit_rejected: time: %v", err)
	}
	maxAge := expected.MaxAge
	if maxAge <= 0 {
		maxAge = DefaultSourceAuditMaxAge
	}
	// A future timestamp is forged, never renewable: it stays in phase 1.
	if created.After(expected.Now.Add(sourceAuditClockSkew)) {
		return rejectAuditf("source_audit_rejected: time: created_at %s is in the future", object.CreatedAt)
	}
	// Phase 2 — renewable conditions only. Every non-renewable binding
	// above passed; nothing below may precede a phase-1 check. These are
	// the only renewAuditf sites: renewal is decided here, never inferred
	// from message text.
	if expected.Now.Sub(created) > maxAge {
		return renewAuditf(auditCauseStale, "source_audit_rejected: time: binding from %s is stale; re-run the gates", object.CreatedAt)
	}
	if !strings.EqualFold(object.PolicySHA256, expected.Policy) {
		return renewAuditf(auditCausePolicy, "source_audit_rejected: policy: trusted policy changed; re-run the gates")
	}
	return nil
}

// packagesEqual compares package identities by canonical value. The commit
// arm is a pointer so local shapes marshal exactly; pointer identity must
// not leak into equality.
func packagesEqual(left, right SourcePackage) bool {
	leftBytes, leftErr := protocoljson.MarshalCanonical(left.object())
	rightBytes, rightErr := protocoljson.MarshalCanonical(right.object())
	if leftErr != nil || rightErr != nil {
		return false
	}
	return bytes.Equal(leftBytes, rightBytes)
}

// findingsEqual compares live and stored findings. Nil and empty both mean
// "no findings" and compare equal; otherwise the verdict inputs must match
// exactly, in order.
func findingsEqual(stored, live []Finding) bool {
	if len(stored) == 0 && len(live) == 0 {
		return true
	}
	if len(stored) != len(live) {
		return false
	}
	for i := range stored {
		if stored[i] != live[i] {
			return false
		}
	}
	return true
}

// labelsEqual compares effective policy labels. Nil and empty both mean "no
// labels" and compare equal so minimal test bindings without labels still
// validate against empty stored label sets.
func labelsEqual(stored, live []string) bool {
	if len(stored) == 0 && len(live) == 0 {
		return true
	}
	if len(stored) != len(live) {
		return false
	}
	for i := range stored {
		if stored[i] != live[i] {
			return false
		}
	}
	return true
}

// isRenewableAuditError reports whether a validation failure is an
// authorized renewal the mutating path may re-issue after a fresh live run:
// a trusted policy change or an aged timestamp. Classification reads only
// the typed sourceAuditError outcome via errors.As — never message text —
// so renewal markers interpolated into untrusted report fields cannot
// promote a refusal. Every other failure is a broken existing record and
// refuses on both paths, never overwriting.
func isRenewableAuditError(err error) bool {
	if err == nil {
		return false
	}
	var typed *sourceAuditError
	if errors.As(err, &typed) {
		return typed.Renewable
	}
	return false
}

// auditRenewalCause returns the validator-assigned renewal cause ("policy"
// or "stale") for a renewable error, or "" otherwise. Test-only helper for
// asserting the typed outcome without parsing diagnostic text.
func auditRenewalCause(err error) string {
	var typed *sourceAuditError
	if errors.As(err, &typed) && typed.Renewable {
		return typed.Cause
	}
	return ""
}

// sourceAuditEntryPresence classifies one store entry without following
// symlinks, so a dangling binding link or a link loop can never look
// absent. Lstat inspects the entry itself: only Lstat reporting ENOENT
// proves absence. Any other outcome — an entry of any kind, or any
// non-absence inspection error (permission, loop, I/O) — keeps the slot
// non-vacant and first issuance must not proceed.
type sourceAuditEntryPresence int

const (
	sourceAuditEntryAbsent sourceAuditEntryPresence = iota
	sourceAuditEntryPresent
	sourceAuditEntryReadFailure
)

// probeSourceAuditEntry classifies one store path. A nil Lstat error (any
// file kind, including a symlink however broken) is present; ENOENT is
// absent; every other inspection error is a read failure. Callers treat
// present and read-failure identically (refuse); the split exists so a
// read failure can never collapse into absence.
func probeSourceAuditEntry(path string) sourceAuditEntryPresence {
	metadata, err := stateread.Lstat(path)
	if err == nil {
		if metadata.Kind == stateread.KindAbsent {
			return sourceAuditEntryAbsent
		}
		return sourceAuditEntryPresent
	}
	return sourceAuditEntryReadFailure
}

// sourceAuditBindingVacant reports whether both store slots are positively
// absent: Lstat → ENOENT for the object entry and likewise for the report
// entry. It distinguishes first-time issuance (no binding entry at all)
// from a broken existing record (any entry present but unreadable, or any
// inspection failure), which must refuse rather than re-establish. Any
// path-derivation failure is not proof of absence and keeps the binding
// non-vacant.
func sourceAuditBindingVacant(home string, pkg SourcePackage) bool {
	objectPath, err := SourceAuditPath(home, pkg)
	if err != nil {
		return false
	}
	evidencePath, err := SourceEvidencePath(home, pkg)
	if err != nil {
		return false
	}
	return probeSourceAuditEntry(objectPath) == sourceAuditEntryAbsent &&
		probeSourceAuditEntry(evidencePath) == sourceAuditEntryAbsent
}

// sourceAuditStoreKey derives the machine-private store key of one package
// identity. The key carries no trust: it only locates the record the live
// gates revalidate.
func sourceAuditStoreKey(pkg SourcePackage) (string, error) {
	canonical, err := protocoljson.MarshalCanonical(pkg.object())
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(canonical)
	return hex.EncodeToString(sum[:]), nil
}

// SourceAuditPath is the deterministic machine-private object path for one
// package identity. Bindings never enter the project tree.
func SourceAuditPath(home string, pkg SourcePackage) (string, error) {
	key, err := sourceAuditStoreKey(pkg)
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "source-audit", key+".json"), nil
}

// SourceEvidencePath is the deterministic machine-private report path next
// to the object.
func SourceEvidencePath(home string, pkg SourcePackage) (string, error) {
	key, err := sourceAuditStoreKey(pkg)
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "source-audit", key+".report.json"), nil
}

// StoreSourceAudit persists one issued binding and its evidence report. The
// report lands first so a crash can only leave an object without evidence —
// which fails closed as unavailable — never evidence without an object. The
// report bytes land verbatim: the evidence digest covers the exact stored
// bytes, so no framing may be added here.
func StoreSourceAudit(home string, object SourceAudit, evidence []byte) error {
	evidencePath, err := SourceEvidencePath(home, object.Package)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(evidencePath), 0o755); err != nil {
		return err
	}
	payload, err := protocoljson.MarshalCanonical(map[string]any{
		"content_sha256":  object.ContentSHA256,
		"created_at":      object.CreatedAt,
		"decision":        object.Decision,
		"evidence_sha256": object.EvidenceSHA256,
		"package":         object.Package.object(),
		"policy_sha256":   object.PolicySHA256,
		"schema_version":  object.SchemaVersion,
	})
	if err != nil {
		return err
	}
	objectPath, err := SourceAuditPath(home, object.Package)
	if err != nil {
		return err
	}
	if err := os.WriteFile(evidencePath, append([]byte(nil), evidence...), 0o644); err != nil {
		return err
	}
	return os.WriteFile(objectPath, append(payload, '\n'), 0o644)
}

// LoadSourceAudit reads one stored binding and its report. An absent pair
// fails as unavailable; a present but malformed object fails as rejected.
// Absence and corruption are different facts and stay distinguishable.
func LoadSourceAudit(home string, pkg SourcePackage) (SourceAudit, []byte, error) {
	objectPath, err := SourceAuditPath(home, pkg)
	if err != nil {
		return SourceAudit{}, nil, err
	}
	objectRead, err := stateread.ReadFile(objectPath)
	if err != nil {
		return SourceAudit{}, nil, fmt.Errorf("source_audit_unavailable: binding is unreadable: %w", err)
	}
	if objectRead.Kind == stateread.KindAbsent {
		return SourceAudit{}, nil, fmt.Errorf("source_audit_unavailable: no machine binding for this package; run an install first")
	}
	object, err := ParseSourceAudit(objectRead.Bytes)
	if err != nil {
		return SourceAudit{}, nil, err
	}
	evidencePath, err := SourceEvidencePath(home, pkg)
	if err != nil {
		return SourceAudit{}, nil, err
	}
	evidenceRead, err := stateread.ReadFile(evidencePath)
	if err != nil {
		return SourceAudit{}, nil, fmt.Errorf("source_audit_unavailable: audit report is unreadable: %w", err)
	}
	if evidenceRead.Kind == stateread.KindAbsent {
		return SourceAudit{}, nil, fmt.Errorf("source_audit_unavailable: audit report is absent; re-run the gates")
	}
	return object, evidenceRead.Bytes, nil
}

// SourceSubject is one draft member under source-audit validation.
type SourceSubject struct {
	Name            string
	Source          string // revocation identity: declared source
	Git             string // revocation identity: declared git URL
	Commit          string
	Snapshot        string // frozen tree for live detection
	SchemaVersion   int
	Capabilities    capabilities.Manifest
	Package         SourcePackage // locked identity
	ContentSHA256   string        // locked context hash
	ScriptLabels    []string
	AssuranceLabels []string
}

// CheckSourceAudit validates — or, on the mutating path, establishes — the
// source-audit binding of one draft member. It is a no-op when audit is
// disabled, mirroring Gate. The stored object is never authority: every
// binding is recomputed under the current trusted policy and the live
// decision is enforced. Returns gate warnings (warn verdicts) or a failure.
//
//   - first issuance (no binding at all) + mutating path: run the full live
//     pipeline, enforce the verdict, then issue and store the binding;
//   - first issuance + read-only path: fail unavailable — planning must not
//     invent trust, and missing evidence is unknown, never current;
//   - broken existing record (binding present but report missing,
//     unreadable, malformed, or mismatching): refuse with a source_audit_*
//     diagnostic on both paths, never overwriting;
//   - present binding: revalidate all six bindings; only an aged timestamp
//     or a trusted policy change renews through a fresh live run on the
//     mutating path, the read-only path refuses stale bindings.
//
// The live verdict is enforced before any stored state is consulted, so
// hostile content reports audit blocked even when a prior allow binding
// exists; a stored allow never authorizes new blocking content.
func CheckSourceAudit(cfg *config.Config, subject SourceSubject, persist bool, now time.Time) ([]string, error) {
	if !cfg.Audit.Enabled {
		return nil, nil
	}
	if !runStaticCanary() {
		return nil, fmt.Errorf("audit blocked: %s: audit canary failed: detectors are not producing expected findings", subject.Name)
	}
	policy, err := PolicyDigest(PolicyInputs{
		AuditMode:       cfg.Audit.Mode,
		AuditFailOn:     cfg.Audit.FailOn,
		Revocations:     cfg.Audit.Revocations,
		ScriptLabels:    subject.ScriptLabels,
		AssuranceLabels: subject.AssuranceLabels,
	})
	if err != nil {
		return nil, fmt.Errorf("audit blocked: %s: %v", subject.Name, err)
	}
	report, err := auditSubject(cfg, Subject{
		Name: subject.Name, Source: subject.Source, Git: subject.Git,
		Commit: subject.Commit, Snapshot: subject.Snapshot,
		SchemaVersion: subject.SchemaVersion, Capabilities: subject.Capabilities,
	}, persist)
	if err != nil {
		return nil, fmt.Errorf("audit blocked: %s: %v", subject.Name, err)
	}
	if err := enforceSourceDecision(subject.Name, report); err != nil {
		return nil, err
	}
	object, storedEvidence, loadErr := LoadSourceAudit(cfg.Home(), subject.Package)
	if loadErr != nil {
		if !isUnavailable(loadErr) {
			return nil, fmt.Errorf("%s: %v", subject.Name, loadErr)
		}
		if !persist {
			return nil, fmt.Errorf("%s: %v", subject.Name, loadErr)
		}
		// First issuance proceeds only on positive proof that no binding
		// entry exists (object and report both Lstat-absent). Any
		// unreadable or existing-but-broken entry refuses here, before
		// either record is written, as a typed non-renewable outcome —
		// never re-established through a dangling link target or over a
		// corrupt report.
		if !sourceAuditBindingVacant(cfg.Home(), subject.Package) {
			return nil, fmt.Errorf("%s: %w", subject.Name, rejectAuditf("source_audit_rejected: evidence: existing binding is missing or unreadable; re-run with fresh state: %v", loadErr))
		}
		return establishSourceAudit(cfg, subject, policy, report, now)
	}
	expected := SourceExpectation{
		Package: subject.Package, Content: subject.ContentSHA256,
		Policy: policy, Evidence: storedEvidence, Decision: report.Decision,
		Now: now, MaxAge: DefaultSourceAuditMaxAge,
		Skill: subject.Name, Findings: report.Findings,
		Pinned:  isPinned(cfg, report.ContentSHA256),
		Revoked: report.Revoked, Revocation: report.Revocation,
		ScriptLabels: subject.ScriptLabels, AssuranceLabels: subject.AssuranceLabels,
	}
	if err := ValidateSourceAudit(object, expected); err != nil {
		if !persist {
			return nil, fmt.Errorf("%s: %w", subject.Name, err)
		}
		if !isRenewableAuditError(err) {
			return nil, fmt.Errorf("%s: %w", subject.Name, err)
		}
		return establishSourceAudit(cfg, subject, policy, report, now)
	}
	return enforceSourceWarnings(report), nil
}

// establishSourceAudit enforces one fresh live verdict and stores its
// binding. It runs only on the mutating path, after the live pipeline.
func establishSourceAudit(cfg *config.Config, subject SourceSubject, policy string, report Report, now time.Time) ([]string, error) {
	if fail := enforceSourceDecision(subject.Name, report); fail != nil {
		return nil, fail
	}
	pinned := isPinned(cfg, report.ContentSHA256)
	evidence, err := MarshalEvidenceReport(EvidenceReport{
		SchemaVersion: 1, Skill: subject.Name, Package: subject.Package,
		ContentSHA256: subject.ContentSHA256, Findings: report.Findings,
		Pinned: pinned, Revoked: report.Revoked, Revocation: report.Revocation,
		ScriptPolicy: subject.ScriptLabels, AssurancePolicy: subject.AssuranceLabels,
		Decision: report.Decision,
	})
	if err != nil {
		return nil, fmt.Errorf("audit blocked: %s: %v", subject.Name, err)
	}
	object := SourceAudit{
		SchemaVersion: SourceAuditSchemaVersion, Package: subject.Package,
		ContentSHA256: subject.ContentSHA256, PolicySHA256: policy,
		EvidenceSHA256: EvidenceDigest(evidence), Decision: report.Decision,
		CreatedAt: now.UTC().Format(time.RFC3339),
	}
	if err := StoreSourceAudit(cfg.Home(), object, evidence); err != nil {
		return nil, fmt.Errorf("audit blocked: %s: %v", subject.Name, err)
	}
	return enforceSourceWarnings(report), nil
}

// enforceSourceDecision fails closed on block and require_pin verdicts,
// mirroring the gate refusal wording for the same outcome.
func enforceSourceDecision(skill string, report Report) error {
	switch report.Decision {
	case DecisionAllow, DecisionWarn:
		return nil
	case DecisionRequirePin:
		return fmt.Errorf(
			"audit requires pin: %s: content hash %s is not pinned with a reason",
			skill, report.ContentSHA256)
	default: // block, including revocations
		if report.Revoked {
			return fmt.Errorf("audit blocked: %s: %s is revoked", skill, report.Revocation)
		}
		if len(report.Findings) == 0 {
			return fmt.Errorf("audit blocked: %s", skill)
		}
		var errs []string
		for _, finding := range report.Findings {
			errs = append(errs, fmt.Sprintf("audit blocked: %s: %s %s - %s", skill, finding.Severity, finding.ID, finding.Evidence))
		}
		return fmt.Errorf("%s", strings.Join(errs, "; "))
	}
}

// enforceSourceWarnings renders warn-verdict findings as gate warnings.
func enforceSourceWarnings(report Report) []string {
	if report.Decision != DecisionWarn {
		return nil
	}
	var warnings []string
	for _, finding := range report.Findings {
		warnings = append(warnings, fmt.Sprintf("audit warning: %s: %s %s - %s", report.Skill, finding.Severity, finding.ID, finding.Evidence))
	}
	return warnings
}

func isUnavailable(err error) bool {
	return err != nil && strings.Contains(err.Error(), "source_audit_unavailable")
}
