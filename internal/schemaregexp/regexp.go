// Package schemaregexp adapts the published ECMA schema patterns without
// confusing backend execution failures with input-validation failures.
package schemaregexp

import (
	"strings"
	"sync"
	"time"

	"github.com/dlclark/regexp2"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

// EngineError preserves the backend cause without rendering its raw input.
// jsonschema.Regexp only returns bool, so a typed panic bridges this error to
// Validate. No other panic is intercepted and no error is stored globally.
type EngineError struct{ Cause error }

func (*EngineError) Error() string   { return "schema regexp engine failure" }
func (e *EngineError) Unwrap() error { return e.Cause }

func Validate(schema *jsonschema.Schema, value any) (err error) {
	defer func() {
		if p := recover(); p != nil {
			if engine, ok := p.(*EngineError); ok {
				err = engine
			} else {
				panic(p)
			}
		}
	}()
	return schema.Validate(value)
}

// regexp2's process-global timeout clock can give a caller a stale deadline
// when concurrent matches restart it after idle. All timed production schema
// matches share this lock; the backend timeout itself remains one second.
var engineClock sync.Mutex

type timedPattern struct {
	source string
	match  func(string) (bool, error)
}

func (p timedPattern) String() string { return p.source }
func (p timedPattern) MatchString(s string) bool {
	engineClock.Lock()
	defer engineClock.Unlock()
	matched, err := p.match(s)
	if err != nil {
		panic(&EngineError{Cause: err})
	}
	return matched
}

type hexPattern struct {
	source, prefix string
	sizes          []int
}

func (p hexPattern) String() string { return p.source }
func (p hexPattern) MatchString(s string) bool {
	if !strings.HasPrefix(s, p.prefix) {
		return false
	}
	s = s[len(p.prefix):]
	lengthOK := false
	for _, size := range p.sizes {
		if len(s) == size {
			lengthOK = true
		}
	}
	if !lengthOK {
		return false
	}
	for i := 0; i < len(s); i++ {
		if !(s[i] >= '0' && s[i] <= '9' || s[i] >= 'a' && s[i] <= 'f') {
			return false
		}
	}
	return true
}

func Compile(pattern string) (jsonschema.Regexp, error) { return compile(pattern, compileTimed) }
func compile(pattern string, fallback jsonschema.RegexpEngine) (jsonschema.Regexp, error) {
	p := hexPattern{source: pattern}
	switch pattern {
	case `^[0-9a-f]{40}(?:[0-9a-f]{24})?$`, `^(?:[0-9a-f]{40}|[0-9a-f]{64})$`:
		p.sizes = []int{40, 64}
	case `^[0-9a-f]{40}$`:
		p.sizes = []int{40}
	case `^[0-9a-f]{64}$`:
		p.sizes = []int{64}
	case `^[0-9a-f]{16}$`:
		p.sizes = []int{16}
	case `^sha256:[0-9a-f]{64}$`:
		p.prefix = "sha256:"
		p.sizes = []int{64}
	default:
		return fallback(pattern)
	}
	return p, nil
}
func compileTimed(pattern string) (jsonschema.Regexp, error) {
	re, err := regexp2.Compile(pattern, regexp2.ECMAScript)
	if err != nil {
		return nil, err
	}
	re.MatchTimeout = time.Second
	return timedPattern{pattern, re.MatchString}, nil
}
