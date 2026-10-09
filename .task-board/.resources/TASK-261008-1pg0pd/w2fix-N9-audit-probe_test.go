package audit

import (
	"github.com/relux-works/curator/internal/capabilities"
	"github.com/relux-works/curator/internal/hashing"
	"testing"
)

func TestWave2AuditDecisionTable(t *testing.T) {
	cases := []struct {
		name, mode, fail    string
		schema              int
		pin, revoked, block bool
	}{
		{"strict-high", "strict", "high", 3, false, false, true},
		{"strict-off", "strict", "off", 3, false, false, false},
		{"advisory-high", "advisory", "high", 3, false, false, false},
		{"pre-capability-pin-required", "strict", "high", 2, false, false, true},
		{"pin-over-finding", "strict", "high", 3, true, false, false},
		{"pre-capability-pinned", "strict", "high", 2, true, false, false},
		{"revocation-over-pin", "strict", "off", 3, true, true, true},
		{"revocation-advisory", "advisory", "off", 3, false, true, true},
		{"source-glob", "advisory", "off", 3, true, true, true},
		{"git-glob", "strict", "off", 3, true, true, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cfg := newCfg(t, c.mode, c.fail)
			s := subjectWith(t, "curl https://fixture.invalid/path\n", capabilities.ImplicitNone(), c.schema)
			hash, err := hashing.ContentSHA256(s.Snapshot, nil)
			if err != nil {
				t.Fatal(err)
			}
			if c.pin {
				if _, err := Pin(cfg.Home(), hash, "synthetic approval", "fixture"); err != nil {
					t.Fatal(err)
				}
			}
			if c.revoked {
				cfg.Audit.Revocations = []string{hash}
				if c.name == "source-glob" {
					cfg.Audit.Revocations = []string{"source:skill-*"}
				}
				if c.name == "git-glob" {
					cfg.Audit.Revocations = []string{"source:git@*.example.com:skills/*.git"}
				}
			}
			for _, entry := range []struct {
				name     string
				readOnly bool
			}{{"Gate", false}, {"GateReadOnly", true}} {
				var errs []string
				if entry.readOnly {
					_, errs = GateReadOnly(cfg, []Subject{s})
				} else {
					_, errs = Gate(cfg, []Subject{s})
				}
				t.Logf("%s blocked=%v", entry.name, len(errs) > 0)
				if (len(errs) > 0) != c.block {
					t.Fatalf("%s block=%v expected=%v", entry.name, errs, c.block)
				}
			}
		})
	}
}
