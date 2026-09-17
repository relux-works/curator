package buildrepo

import (
	"context"
	"testing"
)

// TestReviewRev5GluedLineRefuses is the rev5 reviewer regression: a
// single original stderr line that matches no closed-table entry — an
// availability-shaped prefix glued to git's fixed framing trailer
// without a newline boundary — must fail closed (one fetch, lane
// refusal), never authorize the alternate endpoint.
func TestReviewRev5GluedLineRefuses(t *testing.T) {
	requirePOSIXTransport(t)
	fixture := makeGitFixture(t, "sha1", false)
	msg := "fatal: unable to access 'https://fixture.test/repository.git/': The requested URL returned error: 503fatal: the remote end hung up unexpectedly"
	tool, logPath := fakeTransportGitTool(t, map[string]transportBehavior{transportHTTPS: failTransport(msg), transportSSH: {succeed: true, fileRepo: fixture.bare}})
	tool = transportSSHBaseTool(t, tool)
	base := transportBase(t, tool, fixture)
	plan := TransportPlan{Identity: base.Source.Identity, Attempts: []TransportAttempt{{URL: transportHTTPS}, {URL: transportSSH}}, Fallback: TransportFallbackAvailabilityAuth}
	_, err := AcquireNetworkResolved(context.Background(), base, plan, nil, nil)
	n := len(transportFetchLines(t, logPath))
	if err == nil || n != 1 {
		t.Fatalf("malformed full line produced %d fetches, err=%v", n, err)
	}
}
