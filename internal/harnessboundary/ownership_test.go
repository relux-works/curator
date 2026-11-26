package harnessboundary

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/relux-works/curator/internal/harnessapi"
	"github.com/relux-works/curator/internal/harnessregistry"
)

const ownershipSourceHead = "643e727edd2db21307ca6d8c0332935e33d50ac8"

// Every inventory row has explicit adapter methods and a neutral effect owner.
// Each previously omitted package is scanned, with ID and path plants paired
// with a harness-neutral control. The map describes migration ownership, not
// implementation readiness of the staged methods.
func TestAdapterSurfaceOwnership(t *testing.T) {
	for _, pkg := range []string{"projectlease", "projectmcp", "envauthority", "testruntimes", "contextpkg", "envfiles", "manifest", "closure"} {
		t.Run(pkg, func(t *testing.T) {
			root := t.TempDir()
			rel := "internal/" + pkg + "/plant.go"
			plant(t, root, rel, "package "+pkg+"\nfunc generic(a,b string) bool { return a==b }\n")
			if len(scan(t, root)) != 0 {
				t.Fatal("generic control refused")
			}
			plant(t, root, rel, "package "+pkg+"\nfunc hardwired(id string) bool { return id==\"codex_cli\" }\n")
			requireFinding(t, scan(t, root), rel, 2)
			plant(t, root, rel, "package "+pkg+"\nconst nativePath = \".claude/skills\"\n")
			requireFinding(t, scan(t, root), rel, 2)
		})
	}
	var ownership struct {
		SourceHead string   `json:"source_head"`
		Inventory  []string `json:"inventory_rows"`
		Owners     []struct {
			Name                        string
			Inventory, Sources, Methods []string
			SharedOwner                 string `json:"shared_owner"`
		}
		CapabilityMethods map[string][]string `json:"capability_methods"`
		Adapters          []struct {
			ID            harnessapi.ID
			Owner, Status string
		}
	}
	readJSON(t, "source-ownership.json", &ownership)
	if ownership.SourceHead != ownershipSourceHead {
		t.Fatal("source map baseline changed")
	}
	expected := []string{"managed-catalog", "skill-catalog", "switch-catalog", "home-materialization", "seed-credentials-trust", "native-home-mcp", "shared-context-rendering", "mcp-inspection", "configuration", "installation", "project-delivery", "other-boundaries", "skill-consumers-notices"}
	sort.Strings(expected)
	sort.Strings(ownership.Inventory)
	if !reflect.DeepEqual(expected, ownership.Inventory) {
		t.Fatal("incomplete inventory")
	}
	methods := reflect.TypeOf((*harnessapi.Adapter)(nil)).Elem()
	covered := map[string]bool{}
	names := map[string]bool{}
	for _, owner := range ownership.Owners {
		if names[owner.Name] || owner.Name == "" || owner.SharedOwner == "" || len(owner.Sources) == 0 || len(owner.Methods) == 0 {
			t.Fatal("missing or duplicate surface owner")
		}
		names[owner.Name] = true
		for _, row := range owner.Inventory {
			covered[row] = true
		}
		for _, method := range owner.Methods {
			if _, ok := methods.MethodByName(method); !ok {
				t.Fatalf("unowned method %s", method)
			}
		}
		for _, path := range owner.Sources {
			if _, err := os.Stat(filepath.Join("..", "..", filepath.FromSlash(path))); err != nil {
				t.Fatal(err)
			}
		}
	}
	for _, row := range expected {
		if !covered[row] {
			t.Fatalf("unmapped inventory %s", row)
		}
	}
	adapterOwners := map[harnessapi.ID]string{}
	paths := map[string]bool{}
	for _, adapter := range ownership.Adapters {
		wantStatus := "staged"
		if adapter.ID == "claude_code" {
			wantStatus = "partial"
		}
		if adapterOwners[adapter.ID] != "" || paths[adapter.Owner] || adapter.Owner != "internal/harness/"+string(adapter.ID) || adapter.Status != wantStatus {
			t.Fatal("invalid adapter ownership")
		}
		adapterOwners[adapter.ID] = adapter.Owner
		paths[adapter.Owner] = true
	}
	if len(adapterOwners) != len(harnessregistry.IDs()) {
		t.Fatal("adapter ownership count differs")
	}
	for _, id := range harnessregistry.IDs() {
		if adapterOwners[id] == "" {
			t.Fatalf("missing owner %s", id)
		}
		adapter, err := harnessregistry.Lookup(id)
		if err != nil {
			t.Fatal(err)
		}
		capabilityNames := map[harnessapi.Capability]string{
			harnessapi.ManagedHome:          "ManagedHome",
			harnessapi.InstructionOutput:    "InstructionOutput",
			harnessapi.SkillProjection:      "SkillProjection",
			harnessapi.NativeSkillDiscovery: "NativeSkillDiscovery",
			harnessapi.MCPInspection:        "MCPInspection",
			harnessapi.NativeMCPWrite:       "NativeMCPWrite",
			harnessapi.ProjectContextOutput: "ProjectContextOutput",
			harnessapi.HomeLease:            "HomeLease",
			harnessapi.CredentialMetadata:   "CredentialMetadata",
			harnessapi.TrustObservation:     "TrustObservation",
			harnessapi.ReleaseProbing:       "ReleaseProbing",
		}
		for _, capability := range adapter.Capabilities().Values() {
			name, known := capabilityNames[capability]
			if !known || len(ownership.CapabilityMethods[name]) == 0 {
				t.Fatal("registered capability has no actual contract method")
			}
		}
	}
	// All eleven contract capability classes have a method or generic owner,
	// including those that no staged adapter is yet allowed to advertise.
	for _, name := range []string{"ManagedHome", "InstructionOutput", "SkillProjection", "NativeSkillDiscovery", "MCPInspection", "NativeMCPWrite", "ProjectContextOutput", "HomeLease", "CredentialMetadata", "TrustObservation", "ReleaseProbing"} {
		if len(ownership.CapabilityMethods[name]) == 0 {
			t.Fatalf("unmapped capability %s", name)
		}
		for _, method := range ownership.CapabilityMethods[name] {
			if _, ok := methods.MethodByName(method); !ok {
				t.Fatalf("unknown capability method %s", method)
			}
		}
	}
	TestStagedOperationsAndCapabilities(t)
}
