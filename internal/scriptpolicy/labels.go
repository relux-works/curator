package scriptpolicy

// EffectiveLabels reports the closed script-policy labels a source-audit
// evidence report binds. This manager does not implement script-worker-v1,
// so every enforced command is refused at admission; the label records that
// posture so a future worker ships a new label and ages stored bindings out
// through the policy digest instead of inheriting verdicts taken under a
// different containment promise.
func EffectiveLabels() []string {
	return []string{
		"script-worker-v1:" + StateUnsupported,
	}
}
