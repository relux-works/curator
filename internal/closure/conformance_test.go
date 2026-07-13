package closure

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
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
