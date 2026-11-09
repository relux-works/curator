package profilegraph

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"unicode/utf8"
)

// boundedBuffer refuses before growth. It never exposes a partial payload.
type boundedBuffer struct {
	bytes.Buffer
	limit int
	err   error
}

func (b *boundedBuffer) Write(p []byte) (int, error) {
	if b.err != nil {
		return 0, b.err
	}
	if len(p) > b.limit-b.Len() {
		b.err = refusal("graph_limit_exceeded")
		return 0, b.err
	}
	return b.Buffer.Write(p)
}
func (b *boundedBuffer) WriteString(s string) (int, error) {
	if b.err != nil {
		return 0, b.err
	}
	if len(s) > b.limit-b.Len() {
		b.err = refusal("graph_limit_exceeded")
		return 0, b.err
	}
	return b.Buffer.WriteString(s)
}

// JSON streams the typed wire model into a bounded private buffer. Encoding a
// complete graph through encoding/json would allocate its full encoded size
// before its writer can enforce the output cap.
func JSON(g *Graph, limit int) ([]byte, error) {
	b := &boundedBuffer{limit: limit}
	encodeValue(b, reflect.ValueOf(g), 0)
	_, _ = b.WriteString("\n")
	if b.err != nil {
		return nil, b.err
	}
	return b.Bytes(), nil
}
func encodeValue(b *boundedBuffer, v reflect.Value, depth int) {
	if b.err != nil {
		return
	}
	if v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		if v.IsNil() {
			_, _ = b.WriteString("null")
			return
		}
		encodeValue(b, v.Elem(), depth)
		return
	}
	indent := func(d int) { _, _ = b.WriteString("\n" + strings.Repeat("  ", d)) }
	switch v.Kind() {
	case reflect.Struct:
		_, _ = b.WriteString("{")
		count := 0
		for i := 0; i < v.NumField(); i++ {
			field := v.Type().Field(i)
			tag := strings.Split(field.Tag.Get("json"), ",")
			if field.PkgPath != "" || tag[0] == "-" {
				continue
			}
			value := v.Field(i)
			if len(tag) > 1 && tag[1] == "omitempty" && (value.IsZero() || (value.Kind() == reflect.Slice && value.Len() == 0)) {
				continue
			}
			if count > 0 {
				_, _ = b.WriteString(",")
			}
			count++
			indent(depth + 1)
			encodeScalar(b, tag[0])
			_, _ = b.WriteString(": ")
			encodeValue(b, value, depth+1)
		}
		if count > 0 {
			indent(depth)
		}
		_, _ = b.WriteString("}")
	case reflect.Slice, reflect.Array:
		if v.Kind() == reflect.Slice && v.IsNil() {
			_, _ = b.WriteString("null")
			return
		}
		_, _ = b.WriteString("[")
		for i := 0; i < v.Len(); i++ {
			if i > 0 {
				_, _ = b.WriteString(",")
			}
			indent(depth + 1)
			encodeValue(b, v.Index(i), depth+1)
		}
		if v.Len() > 0 {
			indent(depth)
		}
		_, _ = b.WriteString("]")
	default:
		encodeScalar(b, v.Interface())
	}
}
func encodeScalar(b *boundedBuffer, v any) {
	if s, ok := v.(string); ok {
		// Exact JSON string length, including replacement of invalid UTF-8 and
		// the two JavaScript line separators, before allocating escaped bytes.
		n := 2
		for i := 0; i < len(s); {
			c, size := utf8.DecodeRuneInString(s[i:])
			i += size
			switch {
			case c == '"' || c == '\\' || c == '\n' || c == '\r' || c == '\t' || c == '\b' || c == '\f':
				n += 2
			case c < 0x20 || c == '\u2028' || c == '\u2029':
				n += 6
			case c == utf8.RuneError && size == 1:
				n += 6
			default:
				n += size
			}
			if n > b.limit-b.Len() {
				b.err = refusal("graph_limit_exceeded")
				return
			}
		}
		if n > b.limit-b.Len() {
			b.err = refusal("graph_limit_exceeded")
			return
		}
	}
	// Scalar encoding is bounded above; no graph-sized temporary is allocated.
	var scalar bytes.Buffer
	enc := json.NewEncoder(&scalar)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		if b.err == nil {
			b.err = refusal("graph_invalid_input")
		}
		return
	}
	_, _ = b.Write(scalar.Bytes()[:scalar.Len()-1]) // Omit Encode's final LF.
}
