package contextmaterialize

import "github.com/relux-works/curator/internal/contextlock"

// Part records a renderer-owned part or an exact module occurrence. Slot is
// the original manifest position, including modules excluded by a selector.
type Part struct {
	Kind, Package, Path string
	Slot                int
	Bytes               []byte
}

// Span is a half-open byte interval in the actual Join output.
type Span struct {
	Part
	Start, End int
}

// Rendered couples the emitted bytes with their additive leaf attribution.
type Rendered struct {
	Bytes []byte
	Spans []Span
}

func modulePart(name string, pkg Package, module Module) Part {
	for slot, m := range pkg.Modules {
		if m.Path == module.Path {
			return Part{Kind: "module", Package: name, Path: m.Path, Slot: slot, Bytes: module.Bytes}
		}
	}
	panic("applicable module missing from package")
}

// JoinParts instruments Join itself; attribution never uses content matching.
func JoinParts(parts []Part) Rendered {
	raw := make([][]byte, len(parts))
	for i, p := range parts {
		raw[i] = p.Bytes
	}
	out := Rendered{Bytes: Join(raw), Spans: []Span{}}
	offset := 0
	for i, p := range parts {
		if i > 0 {
			out.Spans = append(out.Spans, Span{Part: Part{Kind: "separator", Bytes: out.Bytes[offset : offset+1]}, Start: offset, End: offset + 1})
			offset++
		}
		if len(p.Bytes) > 0 {
			out.Spans = append(out.Spans, Span{Part: p, Start: offset, End: offset + len(p.Bytes)})
		}
		offset += len(p.Bytes)
	}
	return out
}

// TraceMonolithic uses the same emitter as Monolithic, with a byte ledger.
func TraceMonolithic(lock *contextlock.Lock, hash string, precedence Precedence, environment string, packages map[string]Package) (Rendered, bool, error) {
	var trace Rendered
	_, written, err := monolithic(lock, hash, precedence, environment, packages, &trace)
	return trace, written, err
}

// TraceReferenced includes the root, every fragment and generated configuration.
func TraceReferenced(lock *contextlock.Lock, hash string, precedence Precedence, environment string, packages map[string]Package) (map[string]Rendered, bool, error) {
	traces := map[string]Rendered{}
	_, written, err := referenced(lock, hash, precedence, environment, packages, traces)
	return traces, written, err
}

// TraceSystemPrompt uses the same admission and emission path as SystemPrompt.
func TraceSystemPrompt(lock *contextlock.Lock, precedence Precedence, environment string, packages map[string]Package, admission Admission) (Rendered, bool, error) {
	var trace Rendered
	_, written, _, err := systemPrompt(lock, precedence, environment, packages, admission, &trace)
	return trace, written, err
}
