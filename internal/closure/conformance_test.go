package closure

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/relux-works/curator/internal/skillspec"
)

func TestClosureOrderingConformanceVector(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	payload, err := os.ReadFile(filepath.Join(root, "vectors", "closures.json"))
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Name     string     `json:"name"`
		Nodes    []string   `json:"nodes"`
		Edges    [][]string `json:"edges"`
		Expected []string   `json:"expected_provider_order"`
	}
	if err := json.Unmarshal(payload, &cases); err != nil {
		t.Fatal(err)
	}
	for _, testCase := range cases {
		if testCase.Name != "deterministic-diamond" {
			continue
		}
		nodes := map[string]*Node{}
		for _, name := range testCase.Nodes {
			nodes[name] = &Node{Name: name}
		}
		for _, edge := range testCase.Edges {
			consumer, provider := edge[0], edge[1]
			nodes[provider].Edges = append(nodes[provider].Edges, Edge{Consumer: consumer})
		}
		ordered, err := topologicalOrder(nodes)
		if err != nil {
			t.Fatal(err)
		}
		got := make([]string, len(ordered))
		for index, node := range ordered {
			got[index] = node.Name
		}
		if !reflect.DeepEqual(got, testCase.Expected) {
			t.Fatalf("provider order = %v, want %v", got, testCase.Expected)
		}
	}
}

func TestBuildCommandOrderingConformanceVector(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	payload, err := os.ReadFile(filepath.Join(root, "vectors", "manager-lifecycle.json"))
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		BuildOrderCases []struct {
			Name                string              `json:"name"`
			ActiveBuildCommands map[string][]string `json:"active_build_commands"`
			ClosureEdges        []struct {
				Consumer string `json:"consumer"`
				Provider string `json:"provider"`
			} `json:"closure_edges"`
			ExpectedProviderOrder []string `json:"expected_provider_order"`
			ExpectedBuildOrder    []string `json:"expected_build_order"`
		} `json:"build_order_cases"`
	}
	if err := json.Unmarshal(payload, &document); err != nil {
		t.Fatal(err)
	}
	for _, testCase := range document.BuildOrderCases {
		if testCase.Name != "provider-first-and-lexical-command-order" {
			continue
		}
		nodes := map[string]*Node{}
		for name, commands := range testCase.ActiveBuildCommands {
			exported := map[string]skillspec.Command{}
			for _, command := range commands {
				exported[command] = skillspec.Command{Name: command, Type: "build"}
			}
			nodes[name] = &Node{
				Name:  name,
				Spec:  &skillspec.Spec{Commands: exported},
				Edges: []Edge{{Consumer: ProjectEdge, Mode: "full"}},
			}
		}
		for _, edge := range testCase.ClosureEdges {
			nodes[edge.Provider].Edges = append(nodes[edge.Provider].Edges, Edge{Consumer: edge.Consumer, Mode: "full"})
		}
		ordered, err := topologicalOrder(nodes)
		if err != nil {
			t.Fatal(err)
		}
		providerOrder := make([]string, 0, len(ordered))
		var buildOrder []string
		for _, node := range ordered {
			providerOrder = append(providerOrder, node.Name)
			for _, command := range node.ActiveCommandNames() {
				buildOrder = append(buildOrder, node.Name+"/"+command)
			}
		}
		if !reflect.DeepEqual(providerOrder, testCase.ExpectedProviderOrder) {
			t.Fatalf("provider order = %v, want %v", providerOrder, testCase.ExpectedProviderOrder)
		}
		if !reflect.DeepEqual(buildOrder, testCase.ExpectedBuildOrder) {
			t.Fatalf("build order = %v, want %v", buildOrder, testCase.ExpectedBuildOrder)
		}
		return
	}
	t.Fatal("provider-first-and-lexical-command-order conformance case missing")
}
