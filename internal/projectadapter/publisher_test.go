package projectadapter

import (
	"context"
	"errors"
	"reflect"
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

// Claim: prepared publication delegates the complete immutable final file map
// to the lease owner, propagates that owner's refusal and never mutates its
// baseline. Missing publisher/id and cancellation make zero publisher calls.
// This proves the interface seam, not actual lease transaction or liveness.
func TestCodexPreparedPublisher(t *testing.T) {
	plan := render(t, request(t))
	publisher := &recordingPublisher{}
	if err := plan.PublishPrepared(context.Background(), "lease-fixture", publisher); err != nil {
		t.Fatal(err)
	}
	if publisher.calls != 1 || publisher.lease != "lease-fixture" || publisher.dryRun || !reflect.DeepEqual(publisher.files, plan.Files()) {
		t.Fatal("lease publication contribution differs")
	}
	publisher.files["AGENTS.md"][0] = 'x'
	if plan.Files()["AGENTS.md"][0] == 'x' {
		t.Fatal("publisher changed immutable baseline")
	}
	expected := errors.New("fixture owner rejects receipt")
	publisher.err = expected
	if err := plan.PublishPrepared(context.Background(), "lease-fixture", publisher); !errors.Is(err, expected) {
		t.Fatal("lease owner refusal was hidden")
	}
	publisher.err = nil
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
		if err := plan.PublishPrepared(ctx, id, target); err == nil {
			t.Fatal("invalid publication accepted")
		}
		if publisher.calls != 0 {
			t.Fatal("refused publication reached lease owner")
		}
	}
}
