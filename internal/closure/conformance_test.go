package closure

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/conformancecoverage"
	"github.com/relux-works/curator/internal/skillspec"
)

type closureOrderCase struct {
	Name     string     `json:"name"`
	Nodes    []string   `json:"nodes"`
	Edges    [][]string `json:"edges"`
	Expected []string   `json:"expected_provider_order"`
	Error    string     `json:"error"`
}

type closureBuildOrderEdge struct {
	Consumer string `json:"consumer"`
	Provider string `json:"provider"`
}

type closureBuildOrderCase struct {
	Name                  string                  `json:"name"`
	ActiveBuildCommands   map[string][]string     `json:"active_build_commands"`
	ClosureEdges          []closureBuildOrderEdge `json:"closure_edges"`
	ExpectedProviderOrder []string                `json:"expected_provider_order"`
	ExpectedBuildOrder    []string                `json:"expected_build_order"`
}

func TestClosureOrderingConformanceVector(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	payload, err := os.ReadFile(filepath.Join(root, "vectors", "closures.json"))
	if err != nil {
		t.Fatal(err)
	}
	var cases []closureOrderCase
	if err := json.Unmarshal(payload, &cases); err != nil {
		t.Fatal(err)
	}
	conformancecoverage.RunOutcomes(t, "closure/ordering-cases", cases,
		func(tc closureOrderCase) string { return tc.Name }, func(t *testing.T, testCase closureOrderCase) conformancecoverage.Observation {
			if testCase.Name == "commit-conflict" {
				return conformancecoverage.Observation{BoundReason: "commit conflict is resolved before the topological ordering entry point"}
			}
			nodes := map[string]*Node{}
			for _, name := range testCase.Nodes {
				nodes[name] = &Node{Name: name}
			}
			for _, edge := range testCase.Edges {
				if len(edge) != 2 {
					t.Fatalf("closure vector edge has %d members, want consumer and provider", len(edge))
				}
				consumer, provider := edge[0], edge[1]
				if nodes[consumer] == nil {
					nodes[consumer] = &Node{Name: consumer}
				}
				if nodes[provider] == nil {
					nodes[provider] = &Node{Name: provider}
				}
				nodes[provider].Edges = append(nodes[provider].Edges, Edge{Consumer: consumer})
			}
			ordered, err := topologicalOrder(nodes)
			if testCase.Error != "" {
				if err == nil || !strings.Contains(err.Error(), "dependency cycle") {
					t.Fatalf("topologicalOrder error = %v, want %s", err, testCase.Error)
				}
				return conformancecoverage.Observation{}
			}
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
			return conformancecoverage.Observation{}
		})
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
		BuildOrderCases []closureBuildOrderCase `json:"build_order_cases"`
	}
	if err := json.Unmarshal(payload, &document); err != nil {
		t.Fatal(err)
	}
	conformancecoverage.Run(t, "manager-lifecycle/build-order-cases", document.BuildOrderCases,
		func(tc closureBuildOrderCase) string { return tc.Name }, func(t *testing.T, testCase closureBuildOrderCase) {
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
		})
}
