package protocoljson

import "testing"

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
