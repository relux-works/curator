package closuregraph

import (
	"fmt"

	"github.com/relux-works/curator/internal/protocoljson"
)

// SourceClosure binds the complete pre-execution chain through C5.
type SourceClosure struct {
	SchemaID       string
	C5CheckpointID ID
}

// NewSourceClosure derives the immutable closure reference from C5.
func NewSourceClosure(c5 Checkpoint) (SourceClosure, error) {
	if c5.Name != CheckpointC5 {
		return SourceClosure{}, fmt.Errorf("source closure requires a C5.plan checkpoint")
	}
	id, err := c5.ID()
	if err != nil {
		return SourceClosure{}, err
	}
	closure := SourceClosure{SchemaID: SchemaSourceClosure, C5CheckpointID: id}
	return closure, closure.Validate()
}

// Validate checks the source-closure schema.
func (closure SourceClosure) Validate() error {
	if closure.SchemaID != SchemaSourceClosure {
		return fmt.Errorf("%s: unsupported source closure schema %q", CodeGraphSchemaUnsupported, closure.SchemaID)
	}
	return validateID(closure.C5CheckpointID, "c5_checkpoint_id")
}

// CanonicalBytes returns exact curator-source-closure-v1 bytes.
func (closure SourceClosure) CanonicalBytes() ([]byte, error) { return canonicalBytes(closure) }

// ID derives closure_id.
func (closure SourceClosure) ID() (ID, error)     { return recordID(closure) }
func (closure SourceClosure) domainLabel() string { return LabelSourceClosure }
func (closure SourceClosure) canonicalValue() map[string]any {
	return map[string]any{"c5_checkpoint_id": string(closure.C5CheckpointID), "schema_id": closure.SchemaID}
}

// ExpectedCacheInput is independently derived from the immutable closure and
// expected output declarations, never from observed C6 bytes.
type ExpectedCacheInput struct {
	SchemaID              string
	ClosureID             ID
	ExpectedOutputNodeIDs []ID
}

// Validate checks expected cache input shape and canonical output order.
func (input ExpectedCacheInput) Validate() error {
	if input.SchemaID != SchemaExpectedCacheInput {
		return fmt.Errorf("%s: unsupported expected cache input schema %q", CodeGraphSchemaUnsupported, input.SchemaID)
	}
	if err := validateID(input.ClosureID, "closure_id"); err != nil {
		return err
	}
	if err := validateIDSlice(input.ExpectedOutputNodeIDs, "expected_output_node_ids", true); err != nil {
		return err
	}
	if len(input.ExpectedOutputNodeIDs) == 0 {
		return fmt.Errorf("expected_output_node_ids must not be empty")
	}
	return nil
}

// CanonicalBytes returns exact curator-expected-cache-input-v1 bytes.
func (input ExpectedCacheInput) CanonicalBytes() ([]byte, error) { return canonicalBytes(input) }

// ID derives expected_cache_input_id.
func (input ExpectedCacheInput) ID() (ID, error)     { return recordID(input) }
func (input ExpectedCacheInput) domainLabel() string { return LabelExpectedCacheInput }
func (input ExpectedCacheInput) canonicalValue() map[string]any {
	return map[string]any{"closure_id": string(input.ClosureID), "expected_output_node_ids": idsToAny(input.ExpectedOutputNodeIDs), "schema_id": input.SchemaID}
}

// ProducedArtifactObservation records C6 bytes separately from the immutable
// output node and pre-execution identities.
type ProducedArtifactObservation struct {
	Class                string
	ExpectedOutputNodeID ID
	Path                 string
	ProducerActionID     ID
	ProducesEdgeID       ID
	SHA256               ID
	Size                 int64
}

// Validate checks observation shape independent of graph references.
func (observation ProducedArtifactObservation) Validate() error {
	if err := validatePortableText(observation.Class, "observation class", false); err != nil {
		return err
	}
	if err := validateIDFields(map[string]ID{"expected_output_node_id": observation.ExpectedOutputNodeID, "producer_action_id": observation.ProducerActionID, "produces_edge_id": observation.ProducesEdgeID, "sha256": observation.SHA256}); err != nil {
		return err
	}
	if err := validatePortablePath(observation.Path, "observation path"); err != nil {
		return err
	}
	if observation.Size < 0 || observation.Size > protocoljson.MaxSafeInteger {
		return fmt.Errorf("observation size must be a non-negative safe integer")
	}
	return nil
}

// ValidateAgainst checks that an observation matches the unchanged output,
// action, and produces edge declarations in the capture table.
func (observation ProducedArtifactObservation) ValidateAgainst(records RecordTables) error {
	if err := observation.Validate(); err != nil {
		return err
	}
	nodes := map[ID]Node{}
	for _, node := range records.CaptureNodes {
		id, _ := node.ID()
		nodes[id] = node
	}
	edges := map[ID]Edge{}
	for _, edge := range records.CaptureEdges {
		id, _ := edge.ID()
		edges[id] = edge
	}
	output, ok := nodes[observation.ExpectedOutputNodeID]
	if !ok || output.Kind != NodeOutputArtifact {
		return fmt.Errorf("%s: expected output observation endpoint is missing or wrong-kind", CodeGraphReferenceInvalid)
	}
	action, ok := nodes[observation.ProducerActionID]
	if !ok || action.Kind != NodeAction {
		return fmt.Errorf("%s: producer action observation endpoint is missing or wrong-kind", CodeGraphReferenceInvalid)
	}
	edge, ok := edges[observation.ProducesEdgeID]
	if !ok || edge.Kind != EdgeProduces || edge.FromNodeID != observation.ProducerActionID || edge.ToNodeID != observation.ExpectedOutputNodeID {
		return fmt.Errorf("%s: observation produces edge does not bind the stated action and output", CodeGraphReferenceInvalid)
	}
	declared := output.Payload.(OutputArtifactPayload)
	produces := edge.Payload.(ProducesPayload)
	if observation.Path != declared.LogicalPath || observation.Path != produces.Path || observation.Class != declared.ExpectedClass || (produces.WriteClass != "" && observation.Class != produces.WriteClass) {
		return fmt.Errorf("artifact_local_output_drift: observation path/class differs from immutable declaration")
	}
	return nil
}

// CanonicalBytes returns exact produced-artifact observation bytes.
func (observation ProducedArtifactObservation) CanonicalBytes() ([]byte, error) {
	return canonicalBytes(observation)
}

// ID derives produced_artifact_observation_id.
func (observation ProducedArtifactObservation) ID() (ID, error) { return recordID(observation) }
func (observation ProducedArtifactObservation) domainLabel() string {
	return LabelProducedArtifactObservation
}
func (observation ProducedArtifactObservation) canonicalValue() map[string]any {
	return map[string]any{"class": observation.Class, "expected_output_node_id": string(observation.ExpectedOutputNodeID), "path": observation.Path, "producer_action_id": string(observation.ProducerActionID), "produces_edge_id": string(observation.ProducesEdgeID), "sha256": string(observation.SHA256), "size": observation.Size}
}

// ExecutionReceipt is the compact canonical C6 receipt projection. Detailed
// audit records are separately content-addressed by the protected executor.
type ExecutionReceipt struct {
	SchemaID               string
	ActionOrder            []ID
	ClosureID              ID
	Decision               string
	Network                string
	ProducedObservationIDs []ID
	ToolchainRechecks      string
	WriteSet               []string
}

// Validate checks the successful offline execution receipt schema.
func (receipt ExecutionReceipt) Validate() error {
	if receipt.SchemaID != SchemaExecutionReceipt {
		return fmt.Errorf("%s: unsupported execution receipt schema %q", CodeGraphSchemaUnsupported, receipt.SchemaID)
	}
	if receipt.ActionOrder == nil {
		return fmt.Errorf("action_order must be an explicit array")
	}
	if err := validateIDSlice(receipt.ActionOrder, "action_order", false); err != nil {
		return err
	}
	if err := validateID(receipt.ClosureID, "closure_id"); err != nil {
		return err
	}
	if receipt.Decision != "success" || receipt.Network != "none" || receipt.ToolchainRechecks != "match" {
		return fmt.Errorf("execution receipt must record success, network=none, and matching toolchain rechecks")
	}
	if err := validateIDSlice(receipt.ProducedObservationIDs, "produced_observation_ids", true); err != nil {
		return err
	}
	if receipt.WriteSet == nil {
		return fmt.Errorf("write_set must be an explicit array")
	}
	for index, path := range receipt.WriteSet {
		if err := validatePortablePath(path, fmt.Sprintf("write_set[%d]", index)); err != nil {
			return err
		}
	}
	return validateUniqueStrings(receipt.WriteSet, "write_set", true)
}

// CanonicalBytes returns exact curator-execution-receipt-v1 bytes.
func (receipt ExecutionReceipt) CanonicalBytes() ([]byte, error) { return canonicalBytes(receipt) }

// ID derives execution_receipt_id.
func (receipt ExecutionReceipt) ID() (ID, error)     { return recordID(receipt) }
func (receipt ExecutionReceipt) domainLabel() string { return LabelExecutionReceipt }
func (receipt ExecutionReceipt) canonicalValue() map[string]any {
	return map[string]any{"action_order": idsToAny(receipt.ActionOrder), "closure_id": string(receipt.ClosureID), "decision": receipt.Decision, "network": receipt.Network, "produced_observation_ids": idsToAny(receipt.ProducedObservationIDs), "schema_id": receipt.SchemaID, "toolchain_rechecks": receipt.ToolchainRechecks, "write_set": stringsToAny(receipt.WriteSet)}
}

// PublicationReceipt is the compact canonical C7 protected-publication
// projection.
type PublicationReceipt struct {
	SchemaID                string
	Decision                string
	ExecutionReceiptID      ID
	ExpectedCacheInputID    ID
	ProtectedResult         string
	PublishedObservationIDs []ID
}

// Validate checks the successful publication receipt schema.
func (receipt PublicationReceipt) Validate() error {
	if receipt.SchemaID != SchemaPublicationReceipt {
		return fmt.Errorf("%s: unsupported publication receipt schema %q", CodeGraphSchemaUnsupported, receipt.SchemaID)
	}
	if receipt.Decision != "published" || receipt.ProtectedResult != "exact_write" {
		return fmt.Errorf("publication receipt must record published/exact_write")
	}
	if err := validateID(receipt.ExecutionReceiptID, "execution_receipt_id"); err != nil {
		return err
	}
	if err := validateID(receipt.ExpectedCacheInputID, "expected_cache_input_id"); err != nil {
		return err
	}
	return validateIDSlice(receipt.PublishedObservationIDs, "published_observation_ids", true)
}

// CanonicalBytes returns exact curator-publication-receipt-v1 bytes.
func (receipt PublicationReceipt) CanonicalBytes() ([]byte, error) { return canonicalBytes(receipt) }

// ID derives publication_receipt_id.
func (receipt PublicationReceipt) ID() (ID, error)     { return recordID(receipt) }
func (receipt PublicationReceipt) domainLabel() string { return LabelPublicationReceipt }
func (receipt PublicationReceipt) canonicalValue() map[string]any {
	return map[string]any{"decision": receipt.Decision, "execution_receipt_id": string(receipt.ExecutionReceiptID), "expected_cache_input_id": string(receipt.ExpectedCacheInputID), "protected_result": receipt.ProtectedResult, "published_observation_ids": idsToAny(receipt.PublishedObservationIDs), "schema_id": receipt.SchemaID}
}

// DecodeSourceClosure accepts exact canonical source-closure bytes.
func DecodeSourceClosure(payload []byte) (SourceClosure, error) {
	raw, err := decodeCanonicalObject(payload, "source closure")
	if err != nil {
		return SourceClosure{}, err
	}
	if err := exactFields(raw, "source closure", []string{"c5_checkpoint_id", "schema_id"}, nil); err != nil {
		return SourceClosure{}, err
	}
	schema, err := requiredString(raw, "schema_id", "source closure")
	if err != nil {
		return SourceClosure{}, err
	}
	id, err := requiredString(raw, "c5_checkpoint_id", "source closure")
	if err != nil {
		return SourceClosure{}, err
	}
	record := SourceClosure{SchemaID: schema, C5CheckpointID: ID(id)}
	if err := record.Validate(); err != nil {
		return SourceClosure{}, err
	}
	if err := requireDecodedRecordRoundTrip(payload, record); err != nil {
		return SourceClosure{}, err
	}
	return record, nil
}

// DecodeExpectedCacheInput accepts exact canonical expected-cache bytes.
func DecodeExpectedCacheInput(payload []byte) (ExpectedCacheInput, error) {
	raw, err := decodeCanonicalObject(payload, "expected cache input")
	if err != nil {
		return ExpectedCacheInput{}, err
	}
	if err := exactFields(raw, "expected cache input", []string{"closure_id", "expected_output_node_ids", "schema_id"}, nil); err != nil {
		return ExpectedCacheInput{}, err
	}
	schema, err := requiredString(raw, "schema_id", "expected cache input")
	if err != nil {
		return ExpectedCacheInput{}, err
	}
	closure, err := requiredString(raw, "closure_id", "expected cache input")
	if err != nil {
		return ExpectedCacheInput{}, err
	}
	outputs, err := requiredIDSlice(raw, "expected_output_node_ids", "expected cache input")
	if err != nil {
		return ExpectedCacheInput{}, err
	}
	record := ExpectedCacheInput{SchemaID: schema, ClosureID: ID(closure), ExpectedOutputNodeIDs: outputs}
	if err := record.Validate(); err != nil {
		return ExpectedCacheInput{}, err
	}
	if err := requireDecodedRecordRoundTrip(payload, record); err != nil {
		return ExpectedCacheInput{}, err
	}
	return record, nil
}

// DecodeProducedArtifactObservation accepts exact canonical observation bytes.
func DecodeProducedArtifactObservation(payload []byte) (ProducedArtifactObservation, error) {
	raw, err := decodeCanonicalObject(payload, "produced artifact observation")
	if err != nil {
		return ProducedArtifactObservation{}, err
	}
	fields := []string{"class", "expected_output_node_id", "path", "producer_action_id", "produces_edge_id", "sha256", "size"}
	if err := exactFields(raw, "produced artifact observation", fields, nil); err != nil {
		return ProducedArtifactObservation{}, err
	}
	record := ProducedArtifactObservation{}
	record.Class, err = requiredString(raw, "class", "produced artifact observation")
	if err != nil {
		return ProducedArtifactObservation{}, err
	}
	output, err := requiredString(raw, "expected_output_node_id", "produced artifact observation")
	if err != nil {
		return ProducedArtifactObservation{}, err
	}
	record.ExpectedOutputNodeID = ID(output)
	record.Path, err = requiredString(raw, "path", "produced artifact observation")
	if err != nil {
		return ProducedArtifactObservation{}, err
	}
	action, err := requiredString(raw, "producer_action_id", "produced artifact observation")
	if err != nil {
		return ProducedArtifactObservation{}, err
	}
	record.ProducerActionID = ID(action)
	edge, err := requiredString(raw, "produces_edge_id", "produced artifact observation")
	if err != nil {
		return ProducedArtifactObservation{}, err
	}
	record.ProducesEdgeID = ID(edge)
	digest, err := requiredString(raw, "sha256", "produced artifact observation")
	if err != nil {
		return ProducedArtifactObservation{}, err
	}
	record.SHA256 = ID(digest)
	record.Size, err = requiredInteger(raw, "size", "produced artifact observation")
	if err != nil {
		return ProducedArtifactObservation{}, err
	}
	if err := record.Validate(); err != nil {
		return ProducedArtifactObservation{}, err
	}
	if err := requireDecodedRecordRoundTrip(payload, record); err != nil {
		return ProducedArtifactObservation{}, err
	}
	return record, nil
}

// DecodeExecutionReceipt accepts exact canonical execution-receipt bytes.
func DecodeExecutionReceipt(payload []byte) (ExecutionReceipt, error) {
	raw, err := decodeCanonicalObject(payload, "execution receipt")
	if err != nil {
		return ExecutionReceipt{}, err
	}
	fields := []string{"action_order", "closure_id", "decision", "network", "produced_observation_ids", "schema_id", "toolchain_rechecks", "write_set"}
	if err := exactFields(raw, "execution receipt", fields, nil); err != nil {
		return ExecutionReceipt{}, err
	}
	record := ExecutionReceipt{}
	record.SchemaID, err = requiredString(raw, "schema_id", "execution receipt")
	if err != nil {
		return ExecutionReceipt{}, err
	}
	record.ActionOrder, err = requiredIDSlice(raw, "action_order", "execution receipt")
	if err != nil {
		return ExecutionReceipt{}, err
	}
	closure, err := requiredString(raw, "closure_id", "execution receipt")
	if err != nil {
		return ExecutionReceipt{}, err
	}
	record.ClosureID = ID(closure)
	record.Decision, err = requiredString(raw, "decision", "execution receipt")
	if err != nil {
		return ExecutionReceipt{}, err
	}
	record.Network, err = requiredString(raw, "network", "execution receipt")
	if err != nil {
		return ExecutionReceipt{}, err
	}
	record.ProducedObservationIDs, err = requiredIDSlice(raw, "produced_observation_ids", "execution receipt")
	if err != nil {
		return ExecutionReceipt{}, err
	}
	record.ToolchainRechecks, err = requiredString(raw, "toolchain_rechecks", "execution receipt")
	if err != nil {
		return ExecutionReceipt{}, err
	}
	record.WriteSet, err = requiredStringSlice(raw, "write_set", "execution receipt")
	if err != nil {
		return ExecutionReceipt{}, err
	}
	if err := record.Validate(); err != nil {
		return ExecutionReceipt{}, err
	}
	if err := requireDecodedRecordRoundTrip(payload, record); err != nil {
		return ExecutionReceipt{}, err
	}
	return record, nil
}

// DecodePublicationReceipt accepts exact canonical publication-receipt bytes.
func DecodePublicationReceipt(payload []byte) (PublicationReceipt, error) {
	raw, err := decodeCanonicalObject(payload, "publication receipt")
	if err != nil {
		return PublicationReceipt{}, err
	}
	fields := []string{"decision", "execution_receipt_id", "expected_cache_input_id", "protected_result", "published_observation_ids", "schema_id"}
	if err := exactFields(raw, "publication receipt", fields, nil); err != nil {
		return PublicationReceipt{}, err
	}
	record := PublicationReceipt{}
	record.SchemaID, err = requiredString(raw, "schema_id", "publication receipt")
	if err != nil {
		return PublicationReceipt{}, err
	}
	record.Decision, err = requiredString(raw, "decision", "publication receipt")
	if err != nil {
		return PublicationReceipt{}, err
	}
	execution, err := requiredString(raw, "execution_receipt_id", "publication receipt")
	if err != nil {
		return PublicationReceipt{}, err
	}
	record.ExecutionReceiptID = ID(execution)
	expected, err := requiredString(raw, "expected_cache_input_id", "publication receipt")
	if err != nil {
		return PublicationReceipt{}, err
	}
	record.ExpectedCacheInputID = ID(expected)
	record.ProtectedResult, err = requiredString(raw, "protected_result", "publication receipt")
	if err != nil {
		return PublicationReceipt{}, err
	}
	record.PublishedObservationIDs, err = requiredIDSlice(raw, "published_observation_ids", "publication receipt")
	if err != nil {
		return PublicationReceipt{}, err
	}
	if err := record.Validate(); err != nil {
		return PublicationReceipt{}, err
	}
	if err := requireDecodedRecordRoundTrip(payload, record); err != nil {
		return PublicationReceipt{}, err
	}
	return record, nil
}
