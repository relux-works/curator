package rc5interop

import (
	"os"
	"testing"

	"github.com/relux-works/curator/internal/conformanceconsumer"
)

func TestEveryAcceptedRC5CaseHasACuratorBinding(t *testing.T) {
	root := os.Getenv("CURATOR_EXTERNAL_REPOSITORY_CORPUS_ROOT")
	if root == "" {
		t.Skip("CURATOR_EXTERNAL_REPOSITORY_CORPUS_ROOT is not set")
	}
	corpus, err := conformanceconsumer.OpenCorpus(root, conformanceconsumer.RC5Boundary)
	if err != nil {
		t.Fatal(err)
	}
	cases, err := corpus.Cases()
	if err != nil {
		t.Fatal(err)
	}
	if len(cases) != 60 || len(Bindings) != 60 {
		t.Fatalf("case/binding inventory = %d/%d, want 60/60", len(cases), len(Bindings))
	}
	seen := make(map[string]struct{}, len(cases))
	for _, item := range cases {
		seen[item.ID] = struct{}{}
		binding, ok := Bindings[item.ID]
		if !ok {
			t.Errorf("shared case %q has no Curator binding", item.ID)
			continue
		}
		if binding.Package == "" || binding.Test == "" {
			t.Errorf("shared case %q has an incomplete Curator binding", item.ID)
		}
	}
	for id := range Bindings {
		if _, ok := seen[id]; !ok {
			t.Errorf("Curator binding %q is not present in the accepted corpus", id)
		}
	}
}
