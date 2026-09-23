package goreleaserconfig

import (
	"os"
	"path/filepath"
	"testing"
)

// The wiring table. The gate only guards the committed file if a lane runs
// it, so the lint job must carry exactly one live step running GateRun. The
// commented-run and if:false rows are rev3 R3: the old substring pin
// accepted both. Quoting and trailing comments are YAML syntax, not value,
// so they still count as the same live invocation.
func TestCheckWiring(t *testing.T) {
	tests := []struct {
		name      string
		doc       string
		wantCount int
		wantSubs  []string
	}{
		{
			name: "minimal live step passes",
			doc: `jobs:
  lint:
    steps:
      - name: Verify GoReleaser rc channel values
        run: go test -count=1 ./tools/goreleaserconfig/
`,
			wantCount: 0,
		},
		{
			name: "a commented-out run line is not a live invocation",
			doc: `jobs:
  lint:
    steps:
      - name: Verify GoReleaser rc channel values
        # run: go test -count=1 ./tools/goreleaserconfig/
`,
			wantCount: 1,
			wantSubs:  []string{"no live lint step runs"},
		},
		{
			name: "an if:false step is not live",
			doc: `jobs:
  lint:
    steps:
      - name: Verify GoReleaser rc channel values
        run: go test -count=1 ./tools/goreleaserconfig/
        if: false
`,
			wantCount: 1,
			wantSubs:  []string{"no live lint step runs"},
		},
		{
			name: "any if condition needs human review, even if:true",
			doc: `jobs:
  lint:
    steps:
      - name: Verify GoReleaser rc channel values
        run: go test -count=1 ./tools/goreleaserconfig/
        if: true
`,
			wantCount: 1,
			wantSubs:  []string{"no live lint step runs"},
		},
		{
			name: "a deleted step fails",
			doc: `jobs:
  lint:
    steps:
      - name: No broad suppression
        run: bash .github/ci/no-broad-suppression.sh
`,
			wantCount: 1,
			wantSubs:  []string{"no live lint step runs"},
		},
		{
			name: "a changed gate path fails",
			doc: `jobs:
  lint:
    steps:
      - name: Verify GoReleaser rc channel values
        run: go test -count=1 ./tools/othergate/
`,
			wantCount: 1,
			wantSubs:  []string{"no live lint step runs"},
		},
		{
			name: "a quoted run value is the same invocation",
			doc: `jobs:
  lint:
    steps:
      - name: Verify GoReleaser rc channel values
        run: "go test -count=1 ./tools/goreleaserconfig/"
`,
			wantCount: 0,
		},
		{
			name: "a trailing comment on the run line still counts",
			doc: `jobs:
  lint:
    steps:
      - name: Verify GoReleaser rc channel values
        run: go test -count=1 ./tools/goreleaserconfig/ # the rc channel guard
`,
			wantCount: 0,
		},
		{
			name: "two live gate steps fail",
			doc: `jobs:
  lint:
    steps:
      - name: Verify GoReleaser rc channel values
        run: go test -count=1 ./tools/goreleaserconfig/
      - name: Verify GoReleaser rc channel values again
        run: go test -count=1 ./tools/goreleaserconfig/
`,
			wantCount: 1,
			wantSubs:  []string{"2 live lint steps"},
		},
		{
			name: "duplicate run keys fail instead of resolving by first/last wins",
			doc: `jobs:
  lint:
    steps:
      - name: Verify GoReleaser rc channel values
        run: go test -count=1 ./tools/goreleaserconfig/
        run: go test -count=1 ./tools/goreleaserconfig/
`,
			wantCount: 1,
			wantSubs:  []string{"invalid YAML", "already defined"},
		},
		{
			name: "a workflow without jobs fails",
			doc: `name: CI
on: push
`,
			wantCount: 1,
			wantSubs:  []string{`section "jobs" is absent`},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assertFindings(t, CheckWiring([]byte(tc.doc)), tc.wantCount, tc.wantSubs)
		})
	}
}

// TestCommittedWiring is the wiring production row: the repository's own
// ci.yml must carry the live gate step.
func TestCommittedWiring(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", ".github", "workflows", "ci.yml"))
	if err != nil {
		t.Fatalf("read ci.yml: %v", err)
	}
	assertFindings(t, CheckWiring(raw), 0, nil)
}
