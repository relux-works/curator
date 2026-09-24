package scriptworker
import "testing"
// Values independently sourced from Linux v6.8 include/uapi/linux/landlock.h.
func TestReviewerTruncateUAPIIdentity(t *testing.T) {
 if landlockAccessFSTruncate != 1<<14 { t.Errorf("TRUNCATE = %#x, kernel UAPI requires %#x (current value is MAKE_SYM)",landlockAccessFSTruncate,1<<14) }
 if landlockHandledForABI(4,false,true)&(1<<14)==0 { t.Error("real TRUNCATE is unhandled and therefore unrestricted") }
}
func TestReviewerWriteConfinementMask(t *testing.T) {
 mask := landlockHandledForABI(4, false, true)
 for name, right := range map[string]uint64{"remove-file":1<<5,"remove-dir":1<<4,"make-dir":1<<7,"make-reg":1<<8,"make-fifo":1<<10} {
  if mask & right == 0 { t.Errorf("%s is unrestricted: handled mask %#x omits %#x",name,mask,right) }
 }
}
func TestReviewerRegularFileTruncationGrant(t *testing.T) {
 // Test intended typing independently of the incorrect UAPI value.
 rights:=landlockRuleRights(false,landlockAccessFSWriteFile|landlockAccessFSTruncate,landlockAccessFSWriteFile|landlockAccessFSTruncate)
 if rights & landlockAccessFSTruncate == 0 { t.Errorf("regular-file write grant %#x strips its intended truncation bit",rights) }
}
