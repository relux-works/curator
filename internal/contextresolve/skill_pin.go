package contextresolve

import "github.com/relux-works/curator/internal/contextlock"

// VerifySkillPin applies the existing exact-candidate signer policy to a
// shared-closure skill. It neither enumerates tags nor upgrades a version.
func VerifySkillPin(source Source, pin Requirement, signers map[string][]Signer, required bool) error {
	r := &resolver{source: source, input: Input{SourceSigners: signers, RequireSourceSigners: required}}
	return r.verifyCandidate(&selection{kind: contextlock.KindSkill, name: pin.Name, source: pin.Source, tag: pin.Tag, commit: pin.Revision})
}

// VerifyContextPin rechecks the exact immutable context origin before skill traversal.
func VerifyContextPin(source Source, pin Requirement, signers map[string][]Signer, required bool) error {
	r := &resolver{source: source, input: Input{SourceSigners: signers, RequireSourceSigners: required}}
	return r.verifyCandidate(&selection{kind: contextlock.KindContext, name: pin.Name, source: pin.Source, tag: pin.Tag, commit: pin.Revision})
}
