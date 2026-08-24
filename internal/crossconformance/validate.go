package crossconformance

import (
	"fmt"
	"sort"
	"strings"
)

// Report is the independent verdict this package derives from the accepted
// corpus. Every counter is computed here; none is read from the corpus text or
// from the production graph package.
type Report struct {
	LabeledRecords       int
	ResolvedReferences   int
	ArtifactManifestRefs int

	CGP05CaptureReused   bool
	CGP05TargetBranches  int
	ExplicitTargetEdges  int
	CGP05DivergentKinds  []string
	CGP10ObservationSets int
	CGP10StableRecords   []string
	CGP10AllRefsResolve  bool
}

// GoldenSummary renders the two accepted oracle lines from independently
// derived counters so a reviewer can compare them with the Ruby verifier.
func (report Report) GoldenSummary() []string {
	return []string{
		fmt.Sprintf("canonical_goldens=pass labeled_records=%d cgp05_target_branches=%d cgp10_observation_branches=%d",
			report.LabeledRecords, report.CGP05TargetBranches, report.CGP10ObservationSets),
		fmt.Sprintf("canonical_references=pass cgp05_capture_reused=%t explicit_target_bindings=%d cgp10_all_refs_resolve=%t",
			report.CGP05CaptureReused, report.ExplicitTargetEdges, report.CGP10AllRefsResolve),
	}
}

// bindingNodeKinds is the closed set the accepted decision permits inside a
// selection binding. Any other kind means capture authority leaked.
var bindingNodeKinds = map[string]bool{"target_platform": true, "toolchain_component": true}

// bindingEdgeKinds is the closed set of edge kinds a binding may introduce.
var bindingEdgeKinds = map[string]bool{"targets": true, "uses_tool": true, "requires": true, "provides_interop": true}

// selectionOnlyNodeKinds must never appear inside a selection-neutral capture.
var selectionOnlyNodeKinds = map[string]bool{"target_platform": true, "toolchain_component": true}

// selectionOnlyEdgeKinds must never appear inside a selection-neutral capture.
var selectionOnlyEdgeKinds = map[string]bool{"targets": true, "uses_tool": true}

// Validate proves every claim the accepted decision makes about the corpus:
// independent hashing, complete typed reference resolution, one reused CGP05
// capture with binding-only target authority, and a CGP10 branch pair that
// changes only observation, execution, and publication identity.
func Validate(corpus Corpus) (Report, error) {
	report := Report{LabeledRecords: len(corpus.Records)}
	resolver := &referenceResolver{corpus: corpus}
	for _, record := range corpus.Records {
		if err := resolver.record(record); err != nil {
			return Report{}, err
		}
	}
	report.ResolvedReferences = resolver.resolved
	report.ArtifactManifestRefs = resolver.artifactManifests

	if err := validateCGP05(corpus, &report); err != nil {
		return Report{}, err
	}
	if err := validateCGP10(corpus, &report); err != nil {
		return Report{}, err
	}
	report.CGP10AllRefsResolve = true
	return report, nil
}

type referenceResolver struct {
	corpus            Corpus
	resolved          int
	artifactManifests int
}

func (resolver *referenceResolver) require(from, field, id, label string) error {
	record, ok := resolver.corpus.ByID(id)
	if !ok {
		return fmt.Errorf("%s.%s references unresolved fixture record %s", from, field, id)
	}
	if label != "" && record.Label != label {
		return fmt.Errorf("%s.%s resolves to label %s, want %s", from, field, record.Label, label)
	}
	resolver.resolved++
	return nil
}

func (resolver *referenceResolver) requireAll(from, field string, ids []string, label string) error {
	for _, id := range ids {
		if err := resolver.require(from, field, id, label); err != nil {
			return err
		}
	}
	return nil
}

func (resolver *referenceResolver) requireNode(from, field, id, kind string) error {
	if err := resolver.require(from, field, id, "curator-node-v1"); err != nil {
		return err
	}
	record, _ := resolver.corpus.ByID(id)
	if kind != "" && record.Value.String("kind") != kind {
		return fmt.Errorf("%s.%s resolves to node kind %q, want %q", from, field, record.Value.String("kind"), kind)
	}
	return nil
}

func (resolver *referenceResolver) record(record Record) error {
	name, value := record.Name, record.Value
	switch record.Label {
	case "curator-node-v1":
		if manifest := value.String("payload", "artifact_manifest_id"); manifest != "" {
			if published, ok := resolver.corpus.ByID(manifest); ok {
				if published.Label != "curator-artifact-manifest-v1" {
					return fmt.Errorf("%s.payload.artifact_manifest_id resolves to label %s", name, published.Label)
				}
				resolver.artifactManifests++
				resolver.resolved++
			}
		}
		return nil
	case "curator-edge-v1":
		if err := resolver.requireNode(name, "from_node_id", value.String("from_node_id"), ""); err != nil {
			return err
		}
		return resolver.requireNode(name, "to_node_id", value.String("to_node_id"), "")
	case "curator-capture-graph-v1":
		return resolver.captureGraph(record)
	case "curator-selection-context-v1":
		if err := resolver.requireAll(name, "product_node_ids", value.Strings("product_node_ids"), "curator-node-v1"); err != nil {
			return err
		}
		roles, _ := value.Member("platform_roles")
		for _, member := range roles.Members {
			if err := resolver.requireNode(name, "platform_roles."+member.Key, member.Value.Str, "target_platform"); err != nil {
				return err
			}
		}
		return nil
	case "curator-selection-binding-v1":
		return resolver.selectionBinding(record)
	case "curator-active-graph-v1":
		return resolver.activeGraph(record)
	case "curator-build-plan-v1":
		return resolver.buildPlan(record)
	case "curator-checkpoint-v1":
		return resolver.checkpoint(record)
	case "curator-source-closure-v1":
		return resolver.require(name, "c5_checkpoint_id", value.String("c5_checkpoint_id"), "curator-checkpoint-v1")
	case "curator-expected-cache-input-v1":
		if err := resolver.require(name, "closure_id", value.String("closure_id"), "curator-source-closure-v1"); err != nil {
			return err
		}
		for _, id := range value.Strings("expected_output_node_ids") {
			if err := resolver.requireNode(name, "expected_output_node_ids", id, "output_artifact"); err != nil {
				return err
			}
		}
		return nil
	case "curator-produced-artifact-observation-v1":
		if err := resolver.requireNode(name, "expected_output_node_id", value.String("expected_output_node_id"), "output_artifact"); err != nil {
			return err
		}
		if err := resolver.requireNode(name, "producer_action_id", value.String("producer_action_id"), "action"); err != nil {
			return err
		}
		return resolver.require(name, "produces_edge_id", value.String("produces_edge_id"), "curator-edge-v1")
	case "curator-execution-receipt-v1":
		if err := resolver.require(name, "closure_id", value.String("closure_id"), "curator-source-closure-v1"); err != nil {
			return err
		}
		for _, id := range value.Strings("action_order") {
			if err := resolver.requireNode(name, "action_order", id, "action"); err != nil {
				return err
			}
		}
		return resolver.requireAll(name, "produced_observation_ids", value.Strings("produced_observation_ids"), "curator-produced-artifact-observation-v1")
	case "curator-publication-receipt-v1":
		if err := resolver.require(name, "execution_receipt_id", value.String("execution_receipt_id"), "curator-execution-receipt-v1"); err != nil {
			return err
		}
		if err := resolver.require(name, "expected_cache_input_id", value.String("expected_cache_input_id"), "curator-expected-cache-input-v1"); err != nil {
			return err
		}
		return resolver.requireAll(name, "published_observation_ids", value.Strings("published_observation_ids"), "curator-produced-artifact-observation-v1")
	case "curator-artifact-manifest-v1", "curator-checkpoint-fixture-anchor-v1":
		// Byte-identical fixture anchors owned by the artifact-admission and
		// checkpoint tasks. They are independently hashed above and carry no
		// outgoing fixture reference of their own.
		return nil
	default:
		return fmt.Errorf("record %s has unknown domain label %s", name, record.Label)
	}
}

func (resolver *referenceResolver) captureGraph(record Record) error {
	name, value := record.Name, record.Value
	nodeIDs := value.Strings("node_ids")
	edgeIDs := value.Strings("edge_ids")
	if !sortedUnique(nodeIDs) {
		return fmt.Errorf("%s.node_ids is not sorted and unique", name)
	}
	if !sortedUnique(edgeIDs) {
		return fmt.Errorf("%s.edge_ids is not sorted and unique", name)
	}
	known := map[string]bool{}
	for _, id := range nodeIDs {
		if err := resolver.require(name, "node_ids", id, "curator-node-v1"); err != nil {
			return err
		}
		node, _ := resolver.corpus.ByID(id)
		if selectionOnlyNodeKinds[node.Value.String("kind")] {
			return fmt.Errorf("%s retains selection-specific node kind %q", name, node.Value.String("kind"))
		}
		known[id] = true
	}
	for _, id := range edgeIDs {
		if err := resolver.require(name, "edge_ids", id, "curator-edge-v1"); err != nil {
			return err
		}
		edge, _ := resolver.corpus.ByID(id)
		if selectionOnlyEdgeKinds[edge.Value.String("kind")] {
			return fmt.Errorf("%s retains selection-specific edge kind %q", name, edge.Value.String("kind"))
		}
		if !known[edge.Value.String("from_node_id")] || !known[edge.Value.String("to_node_id")] {
			return fmt.Errorf("%s edge %s has a dangling endpoint", name, edge.Name)
		}
	}
	for _, id := range value.Strings("root_node_ids") {
		if !known[id] {
			return fmt.Errorf("%s root %s is not a capture node", name, id)
		}
	}
	return nil
}

func (resolver *referenceResolver) selectionBinding(record Record) error {
	name, value := record.Name, record.Value
	if err := resolver.require(name, "captured_graph_id", value.String("captured_graph_id"), "curator-capture-graph-v1"); err != nil {
		return err
	}
	if err := resolver.require(name, "selection_context_id", value.String("selection_context_id"), "curator-selection-context-v1"); err != nil {
		return err
	}
	capture, _ := resolver.corpus.ByID(value.String("captured_graph_id"))
	selection, _ := resolver.corpus.ByID(value.String("selection_context_id"))
	nodeIDs := value.Strings("binding_node_ids")
	edgeIDs := value.Strings("binding_edge_ids")
	if !sortedUnique(nodeIDs) || !sortedUnique(edgeIDs) {
		return fmt.Errorf("%s binding records are not sorted and unique", name)
	}
	universe := map[string]bool{}
	for _, id := range capture.Value.Strings("node_ids") {
		universe[id] = true
	}
	for _, id := range nodeIDs {
		if err := resolver.require(name, "binding_node_ids", id, "curator-node-v1"); err != nil {
			return err
		}
		node, _ := resolver.corpus.ByID(id)
		if !bindingNodeKinds[node.Value.String("kind")] {
			return fmt.Errorf("%s binds forbidden node kind %q", name, node.Value.String("kind"))
		}
		if universe[id] {
			return fmt.Errorf("%s binding node %s replaces a capture node", name, id)
		}
		universe[id] = true
	}
	captureEdges := map[string]bool{}
	for _, id := range capture.Value.Strings("edge_ids") {
		captureEdges[id] = true
	}
	targetID := selection.Value.String("platform_roles", "target")
	explicitTargets := 0
	for _, id := range edgeIDs {
		if err := resolver.require(name, "binding_edge_ids", id, "curator-edge-v1"); err != nil {
			return err
		}
		if captureEdges[id] {
			return fmt.Errorf("%s binding edge %s replaces a capture edge", name, id)
		}
		edge, _ := resolver.corpus.ByID(id)
		kind := edge.Value.String("kind")
		if !bindingEdgeKinds[kind] {
			return fmt.Errorf("%s binds forbidden edge kind %q", name, kind)
		}
		if !universe[edge.Value.String("from_node_id")] || !universe[edge.Value.String("to_node_id")] {
			return fmt.Errorf("%s binding edge %s has a dangling endpoint", name, edge.Name)
		}
		if kind == "targets" && edge.Value.String("to_node_id") == targetID {
			explicitTargets++
		}
	}
	if targetID == "" {
		return fmt.Errorf("%s selection declares no target platform role", name)
	}
	if !containsString(nodeIDs, targetID) {
		return fmt.Errorf("%s does not bind its own target platform node", name)
	}
	if explicitTargets == 0 {
		return fmt.Errorf("%s has no explicit targets edge to its target platform", name)
	}
	return nil
}

func (resolver *referenceResolver) activeGraph(record Record) error {
	name, value := record.Name, record.Value
	if err := resolver.require(name, "captured_graph_id", value.String("captured_graph_id"), "curator-capture-graph-v1"); err != nil {
		return err
	}
	if err := resolver.require(name, "selection_binding_id", value.String("selection_binding_id"), "curator-selection-binding-v1"); err != nil {
		return err
	}
	if err := resolver.require(name, "selection_context_id", value.String("selection_context_id"), "curator-selection-context-v1"); err != nil {
		return err
	}
	binding, _ := resolver.corpus.ByID(value.String("selection_binding_id"))
	if binding.Value.String("captured_graph_id") != value.String("captured_graph_id") {
		return fmt.Errorf("%s and its binding disagree about the captured graph", name)
	}
	activations, _ := value.Member("node_activations")
	for _, item := range activations.Items {
		if err := resolver.requireNode(name, "node_activations", item.String("node_id"), ""); err != nil {
			return err
		}
	}
	edgeActivations, _ := value.Member("edge_activations")
	for _, item := range edgeActivations.Items {
		if err := resolver.require(name, "edge_activations", item.String("edge_id"), "curator-edge-v1"); err != nil {
			return err
		}
	}
	return nil
}

func (resolver *referenceResolver) buildPlan(record Record) error {
	name, value := record.Name, record.Value
	if err := resolver.require(name, "active_graph_id", value.String("active_graph_id"), "curator-active-graph-v1"); err != nil {
		return err
	}
	actions := value.Strings("action_node_ids")
	for _, id := range actions {
		if err := resolver.requireNode(name, "action_node_ids", id, "action"); err != nil {
			return err
		}
	}
	for _, id := range value.Strings("declared_output_node_ids") {
		if err := resolver.requireNode(name, "declared_output_node_ids", id, "output_artifact"); err != nil {
			return err
		}
	}
	waves, _ := value.Member("waves")
	planned := map[string]bool{}
	for _, wave := range waves.Items {
		for _, item := range wave.Items {
			if !containsString(actions, item.Str) {
				return fmt.Errorf("%s schedules action %s that the plan does not declare", name, item.Str)
			}
			if planned[item.Str] {
				return fmt.Errorf("%s schedules action %s twice", name, item.Str)
			}
			planned[item.Str] = true
		}
	}
	if len(planned) != len(actions) {
		return fmt.Errorf("%s schedules %d of %d declared actions", name, len(planned), len(actions))
	}
	return nil
}

func (resolver *referenceResolver) checkpoint(record Record) error {
	name, value := record.Name, record.Value
	if previous := value.String("previous_checkpoint_id"); previous != "" {
		if err := resolver.require(name, "previous_checkpoint_id", previous, ""); err != nil {
			return err
		}
		predecessor, _ := resolver.corpus.ByID(previous)
		if predecessor.Label != "curator-checkpoint-v1" && predecessor.Label != "curator-checkpoint-fixture-anchor-v1" {
			return fmt.Errorf("%s chains to non-checkpoint label %s", name, predecessor.Label)
		}
	}
	payload, _ := value.Member("payload")
	fields := map[string]string{
		"active_graph_id":      "curator-active-graph-v1",
		"captured_graph_id":    "curator-capture-graph-v1",
		"selection_binding_id": "curator-selection-binding-v1",
		"selection_context_id": "curator-selection-context-v1",
		"build_plan_id":        "curator-build-plan-v1",
	}
	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if id := payload.String(key); id != "" {
			if err := resolver.require(name, "payload."+key, id, fields[key]); err != nil {
				return err
			}
		}
	}
	if value.String("checkpoint_name") == "C5.plan" {
		for _, member := range payload.Members {
			if member.Key != "build_plan_id" {
				return fmt.Errorf("%s adds %q to C5; C5 may add no graph record", name, member.Key)
			}
		}
	}
	return nil
}

func validateCGP05(corpus Corpus, report *Report) error {
	capture, ok := corpus.ByName("cgp05.capture")
	if !ok {
		return fmt.Errorf("CGP05 capture record is absent")
	}
	branches := []string{"darwin", "linux"}
	targets := map[string]string{}
	for _, branch := range branches {
		binding, ok := corpus.ByName("cgp05.binding." + branch)
		if !ok {
			return fmt.Errorf("CGP05 %s binding is absent", branch)
		}
		if binding.Value.String("captured_graph_id") != capture.Derived {
			return fmt.Errorf("CGP05 %s binding does not reuse the exact capture", branch)
		}
		platform, ok := corpus.ByName("cgp05.platform." + branch)
		if !ok {
			return fmt.Errorf("CGP05 %s platform node is absent", branch)
		}
		if platform.Value.String("kind") != "target_platform" {
			return fmt.Errorf("CGP05 %s platform node has kind %q", branch, platform.Value.String("kind"))
		}
		targets[branch] = platform.Derived
		edge, ok := corpus.ByName("cgp05.targets." + branch)
		if !ok {
			return fmt.Errorf("CGP05 %s targets edge is absent", branch)
		}
		if edge.Value.String("kind") != "targets" || edge.Value.String("to_node_id") != platform.Derived {
			return fmt.Errorf("CGP05 %s targets edge does not bind its platform", branch)
		}
		if !containsString(binding.Value.Strings("binding_edge_ids"), edge.Derived) {
			return fmt.Errorf("CGP05 %s binding omits its explicit targets edge", branch)
		}
		report.ExplicitTargetEdges++
	}
	if targets["darwin"] == targets["linux"] {
		return fmt.Errorf("CGP05 branches share one platform node; the two targets are not distinct")
	}
	divergent := []string{"platform", "selection", "targets", "binding", "active", "plan", "c4", "c5"}
	for _, kind := range divergent {
		first, firstOK := corpus.ByName("cgp05." + kind + ".darwin")
		second, secondOK := corpus.ByName("cgp05." + kind + ".linux")
		if !firstOK || !secondOK {
			return fmt.Errorf("CGP05 %s branch pair is incomplete", kind)
		}
		if first.Derived == second.Derived {
			return fmt.Errorf("CGP05 %s identities are equal across targets", kind)
		}
	}
	report.CGP05CaptureReused = true
	report.CGP05TargetBranches = len(branches)
	report.CGP05DivergentKinds = divergent
	return nil
}

func validateCGP10(corpus Corpus, report *Report) error {
	stable := []string{
		"action", "output", "declares", "reads", "produces", "publishes", "capture", "platform",
		"toolchain", "selection", "uses-tool", "targets.product", "targets.action", "targets.toolchain",
		"targets.output", "binding", "active", "plan", "c4", "c5", "closure", "expected-cache-input",
	}
	for _, name := range stable {
		if _, ok := corpus.ByName("cgp10." + name); !ok {
			return fmt.Errorf("CGP10 stable record %s is absent", name)
		}
	}
	report.CGP10StableRecords = stable

	action, _ := corpus.ByName("cgp10.action")
	slots := map[string]map[string]int{"reads": {}, "produces": {}, "uses_tool": {}}
	slotField := map[string]string{"reads": "read_slot", "produces": "write_slot", "uses_tool": "tool_slot"}
	for _, record := range corpus.Records {
		if record.Label != "curator-edge-v1" || record.Value.String("from_node_id") != action.Derived {
			continue
		}
		kind := record.Value.String("kind")
		field, tracked := slotField[kind]
		if !tracked {
			continue
		}
		slots[kind][record.Value.String("payload", field)]++
	}
	for kind, want := range map[string]string{"reads": "src", "produces": "cli", "uses_tool": "compiler"} {
		if slots[kind][want] != 1 {
			return fmt.Errorf("CGP10 %s slot %q is bound %d times, want exactly once", kind, want, slots[kind][want])
		}
		if len(slots[kind]) != 1 {
			return fmt.Errorf("CGP10 %s declares %d slots, want one", kind, len(slots[kind]))
		}
	}

	closure, _ := corpus.ByName("cgp10.closure")
	cacheInput, _ := corpus.ByName("cgp10.expected-cache-input")
	produces, _ := corpus.ByName("cgp10.produces")
	output, _ := corpus.ByName("cgp10.output")
	branches := []string{"one", "two"}
	for _, branch := range branches {
		observation, ok := corpus.ByName("cgp10.observation." + branch)
		if !ok {
			return fmt.Errorf("CGP10 %s observation is absent", branch)
		}
		if observation.Value.String("expected_output_node_id") != output.Derived ||
			observation.Value.String("produces_edge_id") != produces.Derived ||
			observation.Value.String("producer_action_id") != action.Derived {
			return fmt.Errorf("CGP10 %s observation does not reuse the stable action, edge, and expected output", branch)
		}
		execution, ok := corpus.ByName("cgp10.execution." + branch)
		if !ok {
			return fmt.Errorf("CGP10 %s execution receipt is absent", branch)
		}
		if execution.Value.String("closure_id") != closure.Derived {
			return fmt.Errorf("CGP10 %s execution receipt leaves the stable closure", branch)
		}
		observed := execution.Value.Strings("produced_observation_ids")
		if len(observed) != 1 || observed[0] != observation.Derived {
			return fmt.Errorf("CGP10 %s execution receipt does not carry exactly its own observation", branch)
		}
		publication, ok := corpus.ByName("cgp10.publication." + branch)
		if !ok {
			return fmt.Errorf("CGP10 %s publication receipt is absent", branch)
		}
		if publication.Value.String("execution_receipt_id") != execution.Derived ||
			publication.Value.String("expected_cache_input_id") != cacheInput.Derived {
			return fmt.Errorf("CGP10 %s publication receipt does not chain execution and expected cache input", branch)
		}
		published := publication.Value.Strings("published_observation_ids")
		if len(published) != 1 || published[0] != observation.Derived {
			return fmt.Errorf("CGP10 %s publication receipt publishes another observation", branch)
		}
	}
	for _, kind := range []string{"observation", "execution", "publication"} {
		first, _ := corpus.ByName("cgp10." + kind + ".one")
		second, _ := corpus.ByName("cgp10." + kind + ".two")
		if first.Derived == second.Derived {
			return fmt.Errorf("CGP10 %s branches share one identity", kind)
		}
	}
	one, _ := corpus.ByName("cgp10.observation.one")
	two, _ := corpus.ByName("cgp10.observation.two")
	if one.Value.String("sha256") == two.Value.String("sha256") {
		return fmt.Errorf("CGP10 branches observe the same output bytes; the pair proves nothing")
	}
	report.CGP10ObservationSets = len(branches)
	return nil
}

func sortedUnique(values []string) bool {
	for index := 1; index < len(values); index++ {
		if values[index-1] >= values[index] {
			return false
		}
	}
	return true
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

// DescribeCorpus renders a stable, human-readable inventory of the corpus for
// evidence artifacts.
func DescribeCorpus(corpus Corpus) string {
	var builder strings.Builder
	for _, name := range corpus.Names() {
		record, _ := corpus.ByName(name)
		builder.WriteString(name)
		builder.WriteString(" ")
		builder.WriteString(record.Label)
		builder.WriteString(" ")
		builder.WriteString(record.Derived)
		builder.WriteString("\n")
	}
	return builder.String()
}
