//go:build darwin

package inside

// rlimitNPROC is RLIMIT_NPROC of Darwin's <sys/resource.h>. Go's syscall
// package exports RLIMIT_AS, RLIMIT_CPU, RLIMIT_DATA and RLIMIT_NOFILE on
// darwin but not this one, so the value is transcribed here with the header it
// comes from named. It is not discovered at runtime: a wrong resource number
// would silently bound something other than the process count, and a probe that
// cannot say which resource it measured is not evidence.
const rlimitNPROC = 7

// nprocResourceName is what the report calls the resource, so a reader does not
// have to know the number.
const nprocResourceName = "RLIMIT_NPROC"
