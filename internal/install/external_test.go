package install

import (
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/devsub"
	"github.com/relux-works/curator/internal/skillspec"
)

func TestRevisionSubstitutionDerivesEffectiveObjectFormat(t *testing.T) {
	repository := skillspec.BuildRepository{
		Name: "tools", Identity: "example.test/tools", Transport: "https",
		LockedCommit: skillspec.LockedCommit{ObjectFormat: "sha1", Hex: strings.Repeat("1", 40)},
	}
	substitution := &devsub.BuildRepositorySubstitution{
		Identity: "mirror.example.test/tools", Transport: "https", RefKind: "revision", RefValue: strings.Repeat("2", 64),
	}
	effective, err := effectiveRepository("project", repository, substitution)
	if err != nil {
		t.Fatal(err)
	}
	if effective.ObjectFormat != "sha256" || effective.Commit != substitution.RefValue {
		t.Fatalf("effective revision = %+v, want sha256 %s", effective, substitution.RefValue)
	}
}
