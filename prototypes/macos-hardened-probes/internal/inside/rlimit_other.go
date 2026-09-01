//go:build !darwin

package inside

// rlimitNPROC is RLIMIT_NPROC on Linux (<bits/resource.h>). This prototype is
// macOS-primary and its results are only meaningful there; the constant exists
// so the package still builds and its non-platform logic stays testable
// elsewhere.
const rlimitNPROC = 6

const nprocResourceName = "RLIMIT_NPROC"
