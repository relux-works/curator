package muse1413

import "testing"

// Contradictory loader fields cannot attest a successful zero-skill result.
// Retained under the reviewer's original test name from review revision 1.
func TestReviewContradictoryObservation(t *testing.T) {
	want := map[string]int{"user_skills": 0}
	if err := checkCounts(`event="skills.catalog_load" outcome="ready" user_skills=0`, "skills.catalog_load", want); err != nil {
		t.Fatal(err)
	}
	for name, log := range map[string]string{
		"count-zero-then-one":       `event="skills.catalog_load" outcome="ready" user_skills=0 user_skills=1`,
		"count-one-then-zero":       `event="skills.catalog_load" outcome="ready" user_skills=1 user_skills=0`,
		"outcome-failed-then-ready": `event="skills.catalog_load" outcome="failed" outcome="ready" user_skills=0`,
		"outcome-ready-then-failed": `event="skills.catalog_load" outcome="ready" outcome="failed" user_skills=0`,
		"duplicate-count":           `event="skills.catalog_load" outcome="ready" user_skills=0 user_skills=0`,
		"duplicate-outcome":         `event="skills.catalog_load" outcome="ready" outcome="ready" user_skills=0`,
		"duplicate-event":           `event="skills.catalog_load" event="rules.context_load" outcome="ready" user_skills=0`,
		"quoted-decoy":              `message="event=\"skills.catalog_load\" outcome=\"ready\" user_skills=0"`,
		"malformed-count":           `event="skills.catalog_load" outcome="ready" user_skills=0garbage`,
		"overflow-count":            `event="skills.catalog_load" outcome="ready" user_skills=999999999999999999999999999`,
	} {
		t.Run(name, func(t *testing.T) {
			if err := checkCounts(log, "skills.catalog_load", want); err == nil {
				t.Fatalf("contradictory record admitted: %s", log)
			}
		})
	}
}
