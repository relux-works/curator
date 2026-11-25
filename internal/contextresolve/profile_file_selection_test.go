package contextresolve

import (
	"slices"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/whitelist"
)

func profileSkillFilesEdge(selection *whitelist.FileSelection) constraint {
	return constraint{requirement: Requirement{
		Kind: contextlock.KindSkill, Name: "guide", ProfileSkillRequirement: true, Files: selection,
	}}
}

func TestProfileSkillFileSelectionEdgesRequireAgreement(t *testing.T) {
	left := &whitelist.FileSelection{
		Include: []string{"data/alpha.json", "data/bravo.json"},
		Exclude: []string{".cache/**", "build/tmp"},
	}
	changed := &whitelist.FileSelection{
		Include: []string{"data/alpha.json", "data/charlie.json"},
		Exclude: []string{".cache/**", "build/tmp"},
	}
	reordered := &whitelist.FileSelection{
		Include: []string{"data/bravo.json", "data/alpha.json"},
		Exclude: []string{"build/tmp", ".cache/**"},
	}

	t.Run("different selections refuse", func(t *testing.T) {
		r := &resolver{constraints: map[string][]constraint{"guide": {
			profileSkillFilesEdge(left), profileSkillFilesEdge(changed),
		}}}
		selection, err := r.profileSkillFiles("guide")
		if err == nil || !strings.Contains(err.Error(), DiagProfileSkillFilesConflict) {
			t.Fatalf("selection=%+v error=%v, want %s refusal", selection, err, DiagProfileSkillFilesConflict)
		}
	})

	t.Run("equivalent list orders agree", func(t *testing.T) {
		r := &resolver{constraints: map[string][]constraint{"guide": {
			profileSkillFilesEdge(left), profileSkillFilesEdge(reordered),
		}}}
		selection, err := r.profileSkillFiles("guide")
		if err != nil {
			t.Fatal(err)
		}
		if selection == nil || !slices.Equal(selection.Include, []string{"data/alpha.json", "data/bravo.json"}) || !slices.Equal(selection.Exclude, []string{".cache/**", "build/tmp"}) {
			t.Fatalf("canonical selection=%+v, want sorted equivalent rules", selection)
		}
	})
}
