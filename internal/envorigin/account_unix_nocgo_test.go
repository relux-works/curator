//go:build (darwin || linux) && !cgo

package envorigin

import (
	"strings"
	"testing"
)

// The release-shaped parser consumes raw account database fields, never the
// final authority verdict or ambient HOME. No OS account lookup runs here.
func TestNativeAccountRecordPolicy(t *testing.T) {
	for _, format := range []struct {
		name, valid  string
		fields, home int
	}{
		{"darwin", "fixture:*:501:20::0:0:Fixture:/fixture/account:/bin/sh\n", 10, 8},
		{"linux", "fixture:x:501:20:Fixture:/fixture/account:/bin/sh\n", 7, 5},
	} {
		t.Run(format.name, func(t *testing.T) {
			account, err := parseAccountEntry(format.valid, 501, 501, format.fields, format.home)
			if err != nil || account.Home != "/fixture/account" || account.UID != 501 {
				t.Fatalf("valid database record: %+v %v", account, err)
			}
			for _, raw := range []string{format.valid + format.valid, strings.Replace(format.valid, ":501:", ":0501:", 1), strings.Replace(format.valid, ":501:", ":502:", 1), strings.Replace(format.valid, "/fixture/account", "relative", 1), strings.Replace(format.valid, "fixture:", ":", 1)} {
				if _, err := parseAccountEntry(raw, 501, 501, format.fields, format.home); err == nil {
					t.Fatal("invalid database evidence accepted")
				}
			}
			if _, err := parseAccountEntry(format.valid, 502, 501, format.fields, format.home); err == nil {
				t.Fatal("real/effective identity mismatch accepted")
			}
		})
	}
}
