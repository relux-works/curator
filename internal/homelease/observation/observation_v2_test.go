package observation

import (
	"encoding/json"
	"strings"
	"testing"
)

func v2Fixture(t *testing.T) map[string]any {
	t.Helper()
	raw := `{
		"schema": "curator-run-session-observation-v2",
		"manager": {"id": "manager-1", "executable": "/opt/example/bin/manager", "registered_at": "2026-10-27T00:00:00Z"},
		"session": {
			"session": "SES-fixture", "operation_id": "11111111-1111-4111-8111-111111111111",
			"manager": "manager-1", "runtime": "codex_cli", "native_id": null,
			"readable": true, "state": "busy", "labels": {}, "generation": 7,
			"started_at": "2026-10-27T00:00:00Z", "last_activity_at": null,
			"doorbell": {"state": "off", "uncertain": false, "generation": null, "reason": null},
			"process": {"pid": 4242, "uid": 1001, "boot_id": "6f1c0e2a-8b44-4d1e-9a51-2f0c3d7e9b10",
				"start_time": {"kind": "linux-stat-starttime", "ticks": "123456789", "hz": 100}}
		},
		"observed_at": "2026-10-27T00:00:01Z",
		"capabilities": {"process_observation": true}
	}`
	var doc map[string]any
	if err := json.Unmarshal([]byte(raw), &doc); err != nil {
		t.Fatal(err)
	}
	return doc
}

func sessionOf(doc map[string]any) map[string]any { return doc["session"].(map[string]any) }

// Claim: a capable v2 reply yields its process tuple byte-for-byte, and a
// reply without process stays a valid observation that proves nothing.
func TestParseV2Controls(t *testing.T) {
	doc := v2Fixture(t)
	raw, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	record, err := ParseV2(raw)
	if err != nil {
		t.Fatalf("valid control: %v", err)
	}
	process := record.Session.Process
	if process == nil || process.PID != 4242 || process.UID != 1001 || process.StartTime == nil ||
		process.StartTime.Ticks != "123456789" || !record.ProcessObservation {
		t.Fatalf("tuple not carried byte-for-byte: %+v", process)
	}
	delete(sessionOf(doc), "process")
	raw, err = json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	record, err = ParseV2(raw)
	if err != nil || record.Session.Process != nil {
		t.Fatalf("absent process must stay valid but unproven: %+v %v", record.Session.Process, err)
	}
}

// Claim: process or termination outside their capability and state rules is
// a protocol error beside the valid down/termination control.
func TestParseV2Refusals(t *testing.T) {
	down := v2Fixture(t)
	sessionOf(down)["state"] = "down"
	delete(sessionOf(down), "process")
	sessionOf(down)["termination"] = map[string]any{"scope": "linux-cgroup", "witness": "populated-0"}
	raw, err := json.Marshal(down)
	if err != nil {
		t.Fatal(err)
	}
	record, err := ParseV2(raw)
	if err != nil || record.Session.Termination == nil || record.Session.Termination.Witness != "populated-0" {
		t.Fatalf("termination control: %+v %v", record.Session.Termination, err)
	}

	cases := map[string]func(map[string]any){
		"process without capability": func(doc map[string]any) {
			doc["capabilities"] = map[string]any{"process_observation": false}
		},
		"process on unreadable": func(doc map[string]any) { sessionOf(doc)["readable"] = false },
		"process on down":       func(doc map[string]any) { sessionOf(doc)["state"] = "down" },
		"termination outside down": func(doc map[string]any) {
			delete(sessionOf(doc), "process")
			sessionOf(doc)["termination"] = map[string]any{"scope": "linux-cgroup", "witness": "populated-0"}
		},
		"unknown termination witness": func(doc map[string]any) {
			sessionOf(doc)["state"] = "down"
			delete(sessionOf(doc), "process")
			sessionOf(doc)["termination"] = map[string]any{"scope": "linux-cgroup", "witness": "empty"}
		},
		"pid zero": func(doc map[string]any) {
			sessionOf(doc)["process"].(map[string]any)["pid"] = float64(0)
		},
		"uid sentinel": func(doc map[string]any) {
			sessionOf(doc)["process"].(map[string]any)["uid"] = float64(4294967295)
		},
		"uppercase boot_id": func(doc map[string]any) {
			sessionOf(doc)["process"].(map[string]any)["boot_id"] = "6F1C0E2A-8B44-4D1E-9A51-2F0C3D7E9B10"
		},
		"leading zero ticks": func(doc map[string]any) {
			sessionOf(doc)["process"].(map[string]any)["start_time"].(map[string]any)["ticks"] = "0123"
		},
		"ticks above uint64": func(doc map[string]any) {
			sessionOf(doc)["process"].(map[string]any)["start_time"].(map[string]any)["ticks"] = "18446744073709551616"
		},
		"hz zero": func(doc map[string]any) {
			sessionOf(doc)["process"].(map[string]any)["start_time"].(map[string]any)["hz"] = float64(0)
		},
		"legacy process member": func(doc map[string]any) {
			sessionOf(doc)["process"] = map[string]any{"pid": float64(4242), "started_at": "2026-10-27T00:00:00Z"}
		},
		"extra envelope member":  func(doc map[string]any) { doc["extra"] = true },
		"v1 schema string":       func(doc map[string]any) { doc["schema"] = "curator-run-session-observation-v1" },
		"unknown session member": func(doc map[string]any) { sessionOf(doc)["bogus"] = true },
		"missing session member": func(doc map[string]any) { delete(sessionOf(doc), "doorbell") },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			doc := v2Fixture(t)
			mutate(doc)
			raw, err := json.Marshal(doc)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := ParseV2(raw); err == nil {
				t.Fatal("forbidden v2 observation accepted")
			}
			if strings.Contains(string(raw), "SES-fixture") && err != nil && strings.Contains(err.Error(), "SES-fixture") {
				t.Fatal("refusal echoed observation bytes")
			}
		})
	}
}
