// Command hardened-probe measures whether a macOS host can actually provide
// the six guarantees of the Curator Hardened Execution Profile 1.0.
//
// It is a prototype. It establishes probe domains, never a build domain; it
// runs no Go compiler and reads no package byte. A run that reports every
// capability available would still not be a claim that this host enforces
// hardened builds — that claim belongs to a qualified implementation with
// independent review, per hardened-execution.md section 11.
//
// Exit status:
//
//	0  every capability applied and every guarantee established
//	1  rejected: at least one capability could not be established (fail-closed)
//	2  the harness itself could not produce a trustworthy record
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/relux-works/curator/prototypes/macos-hardened-probes/internal/inside"
	"github.com/relux-works/curator/prototypes/macos-hardened-probes/internal/probe"
	"github.com/relux-works/curator/prototypes/macos-hardened-probes/internal/spec"
)

func main() {
	// The in-domain agent is the same binary re-invoked, so the exact
	// executable allowlist needs exactly one entry.
	if len(os.Args) > 1 && os.Args[1] == "__inside" {
		os.Exit(inside.Main(os.Args[2:]))
	}
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	fs := flag.NewFlagSet("hardened-probe", flag.ContinueOnError)
	var (
		workDir      = fs.String("work-dir", "", "directory for the probe environment (default: a new temporary directory)")
		evidencePath = fs.String("evidence", "", "write the closed hardened-capability-evidence-v1 record here")
		reportPath   = fs.String("report", "", "write the detailed probe report here")
		forced       = fs.String("force-unavailable", "", "comma-separated capability classes to force unavailable (fail-closed control)")
		failClosed   = fs.Bool("fail-closed-sweep", false, "force each capability class unavailable in turn and check the rejection")
		expect       = fs.String("expect", "", "assert the outcome: established or rejected; exit 0 when it matches")
		timeout      = fs.Duration("timeout", 5*time.Minute, "overall run timeout")
		quiet        = fs.Bool("quiet", false, "do not print the evidence record to stdout")
		listClasses  = fs.Bool("list-classes", false, "print the capability-class inventory and exit")
	)
	fs.Usage = func() {
		_, _ = fmt.Fprintf(fs.Output(), "usage: hardened-probe [flags]\n\nProbes macOS for the capability classes of %s.\n\nFlags:\n", spec.SpecRevision)
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return probe.ExitHarnessError
	}

	if *listClasses {
		for _, class := range spec.Classes() {
			fmt.Println(class)
		}
		return probe.ExitEstablished
	}

	self, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "hardened-probe: cannot locate own executable: %v\n", err)
		return probe.ExitHarnessError
	}
	self, err = filepath.EvalSymlinks(self)
	if err != nil {
		fmt.Fprintf(os.Stderr, "hardened-probe: cannot resolve own executable: %v\n", err)
		return probe.ExitHarnessError
	}

	dir := *workDir
	if dir == "" {
		dir, err = os.MkdirTemp("", "hardened-probe-")
		if err != nil {
			fmt.Fprintf(os.Stderr, "hardened-probe: work directory: %v\n", err)
			return probe.ExitHarnessError
		}
		// Best effort: a leftover probe directory is untidy, not a wrong result.
		defer func() { _ = os.RemoveAll(dir) }()
	} else if err := os.MkdirAll(dir, 0o700); err != nil {
		fmt.Fprintf(os.Stderr, "hardened-probe: work directory: %v\n", err)
		return probe.ExitHarnessError
	}

	opts := probe.Options{
		WorkDir:          dir,
		SelfPath:         self,
		ForceUnavailable: splitClasses(*forced),
		Timeout:          *timeout,
	}
	for _, class := range opts.ForceUnavailable {
		if !spec.IsClass(class) {
			fmt.Fprintf(os.Stderr, "hardened-probe: %q is not a capability class of %s\n", class, spec.InventoryVersion)
			return probe.ExitHarnessError
		}
	}

	ctx := context.Background()
	result, err := probe.Run(ctx, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "hardened-probe: %v\n", err)
		if result.ExitCode == 0 {
			return probe.ExitHarnessError
		}
	}

	if *failClosed {
		sweepDir := filepath.Join(dir, "fail-closed")
		if err := os.MkdirAll(sweepDir, 0o700); err != nil {
			fmt.Fprintf(os.Stderr, "hardened-probe: fail-closed sweep: %v\n", err)
			return probe.ExitHarnessError
		}
		sweepOpts := opts
		sweepOpts.WorkDir = sweepDir
		sweep, sweepErr := probe.FailClosedSweep(ctx, sweepOpts)
		result.Report.FailClosed = sweep
		if sweepErr != nil {
			fmt.Fprintf(os.Stderr, "hardened-probe: fail-closed sweep: %v\n", sweepErr)
			return probe.ExitHarnessError
		}
		for _, entry := range sweep {
			if !entry.Pass {
				fmt.Fprintf(os.Stderr,
					"hardened-probe: forcing %s unavailable did not reject before domain entry (outcome=%s before=%s diagnostic=%s exit=%d)\n",
					entry.ForcedClass, entry.Outcome, entry.RejectedBefore, entry.Diagnostic, entry.ExitCode)
				return probe.ExitHarnessError
			}
		}
	}

	if err := probe.WriteArtifacts(result, *evidencePath, *reportPath); err != nil {
		fmt.Fprintf(os.Stderr, "hardened-probe: %v\n", err)
		return probe.ExitHarnessError
	}

	if !*quiet {
		// The evidence record goes to stdout and is the machine-readable
		// result; a failure to deliver it means the caller has no record, which
		// is a harness fault rather than a statement about the host.
		if _, err := os.Stdout.Write(result.Evidence.JSON()); err != nil {
			fmt.Fprintf(os.Stderr, "hardened-probe: write evidence record: %v\n", err)
			return probe.ExitHarnessError
		}
		fmt.Fprint(os.Stderr, Summary(result))
	}

	if *expect != "" {
		switch *expect {
		case spec.OutcomeEstablished, spec.OutcomeRejected:
		default:
			fmt.Fprintf(os.Stderr, "hardened-probe: --expect must be %q or %q\n", spec.OutcomeEstablished, spec.OutcomeRejected)
			return probe.ExitHarnessError
		}
		if result.Evidence.Outcome != *expect {
			fmt.Fprintf(os.Stderr, "hardened-probe: expected outcome %q, observed %q\n", *expect, result.Evidence.Outcome)
			return probe.ExitHarnessError
		}
		// The assertion succeeded. The host result is still reported in the
		// record and the summary; the exit status now answers the assertion.
		return probe.ExitEstablished
	}

	return result.ExitCode
}

// Summary renders the operator-facing view of a run. It is built as a string
// rather than printed directly so a test can read exactly what an operator
// would see, and so the harness never half-writes a summary it cannot finish.
func Summary(result probe.Result) string {
	var out strings.Builder
	fmt.Fprintf(&out, "\nplatform            %s / %s (%s)\n",
		result.Report.Host.ProductName, result.Report.Host.ProductVersion, result.Report.Host.Arch)
	fmt.Fprintf(&out, "backend             %s (%s)\n", result.Report.Backend, spec.QualificationUnqualified)
	fmt.Fprintf(&out, "outcome             %s\n", result.Evidence.Outcome)
	if result.Evidence.Diagnostic != nil {
		fmt.Fprintf(&out, "rejected before     %s\ndiagnostic          %s\n",
			derefOr(result.Evidence.RejectedBefore), derefOr(result.Evidence.Diagnostic))
	}
	out.WriteString("\ncapability classes\n")
	byClass := map[string]probe.ClassResult{}
	for _, class := range result.Report.Classes {
		byClass[class.Class] = class
	}
	for _, capability := range result.Evidence.Capabilities {
		detail := ""
		if class, ok := byClass[capability.Name]; ok && len(class.Reasons) > 0 {
			detail = "  " + class.Reasons[0]
		}
		fmt.Fprintf(&out, "  %-34s %-12s %s%s\n", capability.Name, capability.Availability, capability.Status, detail)
	}
	out.WriteString("\nguarantees\n")
	for _, guarantee := range result.Evidence.Guarantees {
		state := "not established"
		if guarantee.Established {
			state = "established"
		}
		fmt.Fprintf(&out, "  %-44s %s\n", guarantee.Name, state)
	}
	if len(result.Report.FailClosed) > 0 {
		out.WriteString("\nfail-closed sweep\n")
		for _, entry := range result.Report.FailClosed {
			verdict := "FAIL"
			if entry.Pass {
				verdict = "ok"
			}
			fmt.Fprintf(&out, "  %-34s %-4s rejected before %s with %s (exit %d)\n",
				entry.ForcedClass, verdict, entry.RejectedBefore, entry.Diagnostic, entry.ExitCode)
		}
	}
	fmt.Fprintf(&out, "\nThis is a capability observation, not an enforcement claim. %s remains unqualified.\n",
		spec.PlatformMacOS)
	return out.String()
}

func derefOr(s *string) string {
	if s == nil {
		return "null"
	}
	return *s
}

func splitClasses(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var out []string
	for _, part := range strings.Split(raw, ",") {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
