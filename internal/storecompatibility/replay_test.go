package storecompatibility

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func validReplay() []event {
	events := []event{{Action: "start", Package: replayPackage}}
	tops := map[string]bool{}
	for _, row := range roster {
		top := strings.SplitN(row.ID, "/", 2)[0]
		if !tops[top] {
			events = append(events, event{Action: "run", Package: replayPackage, Test: top})
			tops[top] = true
		}
	}
	for _, row := range roster {
		events = append(events, event{Action: "run", Package: replayPackage, Test: row.ID})
		action := "pass"
		if row.Check != "" {
			action = "fail"
			events = append(events, event{Action: "output", Package: replayPackage, Test: row.ID, Output: "environment_store_untrusted: store entry failed " + row.Check + " check: fixture; " + row.Witness})
		}
		events = append(events, event{Action: action, Package: replayPackage, Test: row.ID})
	}
	for top := range tops {
		action := "pass"
		if strings.Contains(top, "DryRun") || strings.Contains(top, "RepairVectors") {
			action = "fail"
		}
		events = append(events, event{Action: action, Package: replayPackage, Test: top})
	}
	return append(events, event{Action: "fail", Package: replayPackage})
}
func encodeReplay(events []event) *bytes.Buffer {
	var buf bytes.Buffer
	for _, e := range events {
		_ = json.NewEncoder(&buf).Encode(e)
	}
	return &buf
}

// The valid exact partition passes. Keeping 32/6 counts while swapping the
// row IDs fails; missing rows, unrelated failures and absent intended witnesses
// cannot turn archive incompatibility into a successful migration property.
func TestExactMigrationPartitionAndControls(t *testing.T) {
	if r, err := Check(encodeReplay(validReplay())); err != nil || len(r.Compatible) != 32 || len(r.Incompatible) != 6 || r.FullV1Conformance || !strings.Contains(r.Property, "not full v1") {
		t.Fatalf("valid replay: %+v %v", r, err)
	}
	for _, name := range []string{"swapped-ID-same-count", "missing-case", "skip", "bound-classification", "crash-instead-of-assertion", "wrong-diagnostic", "wrong-boundary", "missing-witness", "duplicate-verdict", "missing-package-end", "wrong-package", "trailing-event"} {
		t.Run(name, func(t *testing.T) {
			events := validReplay()
			good, bad := -1, -1
			for i, e := range events {
				if e.Test != "" && e.Action == "pass" && good < 0 {
					good = i
				}
				if e.Test != "" && e.Action == "fail" && bad < 0 {
					bad = i
				}
			}
			switch name {
			case "swapped-ID-same-count":
				events[good].Action = "fail"
				events[bad].Action = "pass"
			case "missing-case":
				events = append(events[:good], events[good+1:]...)
			case "skip":
				events[good].Action = "skip"
			case "bound-classification":
				events = append(events[:len(events)-1], event{Action: "output", Package: replayPackage, Test: strings.SplitN(roster[0].ID, "/", 2)[0], Output: "published cases: 31 driven, 0 known-gap, 1 bound, 0 skipped"}, events[len(events)-1])
			case "crash-instead-of-assertion":
				events[bad-1].Output = "panic: unrelated fixture setup"
			case "wrong-diagnostic":
				events[bad-1].Output = strings.ReplaceAll(events[bad-1].Output, "environment_store_untrusted:", "other:")
			case "wrong-boundary":
				events[bad-1].Output = strings.ReplaceAll(events[bad-1].Output, "ownership", "containment")
			case "missing-witness":
				events[bad-1].Output = "environment_store_untrusted: store entry failed ownership check: fixture"
			case "duplicate-verdict":
				events = append(events[:good+1], append([]event{events[good]}, events[good+1:]...)...)
			case "missing-package-end":
				events = events[:len(events)-1]
			case "wrong-package":
				events[0].Package = "example.invalid/other"
			case "trailing-event":
				events = append(events, event{Action: "output", Package: replayPackage})
			}
			_, err := Check(encodeReplay(events))
			if err == nil {
				t.Fatal("forbidden replay passed")
			}
			if name == "swapped-ID-same-count" && !strings.Contains(err.Error(), "compatible case failed") {
				t.Fatalf("swap failed for an unrelated reason: %v", err)
			}
			if name == "bound-classification" && !strings.Contains(err.Error(), "unexecuted or gap") {
				t.Fatalf("bound plant failed for an unrelated reason: %v", err)
			}
		})
	}
}

// Invalid streams refuse without a partial report, alongside valid JSON above.
func TestMalformedReplay(t *testing.T) {
	for _, raw := range []string{"not JSON\n", "{}\n", strings.Repeat("x", (4<<20)+1)} {
		if r, err := Check(strings.NewReader(raw)); err == nil || r.Property != "" {
			t.Fatalf("malformed replay: %+v %v", r, err)
		}
	}
}

// Decoded-key duplicates cannot erase a fail action. A real Go event with
// timestamp and elapsed metadata remains admitted by the closed event reader.
func TestReplayEventDoesNotNormalizeDuplicateVerdicts(t *testing.T) {
	valid := `{"Time":"2026-11-13T00:00:00Z","Action":"pass","Package":"github.com/relux-works/curator/internal/envprofile","Elapsed":0.1}`
	if e, err := decodeEvent([]byte(valid)); err != nil || e.Action != "pass" {
		t.Fatalf("native-shaped event refused: %+v %v", e, err)
	}
	for _, raw := range []string{`{"Action":"fail","Action":"pass"}`, `{"Action":"fail","Act\u0069on":"pass"}`} {
		if _, err := decodeEvent([]byte(raw)); err == nil || !strings.Contains(err.Error(), "duplicate event key") {
			t.Fatalf("duplicate verdict plant: %v", err)
		}
	}
}
