package storecompatibility

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"
)

const replayPackage = "github.com/relux-works/curator/internal/envprofile"

type event struct {
	Action      string          `json:"Action"`
	Package     string          `json:"Package"`
	Test        string          `json:"Test"`
	Output      string          `json:"Output"`
	Time        json.RawMessage `json:"Time"`
	Elapsed     json.RawMessage `json:"Elapsed"`
	FailedBuild json.RawMessage `json:"FailedBuild"`
}

func decodeEvent(raw []byte) (event, error) {
	if !utf8.Valid(raw) {
		return event{}, errors.New("invalid UTF-8 event")
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	first, err := d.Token()
	if err != nil || first != json.Delim('{') {
		return event{}, errors.New("event object required")
	}
	seen := map[string]bool{}
	for d.More() {
		key, err := d.Token()
		if err != nil {
			return event{}, err
		}
		name, ok := key.(string)
		if !ok || seen[name] {
			return event{}, errors.New("duplicate event key")
		}
		seen[name] = true
		var value json.RawMessage
		if err := d.Decode(&value); err != nil {
			return event{}, err
		}
	}
	if _, err := d.Token(); err != nil {
		return event{}, err
	}
	if _, err := d.Token(); err != io.EOF {
		return event{}, errors.New("trailing event JSON")
	}
	d = json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	var e event
	if err := d.Decode(&e); err != nil {
		return event{}, err
	}
	return e, nil
}

// Report separates compatible legacy expectations from intentionally changed
// ones. Incompatible receipts are not passing product negative controls.
type Report struct {
	Property          string   `json:"property"`
	FullV1Conformance bool     `json:"full_v1_store_conformance"`
	Compatible        []string `json:"compatible"`
	Incompatible      []string `json:"incompatible"`
}

// Check requires the actual complete go-test JSON stream, including package
// termination and the intended assertion witness for every incompatible row.
// An unrelated crash, setup failure, skip, missing/duplicate row, or same-count
// different partition cannot satisfy this migration property.
func Check(input io.Reader) (Report, error) {
	report := Report{Property: "historical store migration compatibility; not full v1 store conformance"}
	want := map[string]expectation{}
	tops := map[string]bool{}
	for _, row := range roster {
		want[row.ID] = row
		tops[strings.SplitN(row.ID, "/", 2)[0]] = true
	}
	started, ended := map[string]bool{}, map[string]bool{}
	topStarted, topEnded := map[string]bool{}, map[string]bool{}
	outputs := map[string]string{}
	nonExecution := regexp.MustCompile(`\b[1-9][0-9]* (bound|skipped|known-gap)\b`)
	packageStart, packageEnd := false, false
	scanner := bufio.NewScanner(io.LimitReader(input, 64<<20+1))
	scanner.Buffer(make([]byte, 4096), 4<<20)
	consumed := 0
	for scanner.Scan() {
		raw := scanner.Bytes()
		consumed += len(raw) + 1
		if consumed > 64<<20 {
			return Report{}, errors.New("replay exceeds byte bound")
		}
		e, err := decodeEvent(raw)
		if err != nil {
			return Report{}, fmt.Errorf("invalid go-test event: %w", err)
		}
		if packageEnd {
			return Report{}, errors.New("event after package termination")
		}
		if e.Package != replayPackage {
			return Report{}, errors.New("unexpected replay package")
		}
		if e.Test == "" {
			switch e.Action {
			case "start":
				if packageStart {
					return Report{}, errors.New("duplicate package start")
				}
				packageStart = true
			case "fail":
				packageEnd = true
			case "output":
			default:
				return Report{}, fmt.Errorf("unexpected package verdict %q", e.Action)
			}
			continue
		}
		if !packageStart {
			return Report{}, errors.New("test before package start")
		}
		row, caseExists := want[e.Test]
		if !caseExists && !tops[e.Test] {
			return Report{}, fmt.Errorf("unexpected case ID %s", e.Test)
		}
		if !caseExists {
			switch e.Action {
			case "run":
				if topStarted[e.Test] {
					return Report{}, errors.New("duplicate top-level replay")
				}
				topStarted[e.Test] = true
			case "output":
				if nonExecution.MatchString(e.Output) {
					return Report{}, errors.New("replay contains unexecuted or gap classifications")
				}
			case "pass", "fail":
				wantFail := strings.Contains(e.Test, "DryRun") || strings.Contains(e.Test, "RepairVectors")
				if !topStarted[e.Test] || topEnded[e.Test] || (e.Action == "fail") != wantFail {
					return Report{}, errors.New("unexpected top-level replay verdict")
				}
				topEnded[e.Test] = true
			default:
				return Report{}, errors.New("top-level replay did not execute")
			}
			continue
		}
		switch e.Action {
		case "run":
			if started[e.Test] {
				return Report{}, errors.New("duplicate case run")
			}
			started[e.Test] = true
		case "output":
			if len(outputs[e.Test])+len(e.Output) > 1<<20 {
				return Report{}, errors.New("case output exceeds bound")
			}
			outputs[e.Test] += e.Output
		case "pass", "fail":
			if !started[e.Test] || ended[e.Test] {
				return Report{}, errors.New("missing run or duplicate case verdict")
			}
			ended[e.Test] = true
			if row.Check == "" {
				if e.Action != "pass" {
					return Report{}, fmt.Errorf("compatible case failed: %s", e.Test)
				}
				report.Compatible = append(report.Compatible, e.Test)
			} else {
				out := outputs[e.Test]
				if e.Action != "fail" || !strings.Contains(out, "environment_store_untrusted:") || !strings.Contains(out, "failed "+row.Check+" check:") || !strings.Contains(out, row.Witness) {
					return Report{}, fmt.Errorf("incompatible case lacks intended assertion witness: %s", e.Test)
				}
				report.Incompatible = append(report.Incompatible, e.Test)
			}
		default:
			return Report{}, fmt.Errorf("case did not execute to its intended verdict: %s/%s", e.Test, e.Action)
		}
	}
	if err := scanner.Err(); err != nil {
		return Report{}, err
	}
	if !packageStart || !packageEnd || len(topEnded) != 4 || len(ended) != 38 || len(report.Compatible) != 32 || len(report.Incompatible) != 6 {
		return Report{}, errors.New("incomplete exact 38-row 32/6 migration replay")
	}
	sort.Strings(report.Compatible)
	sort.Strings(report.Incompatible)
	return report, nil
}
