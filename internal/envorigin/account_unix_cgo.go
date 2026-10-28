//go:build (darwin || linux) && cgo

package envorigin

/*
#include <sys/types.h>
#include <pwd.h>
#include <unistd.h>
#include <stdlib.h>
#include <errno.h>
static int curator_origin_account(uid_t id, char *buf, size_t n, struct passwd *pw) {
 struct passwd *result = NULL;
 int code = getpwuid_r(id, pw, buf, n, &result);
 if (code != 0) return code;
 if (result == NULL) return ENOENT;
 return 0;
}
*/
import "C"
import (
	"fmt"
	"os"
	"unsafe"
)

// LookupAccount uses the numeric effective UID's native account/NSS record,
// never os/user.Current, HOME, USER, LOGNAME or os.UserHomeDir.
func LookupAccount() (Account, error) {
	realUID, effective := os.Getuid(), os.Geteuid()
	if realUID != effective {
		return Account{}, fmt.Errorf("realUID/effective UID mismatch")
	}
	for size := 1024; size <= 1<<20; size *= 2 {
		buf := C.malloc(C.size_t(size))
		if buf == nil {
			return Account{}, fmt.Errorf("account buffer allocation failed")
		}
		var pw C.struct_passwd
		code := C.curator_origin_account(C.uid_t(effective), (*C.char)(buf), C.size_t(size), &pw)
		if code == C.ERANGE {
			C.free(buf)
			continue
		}
		if code != 0 {
			C.free(buf)
			return Account{}, fmt.Errorf("native account lookup code %d", int(code))
		}
		home := C.GoString(pw.pw_dir)
		uid := uint64(pw.pw_uid)
		C.free(unsafe.Pointer(buf))
		return Account{UID: uid, RealUID: uint64(realUID), EffectiveUID: uint64(effective), Home: home}, nil
	}
	return Account{}, fmt.Errorf("native account record exceeds bound")
}
