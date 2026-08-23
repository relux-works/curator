package artifactpolicy

import "testing"

func TestNodeManagerNPMRCIsClosedTextMetadata(t *testing.T) {
	for _, name := range []string{".npmrc", ".yarnrc"} {
		declaration, ok := metadataDeclaration(name)
		if !ok || declaration.Grammar != GrammarPlain || declaration.Class != ClassTextMetadata {
			t.Fatalf("%s declaration = %#v, %v", name, declaration, ok)
		}
		if !profileAllowsGrammar(ProfileNodeV1, declaration.Grammar) {
			t.Fatalf("node profile rejected the closed %s text grammar", name)
		}
	}
}

func TestDependencyPatchIsClosedTextMetadata(t *testing.T) {
	for _, name := range []string{"dependency.patch", "dependency.diff"} {
		declaration, ok := metadataDeclaration(name)
		if !ok || declaration.Grammar != GrammarPlain || declaration.Class != ClassTextMetadata {
			t.Fatalf("%s declaration = %#v, %v", name, declaration, ok)
		}
	}
}
