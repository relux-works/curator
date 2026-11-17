package state

import (
	"encoding/json"
	"testing"
)

// Claim: the bound tuple and the prepared-only legacy form round-trip in
// their closed member sets and never mix members across forms or variants.
func TestProcessClosedRoundTrip(t *testing.T) {
	good := []string{
		`{"pid":4242,"uid":1001,"boot_id":"6f1c0e2a-8b44-4d1e-9a51-2f0c3d7e9b10","start_time":{"kind":"linux-stat-starttime","ticks":"123456789","hz":100}}`,
		`{"pid":4242,"uid":1001,"boot_id":"6f1c0e2a-8b44-4d1e-9a51-2f0c3d7e9b10","start_time":{"kind":"darwin-pbi-start","sec":"1731715200","usec":123456}}`,
		`{"pid":4242,"started_at":"2026-10-27T00:00:01Z"}`,
	}
	for _, raw := range good {
		var p Process
		if err := json.Unmarshal([]byte(raw), &p); err != nil {
			t.Fatalf("valid control %s: %v", raw, err)
		}
		out, err := json.Marshal(p)
		if err != nil || string(out) != raw {
			t.Fatalf("round trip changed bytes: %s -> %s (%v)", raw, out, err)
		}
		if err := ValidateProcessSemantics(&p); err != nil {
			t.Fatalf("valid control semantics %s: %v", raw, err)
		}
	}
	bad := map[string]string{
		"mixed legacy and tuple": `{"pid":1,"uid":1,"boot_id":"6f1c0e2a-8b44-4d1e-9a51-2f0c3d7e9b10","start_time":{"kind":"linux-stat-starttime","ticks":"1","hz":100},"started_at":"2026-10-27T00:00:01Z"}`,
		"extra member":           `{"pid":1,"uid":1,"boot_id":"6f1c0e2a-8b44-4d1e-9a51-2f0c3d7e9b10","start_time":{"kind":"linux-stat-starttime","ticks":"1","hz":100},"extra":1}`,
		"missing uid":            `{"pid":1,"boot_id":"6f1c0e2a-8b44-4d1e-9a51-2f0c3d7e9b10","start_time":{"kind":"linux-stat-starttime","ticks":"1","hz":100}}`,
		"unknown kind":           `{"pid":1,"uid":1,"boot_id":"6f1c0e2a-8b44-4d1e-9a51-2f0c3d7e9b10","start_time":{"kind":"other","ticks":"1","hz":100}}`,
		"cross-variant members":  `{"pid":1,"uid":1,"boot_id":"6f1c0e2a-8b44-4d1e-9a51-2f0c3d7e9b10","start_time":{"kind":"linux-stat-starttime","sec":"1","usec":1}}`,
		"float pid":              `{"pid":1.5,"uid":1,"boot_id":"6f1c0e2a-8b44-4d1e-9a51-2f0c3d7e9b10","start_time":{"kind":"linux-stat-starttime","ticks":"1","hz":100}}`,
		"legacy extra member":    `{"pid":1,"started_at":"2026-10-27T00:00:01Z","uid":1}`,
	}
	for name, raw := range bad {
		t.Run(name, func(t *testing.T) {
			var p Process
			if err := json.Unmarshal([]byte(raw), &p); err == nil {
				t.Fatalf("accepted forbidden process member set: %s", raw)
			}
		})
	}
}

// Claim: numeric ceilings are enforced semantically at their exact boundary
// (2^64-1 ticks/sec, uid_t excluding -1, pid_t), one step past each refusing.
func TestProcessSemanticCeilings(t *testing.T) {
	base := func() *Process {
		return &Process{PID: 4242, UID: 1001, BootID: "6f1c0e2a-8b44-4d1e-9a51-2f0c3d7e9b10",
			StartTime: &StartTime{Kind: StartLinux, Ticks: "123456789", Hz: 100}}
	}
	if err := ValidateProcessSemantics(base()); err != nil {
		t.Fatalf("valid control: %v", err)
	}
	cases := map[string]func(*Process){
		"ticks exactly 2^64-1": func(p *Process) { p.StartTime.Ticks = "18446744073709551615" },
		"ticks 2^64":           func(p *Process) { p.StartTime.Ticks = "18446744073709551616" },
		"ticks leading zero":   func(p *Process) { p.StartTime.Ticks = "01" },
		"ticks zero":           func(p *Process) { p.StartTime.Ticks = "0" },
		"hz zero":              func(p *Process) { p.StartTime.Hz = 0 },
		"pid zero":             func(p *Process) { p.PID = 0 },
		"pid max":              func(p *Process) { p.PID = 2147483647 },
		"pid above max":        func(p *Process) { p.PID = 2147483648 },
		"uid max-1 sentinel":   func(p *Process) { p.UID = 4294967295 },
		"uid top valid":        func(p *Process) { p.UID = 4294967294 },
		"uid negative":         func(p *Process) { p.UID = -1 },
		"uppercase boot_id":    func(p *Process) { p.BootID = "6F1C0E2A-8B44-4D1E-9A51-2F0C3D7E9B10" },
		"sec exactly 2^64-1": func(p *Process) {
			p.StartTime = &StartTime{Kind: StartDarwin, Sec: "18446744073709551615", Usec: 999999}
		},
		"sec 2^64": func(p *Process) {
			p.StartTime = &StartTime{Kind: StartDarwin, Sec: "18446744073709551616", Usec: 0}
		},
		"usec 1000000": func(p *Process) {
			p.StartTime = &StartTime{Kind: StartDarwin, Sec: "1", Usec: 1000000}
		},
	}
	wantValid := map[string]bool{
		"ticks exactly 2^64-1": true, "ticks zero": true, "pid max": true,
		"uid top valid": true, "sec exactly 2^64-1": true,
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			p := base()
			mutate(p)
			err := ValidateProcessSemantics(p)
			if wantValid[name] && err != nil {
				t.Fatalf("boundary control refused: %v", err)
			}
			if !wantValid[name] && err == nil {
				t.Fatal("value above its ceiling accepted")
			}
		})
	}
	if err := ValidateProcessSemantics(nil); err == nil {
		t.Fatal("missing process accepted as process evidence")
	}
}

// Claim: an over-ceiling bound holder is a usage refusal beside the accepted
// boundary control, through the same ParseHolder callers use.
func TestParseHolderSemanticCeilingRefusal(t *testing.T) {
	holder := func(ticks string) string {
		return `{"kind":"bound","launcher":"curator-run","manager_id":"manager-1","manager_registered_at":"2026-10-27T00:00:00Z","request_id":"request-1","werkstatt":null,"session":"SES-fixture","generation":7,"process":{"pid":4242,"uid":1001,"boot_id":"6f1c0e2a-8b44-4d1e-9a51-2f0c3d7e9b10","start_time":{"kind":"linux-stat-starttime","ticks":"` + ticks + `","hz":100}}}`
	}
	if _, err := ParseHolder([]byte(holder("18446744073709551615"))); err != nil {
		t.Fatalf("boundary control refused: %v", err)
	}
	_, err := ParseHolder([]byte(holder("18446744073709551616")))
	refusal, ok := err.(*Refusal)
	if !ok || refusal.Code != "usage" || refusal.Reason != "invalid_holder" {
		t.Fatalf("over-ceiling holder not a closed usage refusal: %v", err)
	}
}
