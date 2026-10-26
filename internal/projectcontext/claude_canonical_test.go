package projectcontext

import (
	"context"
	"strings"
	"testing"
)

// Claim: absent and explicit empty manager tool grants both survive protected
// publication and re-resolution without changing the caller's inputs. A stored
// null grant array remains invalid under the exact closed record reader.
func TestClaudeEmptyControlsPublication(t *testing.T) {
	for _, explicit := range []bool{false, true} {
		name := "absent"
		if explicit {
			name = "empty"
		}
		t.Run(name, func(t *testing.T) {
			f, req, _, owner := claudeFixture(t, false, false)
			if explicit {
				req.Claude.AllowTools = []string{}
				req.Claude.DenyTools = []string{}
			}
			p, err := f.store.PreviewComposition(context.Background(), req)
			if err != nil {
				t.Fatal(err)
			}
			a := Authority{Store: f.store, CWD: f.root}
			c, err := a.PublishComposition(context.Background(), req, p.SHA256, owner)
			if err != nil {
				t.Fatalf("empty controls publication refused: %v", err)
			}
			pub := &claudePublisherFixture{t: t, home: t.TempDir(), expected: c}
			if err = a.PublishClaude(context.Background(), c.Ref, c.SHA256, owner, "reserved", pub); err != nil {
				t.Fatalf("published empty controls cannot resolve: %v", err)
			}
			if !pub.called {
				t.Fatal("valid empty controls never reached publisher")
			}
			if !explicit && (req.Claude.AllowTools != nil || req.Claude.DenyTools != nil) {
				t.Fatal("caller controls were mutated")
			}
		})
	}
	var controls ClaudeControls
	if err := decode([]byte(`{"Version":1,"AllowTools":null,"DenyTools":[]}`), &controls); err == nil || !strings.Contains(err.Error(), "expected array") {
		t.Fatalf("stored null tool grants were accepted: %v", err)
	}
}
