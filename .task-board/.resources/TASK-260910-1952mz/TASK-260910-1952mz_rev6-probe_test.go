package shell

import "testing"

// Throwaway rev6 probe (deleted after the run): replays the exact Windows CI
// observation from run 35201254365
// (TestShellHookTrustNativeRecordAuthorizesMSYSSpelling, changed-bytes
// A-warning phase) through assertTrustOutcome. The Old variant carries the
// rev5 expectation (Sourced, hardcoded marker 1) and must FAIL with the CI
// message; the New variant carries the rev6 expectation (SourcedMarker 2)
// and must PASS. CRLF framing is kept verbatim to exercise normalization.

const rev6ProbeCandidate = `C:\Users\runneradmin\AppData\Local\Temp\TestShellHookTrustNativeRecordAuthorizesMSYSSpellingbash3966133929\001\project\.agents\env.sh`

const rev6ProbeStdout = "sourced1=2\r\nsourced2=2\r\n"

var rev6ProbeStderr = "curator: shell_hook_env_changed: " + rev6ProbeCandidate +
	" changed since approval; run curator hook approve " + rev6ProbeCandidate +
	" to approve the new bytes\r\nPROBE-1\r\n"

func rev6ProbeCase(marker string) hookTrustCase {
	changed := DiagnosticEnvChanged
	return hookTrustCase{
		Name:                        "rev6-probe/changed/A-warning",
		Diagnostic:                  &changed,
		Sourced:                     true,
		SourcedMarker:               marker,
		WarningFirst:                true,
		WarningSecond:               false,
		WarningsTotal:               1,
		WarningNamesPath:            true,
		WarningNamesApprovalCommand: true,
	}
}

func TestRev6ProbeOldExpectation(t *testing.T) {
	assertTrustOutcome(t, rev6ProbeCase(""), rev6ProbeCandidate, rev6ProbeStdout, rev6ProbeStderr)
}

func TestRev6ProbeNewExpectation(t *testing.T) {
	assertTrustOutcome(t, rev6ProbeCase("2"), rev6ProbeCandidate, rev6ProbeStdout, rev6ProbeStderr)
}
