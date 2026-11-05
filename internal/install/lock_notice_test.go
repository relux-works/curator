package install

import (
	"testing"
	"time"
)

// A real waiting interval emits its scope once; an uncontended acquisition
// cancels notification, and joining prevents later writes after result rendering.
func TestLockWaitNotice(t *testing.T) {
	for _, wait := range []bool{false, true} {
		notified := make(chan string, 1)
		finish := lockWaitNotice(func(scope string) { notified <- scope }, "project")
		if wait {
			select {
			case scope := <-notified:
				if scope != "project" {
					t.Fatal(scope)
				}
			case <-time.After(2 * time.Second):
				t.Fatal("silent lock wait")
			}
		}
		finish()
		select {
		case <-notified:
			t.Fatal("unexpected notification")
		default:
		}
	}
}
