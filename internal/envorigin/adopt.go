package envorigin

import (
	"encoding/json"
	"time"
)

// AdoptRequest is explicit historical input. SelectorValue's nil versus empty
// distinction must come from CLI flag presence, never an environment default.
// Building these bytes does not admit an adoption: the action owner must still
// verify installed marker/lock, credential entries and a complete durable unit.
type AdoptRequest struct {
	Adapter       string
	Binding       Binding
	OperatorHome  string
	SelectorState string
	SelectorValue *string
	NativeRoot    string
}

// Attest prepares a closed carrier without account/selector discovery or live
// writes. Existing authority is immutable; an identical repeated attestation
// returns its exact original bytes and capture timestamp/provenance.
func Attest(manager string, req AdoptRequest, existing []byte) ([]byte, bool, error) {
	r := &Record{Record: Version, Adapter: req.Adapter, Binding: req.Binding, OperatorHome: req.OperatorHome, Selector: Selector{Name: selectorName(req.Adapter), State: req.SelectorState, Value: req.SelectorValue}, NativeRoot: req.NativeRoot, Capture: Capture{Method: "adoption", Authority: "explicit-operator-input", Time: time.Now().UTC().Format(time.RFC3339Nano)}}
	raw, err := r.Marshal()
	if err != nil {
		return nil, false, err
	}
	if _, err := Parse(raw, req.Binding, req.Adapter, manager); err != nil {
		return nil, false, err
	}
	if len(existing) == 0 {
		return raw, true, nil
	}
	prior, err := Parse(existing, req.Binding, req.Adapter, manager)
	if err != nil {
		return nil, false, err
	}
	left, _ := json.Marshal(prior.Selector)
	right, _ := json.Marshal(r.Selector)
	if prior.OperatorHome != r.OperatorHome || prior.NativeRoot != r.NativeRoot || string(left) != string(right) {
		return nil, false, invalid("explicit attestation differs from existing recorded origin; authority cannot be replaced")
	}
	return append([]byte(nil), existing...), false, nil
}
