package install

import "time"

// Join the timer callback before returning so an operator writer is never used
// concurrently with the caller's subsequent result rendering.
func lockWaitNotice(notify func(string), scope string) func() {
	if notify == nil {
		return func() {}
	}
	done := make(chan struct{})
	timer := time.AfterFunc(250*time.Millisecond, func() { defer close(done); notify(scope) })
	return func() {
		if !timer.Stop() {
			<-done
		}
	}
}
