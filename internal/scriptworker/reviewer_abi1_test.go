package scriptworker

import "testing"

// Linux v5.13 include/uapi/linux/landlock.h defines all nine flags
// below in ABI 1. Only REFER (bit 13) first appeared in ABI 2.
func TestReviewerABI1MutationRights(t *testing.T) {
	got := landlockHandledForABI(1, false, true)
	rights := map[string]uint64{
		"REMOVE_DIR": 1 << 4, "REMOVE_FILE": 1 << 5,
		"MAKE_CHAR": 1 << 6, "MAKE_DIR": 1 << 7,
		"MAKE_REG": 1 << 8, "MAKE_SOCK": 1 << 9,
		"MAKE_FIFO": 1 << 10, "MAKE_BLOCK": 1 << 11,
		"MAKE_SYM": 1 << 12,
	}
	for name, right := range rights {
		t.Run(name, func(t *testing.T) {
			if got&right == 0 {
				t.Fatalf("ABI 1 supports %s=%#x but production handled mask %#x leaves it unrestricted", name, right, got)
			}
		})
	}
	const want = uint64(0x1ff2)
	if got != want {
		t.Errorf("ABI 1 write mask = %#x; want %#x", got, want)
	}
	if got&(1<<13) != 0 {
		t.Error("ABI 1 must not request REFER")
	}
}
