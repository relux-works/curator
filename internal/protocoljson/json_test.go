package protocoljson

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestValidate(t *testing.T) {
	valid := []string{
		`{"a":[1,1.5,true,null],"s":"\ud83d\ude00"}`,
		` {"extension":{"integer":9007199254740992}} `,
	}
	for _, payload := range valid {
		if err := Validate([]byte(payload)); err != nil {
			t.Errorf("Validate(%s): %v", payload, err)
		}
	}
	invalid := [][]byte{
		[]byte("\xef\xbb\xbf{}"),
		[]byte(`{"a":1,"a":2}`),
		[]byte(`{"s":"\ud800"}`),
		[]byte(`{"s":"\udc00"}`),
		[]byte(`{} trailing`),
		{0xff},
	}
	for _, payload := range invalid {
		if err := Validate(payload); err == nil {
			t.Errorf("Validate(%q) succeeded, want error", payload)
		}
	}
}

func TestMarshalCanonical(t *testing.T) {
	value := map[string]any{
		"z": []any{"м", json.Number("2"), "\b\f<>/&"},
		"a": json.Number("-12"),
		"b": true,
		"n": nil,
	}
	got, err := MarshalCanonical(value)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"a":-12,"b":true,"n":null,"z":["м",2,"\b\f<>/&"]}`
	if string(got) != want {
		t.Fatalf("canonical bytes:\n got %s\nwant %s", got, want)
	}
}

func TestMarshalCanonicalRejectsUnsupportedValues(t *testing.T) {
	for name, value := range map[string]any{
		"fraction":       json.Number("1.5"),
		"exponent":       json.Number("1e2"),
		"negative zero":  json.Number("-0"),
		"unsafe integer": json.Number("9007199254740992"),
		"float":          1.0,
		"invalid utf8":   string([]byte{0xff}),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := MarshalCanonical(map[string]any{"value": value}); err == nil {
				t.Fatal("value was accepted")
			}
		})
	}
}

func TestRequireCanonical(t *testing.T) {
	canonical := []byte(`{"a":1,"s":"é"}`)
	if err := RequireCanonical(canonical); err != nil {
		t.Fatalf("canonical payload rejected: %v", err)
	}

	invalid := map[string][]byte{
		"leading whitespace":  []byte(` {"a":1,"s":"é"}`),
		"pretty printed":      []byte("{\n  \"a\": 1,\n  \"s\": \"é\"\n}"),
		"trailing newline":    append(append([]byte{}, canonical...), '\n'),
		"escaped nonascii":    []byte(`{"a":1,"s":"\u00e9"}`),
		"escaped slash":       []byte(`{"a":1,"s":"é\/"}`),
		"wrong member order":  []byte(`{"s":"é","a":1}`),
		"duplicate member":    []byte(`{"a":1,"a":1}`),
		"negative zero":       []byte(`{"a":-0}`),
		"noninteger":          []byte(`{"a":1.0}`),
		"byte order mark":     append([]byte{0xef, 0xbb, 0xbf}, canonical...),
		"terminal whitespace": append(append([]byte{}, canonical...), ' '),
	}
	for name, payload := range invalid {
		t.Run(name, func(t *testing.T) {
			if err := RequireCanonical(payload); err == nil {
				t.Fatal("noncanonical payload was accepted")
			}
		})
	}
}

func TestCanonicalEqualAndUnmarshal(t *testing.T) {
	value := map[string]any{"z": "last", "a": json.Number("1")}
	payload := []byte(`{"a":1,"z":"last"}`)
	equal, err := CanonicalEqual(payload, value)
	if err != nil || !equal {
		t.Fatalf("CanonicalEqual() = %v, %v", equal, err)
	}
	equal, err = CanonicalEqual([]byte(` {"a":1,"z":"last"}`), value)
	if err != nil || equal {
		t.Fatalf("CanonicalEqual(noncanonical) = %v, %v", equal, err)
	}

	var decoded struct {
		A int64 `json:"a"`
	}
	if err := UnmarshalCanonical([]byte(`{"a":1,"extra":true}`), &decoded); err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("unknown field error = %v", err)
	}
	if err := UnmarshalCanonical([]byte("{\n\"a\":1\n}"), &decoded); err == nil {
		t.Fatal("noncanonical bytes were accepted")
	}
	if err := UnmarshalCanonical([]byte(`{"a":1}`), &decoded); err != nil || decoded.A != 1 {
		t.Fatalf("UnmarshalCanonical() = %+v, %v", decoded, err)
	}
}

func TestMarshalCanonicalIntegerAndCollectionTypes(t *testing.T) {
	value := map[string]any{
		"bool":     false,
		"controls": "\x00\x01\b\f\n\r\t",
		"ints":     []any{int(-1), int8(2), int16(3), int32(4), int64(5)},
		"uints":    []any{uint(1), uint8(2), uint16(3), uint32(4), uint64(5)},
		"strings":  []string{"a", "b"},
		"object":   map[string]string{"b": "2", "a": "1"},
	}
	got, err := MarshalCanonical(value)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"bool":false,"controls":"\u0000\u0001\b\f\n\r\t","ints":[-1,2,3,4,5],"object":{"a":"1","b":"2"},"strings":["a","b"],"uints":[1,2,3,4,5]}`
	if string(got) != want {
		t.Fatalf("canonical bytes:\n got %s\nwant %s", got, want)
	}
	if _, err := MarshalCanonical(uint64(MaxSafeInteger) + 1); err == nil {
		t.Fatal("unsafe unsigned integer was accepted")
	}
	if _, err := MarshalCanonical(map[string]any{string([]byte{0xff}): true}); err == nil {
		t.Fatal("invalid UTF-8 object key was accepted")
	}
	if _, err := CanonicalEqual(nil, make(chan int)); err == nil {
		t.Fatal("unsupported comparison value was accepted")
	}
}
