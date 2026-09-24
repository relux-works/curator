//go:build !unix && !windows

package scriptworker

// The script inventory covers Linux, macOS, and Windows. The parent refuses
// before the worker starts on any other host, so these entry points exist
// only to keep the package buildable and must never confirm anything.

func confirmScriptTermination(_ *workerRequest) error {
	return diagnostic(CodeCapabilityEvidenceInvalid, "the enforced script policy is specified for Linux, macOS, and Windows only")
}

func confirmScriptAggregateLimits(_ *workerRequest, _ map[string]bool) error {
	return diagnostic(CodeCapabilityEvidenceInvalid, "the enforced script policy is specified for Linux, macOS, and Windows only")
}

func confirmScriptFileSize(_ *workerRequest) error {
	return diagnostic(CodeCapabilityEvidenceInvalid, "the enforced script policy is specified for Linux, macOS, and Windows only")
}

func confirmScriptHandles() error {
	return diagnostic(CodeCapabilityEvidenceInvalid, "the enforced script policy is specified for Linux, macOS, and Windows only")
}

// currentScriptFileSizeLimit reports no bound off the supported hosts.
func currentScriptFileSizeLimit() (uint64, bool) { return 0, false }
