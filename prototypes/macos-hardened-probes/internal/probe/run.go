package probe

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/relux-works/curator/prototypes/macos-hardened-probes/internal/evidence"
	"github.com/relux-works/curator/prototypes/macos-hardened-probes/internal/inside"
	"github.com/relux-works/curator/prototypes/macos-hardened-probes/internal/seatbelt"
	"github.com/relux-works/curator/prototypes/macos-hardened-probes/internal/spec"
)

// execCommand exists so tests can see one place where subprocesses are created.
func execCommand(ctx context.Context, name string, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, name, args...)
}

// Options configure a probe run.
type Options struct {
	// WorkDir is where the probe environment is built.
	WorkDir string
	// SelfPath is the probe binary, which doubles as the in-domain agent.
	SelfPath string
	// ForceUnavailable names classes to report unavailable regardless of what
	// the probes observed. It exists only to exercise the fail-closed boundary
	// and is recorded in the report so an injected result can never be mistaken
	// for a measured one.
	ForceUnavailable []string
	// Timeout bounds the whole run.
	Timeout time.Duration
}

// Result is everything one run produced.
type Result struct {
	Report   Report
	Evidence evidence.Record
	// ExitCode is the process exit status the harness should use.
	ExitCode int
}

// Exit codes. They are part of the harness contract: a caller distinguishes
// "the host cannot do this" from "the harness itself broke".
const (
	// ExitEstablished means every class is applied and every guarantee holds.
	ExitEstablished = 0
	// ExitRejected means the run completed and the host rejected: at least one
	// capability could not be established. This is the fail-closed outcome, and
	// it is the expected result on an unqualified platform.
	ExitRejected = 1
	// ExitHarnessError means the harness could not produce a trustworthy
	// record. It is distinct from ExitRejected because an unusable harness is
	// not evidence about the host.
	ExitHarnessError = 2
)

// Run executes every class probe and returns the evidence record.
//
// The run rejects rather than reports success whenever it is not certain: an
// unsupported platform, a probe that could not execute, a record that fails its
// own validation. There is no partial mode.
func Run(ctx context.Context, opts Options) (Result, error) {
	started := time.Now()
	report := Report{
		ReportVersion: ReportVersion,
		SpecRevision:  spec.SpecRevision,
		Platform:      spec.PlatformMacOS,
		Backend:       spec.BackendMacOSSandbox,
		Host:          DescribeHost(),
		StartedAt:     started.UTC().Format(time.RFC3339),
	}

	if runtime.GOOS != "darwin" {
		record := evidence.Build(spec.PlatformMacOS, spec.BackendMacOSSandbox, spec.QualificationUnqualified, nil)
		record.Reject(spec.PhasePlatformQual, spec.DiagProfileUnsupported)
		report.Mechanisms = supportedMechanisms()
		report.EvidenceRecord = record
		report.ExitCode = ExitRejected
		report.DurationMS = time.Since(started).Milliseconds()
		return Result{Report: report, Evidence: record, ExitCode: ExitRejected}, nil
	}

	if ok, why := seatbelt.Available(); !ok {
		record := evidence.Build(spec.PlatformMacOS, spec.BackendMacOSSandbox, spec.QualificationUnqualified, nil)
		record.Reject(spec.PhasePlatformQual, spec.DiagProfileUnsupported)
		report.Mechanisms = append(supportedMechanisms(), Mechanism{
			Name:   seatbelt.ExecPath,
			Status: "unavailable",
			Note:   why,
		})
		report.EvidenceRecord = record
		report.ExitCode = ExitRejected
		report.DurationMS = time.Since(started).Milliseconds()
		return Result{Report: report, Evidence: record, ExitCode: ExitRejected}, nil
	}

	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	env, err := NewEnvironment(opts.WorkDir, opts.SelfPath)
	if err != nil {
		return Result{ExitCode: ExitHarnessError}, err
	}
	defer env.Close()

	run := runContext{ctx: ctx, env: env}

	// Domain creation is checked first: if a probe domain cannot be established
	// at all, every later observation would be unattributable.
	domainCreated := run.canCreateProbeDomain()

	var results []ClassResult
	if domainCreated {
		// The domain and wall-clock probes run first because the aggregate-bound
		// class reduces its supervisor-accounting conclusion from what they
		// measured, rather than from a platform table.
		domain := run.probeDomain()
		wall := run.probeWallClock()
		results = append(results,
			run.probeDomainMembership(domain),
			run.probeDomainTermination(domain),
			run.probeNetworkDenial(),
			run.probeEndpointRevocation(),
			run.probeReadOnlyView(spec.ClassReadOnlySource, "source", env.SourceDir),
			run.probeReadOnlyView(spec.ClassReadOnlyToolchain, "toolchain", env.GorootDir),
			run.probeWriteConfinement(),
			run.probeViewRestriction(),
			run.probeExecAllowlist(),
			run.probeAggregateBounds(domain, wall),
		)
	} else {
		for _, class := range spec.Classes() {
			if class == spec.ClassActiveProbe {
				continue
			}
			results = append(results, ClassResult{
				Class:   class,
				Verdict: VerdictInconclusive,
				Reasons: []string{"no probe domain could be established on this host"},
				Checks: []Check{failedCheck(class, kindPositive,
					"the class is observed in a probe domain", "probe domain could not be established")},
			})
		}
	}
	results = append(results, run.probeActive(results, domainCreated))
	results = applyForcedUnavailable(results, opts.ForceUnavailable)

	report.Classes = results
	report.Mechanisms = annotateMechanisms(supportedMechanisms(), results)

	record := evidence.Build(spec.PlatformMacOS, spec.BackendMacOSSandbox, spec.QualificationUnqualified, Observations(results))

	// The record must survive its own validation before it is emitted. A record
	// that does not is a corrupt result, not a rejection about the host.
	if diag, reason := record.Validate(); diag != "" {
		report.EvidenceRecord = record
		report.ExitCode = ExitHarnessError
		report.DurationMS = time.Since(started).Milliseconds()
		return Result{Report: report, Evidence: record, ExitCode: ExitHarnessError},
			fmt.Errorf("%s: %s", diag, reason)
	}

	exitCode := ExitEstablished
	if record.Outcome != spec.OutcomeEstablished {
		exitCode = ExitRejected
	}
	report.EvidenceRecord = record
	report.ExitCode = exitCode
	report.DurationMS = time.Since(started).Milliseconds()
	return Result{Report: report, Evidence: record, ExitCode: exitCode}, nil
}

// applyForcedUnavailable overrides measured verdicts. Every override is written
// into the class result, so a forced value is visible in the artifact.
func applyForcedUnavailable(results []ClassResult, forced []string) []ClassResult {
	if len(forced) == 0 {
		return results
	}
	set := map[string]bool{}
	for _, class := range forced {
		set[class] = true
	}
	for i := range results {
		if !set[results[i].Class] {
			continue
		}
		results[i].Verdict = VerdictUnavailable
		results[i].Applied = false
		results[i].Reasons = append(results[i].Reasons,
			"forced unavailable by --force-unavailable; this value was injected, not measured")
		results[i].Checks = append(results[i].Checks, Check{
			Name:        "forced-unavailable-injection",
			Kind:        kindAdversarial,
			Expectation: "with this class forced unavailable, the operation rejects before domain entry",
			Observed:    "injected-unavailable",
			Pass:        false,
			Detail:      "fail-closed control; not a host observation",
		})
	}
	return results
}

// canCreateProbeDomain runs the smallest possible in-domain agent to confirm a
// probe domain can be established and can report back.
func (r runContext) canCreateProbeDomain() bool {
	report, _, err := r.runAgent("domain-smoke", r.baseProfile(), inside.OpHello, nil, nil)
	return err == nil && report.Values["hello"] == "ok"
}

// FailClosedSweep forces each capability class unavailable in turn and checks
// that the run rejects before domain entry with the mapped diagnostic. It is
// the executable form of the section 11 preflight evidence.
func FailClosedSweep(ctx context.Context, opts Options) ([]FailClosed, error) {
	var out []FailClosed
	for _, class := range spec.Classes() {
		sweepOpts := opts
		sweepOpts.ForceUnavailable = []string{class}
		sweepOpts.WorkDir = opts.WorkDir
		result, err := Run(ctx, sweepOpts)
		if err != nil {
			return out, fmt.Errorf("fail-closed sweep for %s: %w", class, err)
		}
		entry := FailClosed{
			ForcedClass: class,
			Outcome:     result.Evidence.Outcome,
			ExitCode:    result.ExitCode,
		}
		if result.Evidence.RejectedBefore != nil {
			entry.RejectedBefore = *result.Evidence.RejectedBefore
		}
		if result.Evidence.Diagnostic != nil {
			entry.Diagnostic = *result.Evidence.Diagnostic
		}
		entry.Pass = entry.Outcome == spec.OutcomeRejected &&
			entry.RejectedBefore == spec.PhaseCapabilityProbe &&
			entry.Diagnostic == spec.DiagCapabilityUnavailable &&
			entry.ExitCode == ExitRejected &&
			classIsUnapplied(result.Evidence, class)
		out = append(out, entry)
	}
	return out, nil
}

func classIsUnapplied(record evidence.Record, class string) bool {
	for _, capability := range record.Capabilities {
		if capability.Name == class {
			return capability.Status == spec.StatusNotApplied &&
				capability.Availability == spec.AvailabilityUnavailable
		}
	}
	return false
}

// WriteArtifacts writes the closed evidence record and the detailed report.
func WriteArtifacts(result Result, evidencePath, reportPath string) error {
	if evidencePath != "" {
		if err := os.WriteFile(evidencePath, result.Evidence.JSON(), 0o644); err != nil {
			return fmt.Errorf("write evidence: %w", err)
		}
	}
	if reportPath != "" {
		if err := os.WriteFile(reportPath, result.Report.JSON(), 0o644); err != nil {
			return fmt.Errorf("write report: %w", err)
		}
	}
	return nil
}
