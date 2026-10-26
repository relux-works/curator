package projectadapter

import (
	"encoding/json"
	"reflect"
	"testing"
)

// Claim: all adapters share one receipt/report vocabulary, preserving Codex's
// tagged candidate status shape and Muse's mode aliases. These inert values do
// not certify a launch; a typed MCP refusal retains its distinct diagnostic.
func TestSharedAdapterTypes(t *testing.T) {
	if ModeStrict != Strict || ModeNativeProject != NativeProject {
		t.Fatal("mode alias changed")
	}
	receipt := Surface{SHA256: "digest", Size: 7, Mode: 0600}
	b, err := json.Marshal(receipt)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != `{"sha256":"digest","size":7,"mode":384}` {
		t.Fatal("shared receipt spelling changed", string(b))
	}
	report := Certification{Notice: NativeProjectNotice, Environment: "muse", ProviderVersion: MuseVersion, Platform: "macos", Mode: NativeProject, Certified: []SurfaceName{}, Refused: []Refusal{{Permissions, "manager_permissions_unqualified"}}}
	b, err = json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"notice":"project sources not isolated","environment":"muse","provider_version":"1.4.11-R6667.1","platform":"macos","mode":"native-project","certified":[],"refused":[{"surface":"permissions","reason":"manager_permissions_unqualified"}]}`
	if string(b) != want {
		t.Fatal("candidate report mapping changed", string(b))
	}
	var roundtrip Certification
	if err = json.Unmarshal(b, &roundtrip); err != nil || !reflect.DeepEqual(report, roundtrip) {
		t.Fatal("report did not roundtrip", err)
	}
	report.Form = "exec"
	b, err = json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(b, &roundtrip); err != nil || roundtrip.Form != "exec" {
		t.Fatal("launch form lost", err)
	}
	if (&Unsupported{MCP, "native_writer_unqualified"}).Code() != "mcp_contribution_policy_refused" || (&Unsupported{Permissions, "unqualified"}).Code() != "project_context_unsupported" {
		t.Fatal("refusal family changed")
	}
	contribution, err := MuseLaunch(MuseVersion, "macos", NativeProject, "exec")
	if err == nil || !reflect.DeepEqual(contribution, LaunchContribution{}) {
		t.Fatal("shared superset enabled refused launch")
	}
}
