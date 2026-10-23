package projectadapter

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type recordingPublisher struct {
	calls  int
	files  map[string][]byte
	lease  string
	dryRun bool
	err    error
}

func (p *recordingPublisher) PublishPrepared(_ context.Context, lease string, files map[string][]byte, dryRun bool) error {
	p.calls++
	p.files = files
	p.lease = lease
	p.dryRun = dryRun
	return p.err
}

// Claim: a valid preview cannot reach the lease publisher while native-project
// permission policy is refused. Invalid publisher/id and cancellation retain their exact
// earlier errors and make zero calls. Materialize tests provide the nearby
// successful preview-publication control; no supported native launch is claimed.
func TestCodexPreparedPublisher(t *testing.T) {
	plan := render(t, request(t))
	publisher := &recordingPublisher{}
	var refusal *Unsupported
	if err := plan.PublishPrepared(context.Background(), "lease-fixture", publisher); !errors.As(err, &refusal) || refusal.Surface != Permissions {
		t.Fatalf("incompatible publication accepted: %v", err)
	}
	if publisher.calls != 0 {
		t.Fatal("refused permissions reached lease owner")
	}

	for _, kind := range []string{"publisher", "identifier", "cancelled"} {
		publisher.calls = 0
		ctx := context.Background()
		var target PreparedPublisher = publisher
		id := "lease-fixture"
		switch kind {
		case "publisher":
			target = nil
		case "identifier":
			id = "../escape"
		case "cancelled":
			var cancel context.CancelFunc
			ctx, cancel = context.WithCancel(ctx)
			cancel()
		}
		err := plan.PublishPrepared(ctx, id, target)
		if kind == "cancelled" {
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("invalid publication accepted: want cancellation, got %v", err)
			}
		} else if err == nil || !strings.Contains(err.Error(), "project_lease_mismatch") {
			t.Fatalf("invalid publication accepted: want lease mismatch, got %v", err)
		}
		if publisher.calls != 0 {
			t.Fatal("refused publication reached lease owner")
		}
	}
}
