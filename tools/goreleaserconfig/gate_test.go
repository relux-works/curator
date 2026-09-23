package goreleaserconfig

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The value table. Every fail row asserts the exact finding count plus the
// wording contract (field, entry, observed value); pass rows assert silence.
// Fixture provenance: block_first.yml, committed_block_bypass.yml and
// duplicate_good_last.yml are the rev3 reviewer's R1/R2 fixtures byte for
// byte (TASK-260908-2kqa77_review-evidence-rev3.tar.gz); rows C and E are
// the named negatives from the TASK-260908-1jv1h3 cycle-3 verdict.
func TestCheck(t *testing.T) {
	tests := []struct {
		name      string
		doc       string
		file      string // testdata file instead of doc, when set
		wantCount int
		wantSubs  []string
	}{
		{
			name: "minimal good document passes",
			doc: `homebrew_casks:
  - name: curator
    skip_upload: "auto"
scoops:
  - name: curator
    skip_upload: "auto"
release:
  prerelease: "auto"
`,
			wantCount: 0,
		},
		{
			name: "row C: both skip_upload keys absent",
			doc: `homebrew_casks:
  - name: curator
scoops:
  - name: curator
release:
  prerelease: "auto"
`,
			wantCount: 2,
			wantSubs: []string{
				`homebrew_casks[0].skip_upload is absent`,
				`scoops[0].skip_upload is absent`,
			},
		},
		{
			name: "row C narrowed: one entry losing its key is already a publish path",
			doc: `homebrew_casks:
  - name: curator
scoops:
  - name: curator
    skip_upload: "auto"
release:
  prerelease: "auto"
`,
			wantCount: 1,
			wantSubs:  []string{`homebrew_casks[0].skip_upload is absent, want "auto"`},
		},
		{
			name: "row E: Auto in the cask entry fails case-sensitively",
			doc: `homebrew_casks:
  - name: curator
    skip_upload: "Auto"
scoops:
  - name: curator
    skip_upload: "auto"
release:
  prerelease: "auto"
`,
			wantCount: 1,
			wantSubs:  []string{`homebrew_casks[0].skip_upload = "Auto"`, `want "auto"`},
		},
		{
			name: "Auto in the scoop entry fails",
			doc: `homebrew_casks:
  - name: curator
    skip_upload: "auto"
scoops:
  - name: curator
    skip_upload: "Auto"
release:
  prerelease: "auto"
`,
			wantCount: 1,
			wantSubs:  []string{`scoops[0].skip_upload = "Auto"`},
		},
		{
			name: "sometimes fails",
			doc: `homebrew_casks:
  - name: curator
    skip_upload: sometimes
scoops:
  - name: curator
    skip_upload: "auto"
release:
  prerelease: "auto"
`,
			wantCount: 1,
			wantSubs:  []string{`homebrew_casks[0].skip_upload = "sometimes"`},
		},
		{
			name: "boolean true fails",
			doc: `homebrew_casks:
  - name: curator
    skip_upload: true
scoops:
  - name: curator
    skip_upload: "auto"
release:
  prerelease: "auto"
`,
			wantCount: 1,
			wantSubs:  []string{`homebrew_casks[0].skip_upload = "true"`},
		},
		{
			name: `quoted "true" fails too`,
			doc: `homebrew_casks:
  - name: curator
    skip_upload: "true"
scoops:
  - name: curator
    skip_upload: "auto"
release:
  prerelease: "auto"
`,
			wantCount: 1,
			wantSubs:  []string{`homebrew_casks[0].skip_upload = "true"`},
		},
		{
			name: `prerelease "ato" fails`,
			doc: `homebrew_casks:
  - name: curator
    skip_upload: "auto"
scoops:
  - name: curator
    skip_upload: "auto"
release:
  prerelease: "ato"
`,
			wantCount: 1,
			wantSubs:  []string{`release.prerelease = "ato"`},
		},
		{
			name: `prerelease "Auto" fails`,
			doc: `homebrew_casks:
  - name: curator
    skip_upload: "auto"
scoops:
  - name: curator
    skip_upload: "auto"
release:
  prerelease: "Auto"
`,
			wantCount: 1,
			wantSubs:  []string{`release.prerelease = "Auto"`},
		},
		{
			name: "absent prerelease fails",
			doc: `homebrew_casks:
  - name: curator
    skip_upload: "auto"
scoops:
  - name: curator
    skip_upload: "auto"
release:
  draft: false
`,
			wantCount: 1,
			wantSubs:  []string{`release.prerelease is absent, want "auto"`},
		},
		{
			name:      "R1a minimal: first-position block scalar body is not a key",
			file:      "block_first.yml",
			wantCount: 1,
			wantSubs:  []string{`homebrew_casks[0].skip_upload is absent`},
		},
		{
			name:      "R1a full config: committed-shaped block scalar bypass fails",
			file:      "committed_block_bypass.yml",
			wantCount: 1,
			wantSubs:  []string{`homebrew_casks[0].skip_upload is absent`},
		},
		{
			name: "R1b: first-position nested repository map is not a key",
			doc: `homebrew_casks:
  - repository:
      skip_upload: auto
    name: curator
scoops:
  - name: curator
    skip_upload: auto
release:
  prerelease: auto
`,
			wantCount: 1,
			wantSubs:  []string{`homebrew_casks[0].skip_upload is absent`},
		},
		{
			name:      "R2: duplicate keys fail even when the last one reads auto",
			file:      "duplicate_good_last.yml",
			wantCount: 1,
			wantSubs:  []string{"invalid YAML", "already defined"},
		},
		{
			name: "a bad second entry fails naming its index",
			doc: `homebrew_casks:
  - name: curator
    skip_upload: "auto"
scoops:
  - name: curator
    skip_upload: "auto"
  - name: extra
    skip_upload: "sometimes"
release:
  prerelease: "auto"
`,
			wantCount: 1,
			wantSubs:  []string{`scoops[1].skip_upload = "sometimes"`},
		},
		{
			name: "a good second entry passes",
			doc: `homebrew_casks:
  - name: curator
    skip_upload: "auto"
scoops:
  - name: curator
    skip_upload: "auto"
  - name: extra
    skip_upload: "auto"
release:
  prerelease: "auto"
`,
			wantCount: 0,
		},
		{
			name: "unquoted auto passes (same parsed string)",
			doc: `homebrew_casks:
  - name: curator
    skip_upload: auto
scoops:
  - name: curator
    skip_upload: auto
release:
  prerelease: auto
`,
			wantCount: 0,
		},
		{
			name: "single-quoted auto passes",
			doc: `homebrew_casks:
  - name: curator
    skip_upload: 'auto'
scoops:
  - name: curator
    skip_upload: 'auto'
release:
  prerelease: 'auto'
`,
			wantCount: 0,
		},
		{
			name: "trailing comments pass",
			doc: `homebrew_casks:
  - name: curator
    skip_upload: "auto" # keep the rc out of the tap
scoops:
  - name: curator
    skip_upload: "auto" # same
release:
  prerelease: "auto" # an rc says what it is
`,
			wantCount: 0,
		},
		{
			name: "a nested skip_upload after an ordinary key does not satisfy the entry",
			doc: `homebrew_casks:
  - name: curator
    repository:
      skip_upload: "auto"
scoops:
  - name: curator
    skip_upload: "auto"
release:
  prerelease: "auto"
`,
			wantCount: 1,
			wantSubs:  []string{`homebrew_casks[0].skip_upload is absent`},
		},
		{
			name: "a removed channel stanza fails",
			doc: `scoops:
  - name: curator
    skip_upload: "auto"
release:
  prerelease: "auto"
`,
			wantCount: 1,
			wantSubs:  []string{`section "homebrew_casks" is absent`},
		},
		{
			name: "a null channel stanza has no entries",
			doc: `homebrew_casks:
scoops:
  - name: curator
    skip_upload: "auto"
release:
  prerelease: "auto"
`,
			wantCount: 1,
			wantSubs:  []string{`section "homebrew_casks" has no entries`},
		},
		{
			name: "an empty channel list has no entries",
			doc: `homebrew_casks: []
scoops:
  - name: curator
    skip_upload: "auto"
release:
  prerelease: "auto"
`,
			wantCount: 1,
			wantSubs:  []string{`section "homebrew_casks" has no entries`},
		},
		{
			name:      "tab indentation fails closed",
			doc:       "homebrew_casks:\n\t- name: curator\n\t  skip_upload: auto\nscoops:\n  - name: curator\n    skip_upload: auto\nrelease:\n  prerelease: auto\n",
			wantCount: 1,
			wantSubs:  []string{"invalid YAML"},
		},
		{
			name: `" auto " fails strictly`,
			doc: `homebrew_casks:
  - name: curator
    skip_upload: " auto "
scoops:
  - name: curator
    skip_upload: "auto"
release:
  prerelease: "auto"
`,
			wantCount: 1,
			wantSubs:  []string{`homebrew_casks[0].skip_upload = " auto "`},
		},
		{
			name: "a missing release section fails",
			doc: `homebrew_casks:
  - name: curator
    skip_upload: "auto"
scoops:
  - name: curator
    skip_upload: "auto"
`,
			wantCount: 1,
			wantSubs:  []string{`section "release" is absent`},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			doc := tc.doc
			if tc.file != "" {
				raw, err := os.ReadFile(filepath.Join("testdata", tc.file))
				if err != nil {
					t.Fatalf("read fixture %s: %v", tc.file, err)
				}
				doc = string(raw)
			}
			assertFindings(t, Check([]byte(doc)), tc.wantCount, tc.wantSubs)
		})
	}
}

// TestCommittedConfig is the gate's production row: the repository's own
// .goreleaser.yml must parse with every channel value exactly "auto".
func TestCommittedConfig(t *testing.T) {
	got := CheckFile(filepath.Join("..", "..", ".goreleaser.yml"))
	assertFindings(t, got, 0, nil)
}

// TestCheckFileMissing proves the gate fails closed on an unreadable file
// instead of reporting a healthy empty finding set.
func TestCheckFileMissing(t *testing.T) {
	got := CheckFile(filepath.Join("testdata", "grl-absent.yml"))
	assertFindings(t, got, 1, []string{"cannot read"})
}

func assertFindings(t *testing.T, got []string, wantCount int, wantSubs []string) {
	t.Helper()
	if len(got) != wantCount {
		t.Fatalf("got %d finding(s), want %d:\n%s", len(got), wantCount, strings.Join(got, "\n"))
	}
	joined := strings.Join(got, "\n")
	for _, sub := range wantSubs {
		if !strings.Contains(joined, sub) {
			t.Errorf("findings do not mention %q:\n%s", sub, joined)
		}
	}
}
