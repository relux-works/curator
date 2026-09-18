package artifactpolicy

import "strconv"

// EffectiveLabels reports the closed assurance-policy labels a source-audit
// evidence report binds: the policy identity, its major version, the
// detector set, and the limit vector the verdict was admitted under. The
// labels are constant for the compiled policy; any policy change ships new
// constants and ages stored source-audit bindings out through the policy
// digest instead of silently re-authorizing old verdicts.
func EffectiveLabels() []string {
	return []string{
		"policy:" + PolicyID,
		"policy-version:" + strconv.Itoa(PolicyVersion),
		"detectors:" + DetectorRegistryID,
		"limits:" + LimitVectorID,
	}
}
