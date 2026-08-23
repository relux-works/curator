package artifactpolicy

import (
	"fmt"
	"math"
)

type limitAccountant struct {
	limits                  LimitVector
	rawBytes                int64
	emittedBytes            int64
	containerCount          int64
	entryCount              int64
	maxArchiveDepth         int64
	maxLeafBytes            int64
	maxStreamInputBytes     int64
	maxStreamEmittedBytes   int64
	maxStreamExpansionRatio int64
}

type limitFailure struct {
	name     string
	limit    int64
	observed int64
}

type streamOutputBudget struct {
	maximum         int64
	limitName       string
	limit           int64
	compressed      int64
	startingEmitted int64
	rawBytes        int64
}

func (failure *limitFailure) Error() string {
	return fmt.Sprintf("%s exceeded: observed %d, limit %d", failure.name, failure.observed, failure.limit)
}

func newLimitAccountant(limits LimitVector, rawBytes int64) (*limitAccountant, error) {
	if rawBytes < 0 {
		return nil, fmt.Errorf("raw payload size is negative")
	}
	if rawBytes > limits.MaxRawPayloadBytes {
		return nil, &limitFailure{name: "max_raw_payload_bytes", limit: limits.MaxRawPayloadBytes, observed: rawBytes}
	}
	return &limitAccountant{limits: limits, rawBytes: rawBytes}, nil
}

func (accountant *limitAccountant) addRaw(amount int64) error {
	observed, ok := checkedAdd(accountant.rawBytes, amount)
	if !ok {
		return &limitFailure{name: "max_raw_payload_bytes", limit: accountant.limits.MaxRawPayloadBytes, observed: math.MaxInt64}
	}
	accountant.rawBytes = observed
	if observed > accountant.limits.MaxRawPayloadBytes {
		return &limitFailure{name: "max_raw_payload_bytes", limit: accountant.limits.MaxRawPayloadBytes, observed: observed}
	}
	return nil
}

func (accountant *limitAccountant) addContainer(depth int64) error {
	if depth > accountant.maxArchiveDepth {
		accountant.maxArchiveDepth = depth
	}
	count, ok := checkedAdd(accountant.containerCount, 1)
	if !ok || count > accountant.limits.MaxContainerCount {
		if !ok {
			count = math.MaxInt64
		}
		accountant.containerCount = count
		return &limitFailure{name: "max_container_count", limit: accountant.limits.MaxContainerCount, observed: count}
	}
	accountant.containerCount = count
	if depth > accountant.limits.MaxArchiveDepth {
		return &limitFailure{name: "max_archive_depth", limit: accountant.limits.MaxArchiveDepth, observed: depth}
	}
	return nil
}

func (accountant *limitAccountant) addEntry(amount int64) error {
	count, ok := checkedAdd(accountant.entryCount, amount)
	if !ok || count > accountant.limits.MaxEntryCount {
		if !ok {
			count = math.MaxInt64
		}
		accountant.entryCount = count
		return &limitFailure{name: "max_entry_count", limit: accountant.limits.MaxEntryCount, observed: count}
	}
	accountant.entryCount = count
	return nil
}

func (accountant *limitAccountant) checkLeaf(size int64) error {
	if size > accountant.maxLeafBytes {
		accountant.maxLeafBytes = size
	}
	return accountant.preflightLeaf(size)
}

// preflightLeaf rejects a format-declared leaf before its payload is read or
// allocated. It does not record observed work; checkLeaf records the size only
// after the payload is read, except that callers may use checkLeaf on the
// rejecting branch to retain the declared over-limit observation.
func (accountant *limitAccountant) preflightLeaf(size int64) error {
	if size < 0 || size > accountant.limits.MaxSingleLeafBytes {
		return &limitFailure{name: "max_single_leaf_bytes", limit: accountant.limits.MaxSingleLeafBytes, observed: size}
	}
	return nil
}

func (accountant *limitAccountant) addEmitted(emitted, compressed int64) error {
	if emitted < 0 || compressed < 0 {
		return fmt.Errorf("negative stream accounting")
	}
	streamRatio := int64(0)
	if emitted > 0 && compressed > 0 {
		streamRatio = ceilingRatio(emitted, compressed)
	}
	if streamRatio > accountant.maxStreamExpansionRatio ||
		(streamRatio == accountant.maxStreamExpansionRatio && emitted > accountant.maxStreamEmittedBytes) {
		accountant.maxStreamExpansionRatio = streamRatio
		accountant.maxStreamEmittedBytes = emitted
		accountant.maxStreamInputBytes = compressed
	}
	total, ok := checkedAdd(accountant.emittedBytes, emitted)
	if !ok || total > accountant.limits.MaxTotalEmittedBytes {
		if !ok {
			total = math.MaxInt64
		}
		accountant.emittedBytes = total
		return &limitFailure{name: "max_total_emitted_bytes", limit: accountant.limits.MaxTotalEmittedBytes, observed: total}
	}
	accountant.emittedBytes = total
	if emitted > 0 && compressed == 0 {
		return &limitFailure{name: "max_expansion_ratio", limit: accountant.limits.MaxExpansionRatio, observed: math.MaxInt64}
	}
	if compressed > 0 && ratioExceeded(emitted, compressed, accountant.limits.MaxExpansionRatio) {
		return &limitFailure{name: "max_expansion_ratio", limit: accountant.limits.MaxExpansionRatio, observed: streamRatio}
	}
	if accountant.rawBytes == 0 {
		if total > 0 {
			return &limitFailure{name: "max_expansion_ratio", limit: accountant.limits.MaxExpansionRatio, observed: math.MaxInt64}
		}
		return nil
	}
	if ratioExceeded(total, accountant.rawBytes, accountant.limits.MaxExpansionRatio) {
		return &limitFailure{name: "max_expansion_ratio", limit: accountant.limits.MaxExpansionRatio, observed: ceilingRatio(total, accountant.rawBytes)}
	}
	return nil
}

// preflightEmitted rejects a format-declared expansion before the inspector
// allocates or reads the member. It deliberately does not charge emitted
// bytes: the accounting record distinguishes declared early refusal from work
// actually performed. A successful materialization is charged by addEmitted.
func (accountant *limitAccountant) preflightEmitted(emitted, compressed int64) error {
	if emitted < 0 || compressed < 0 {
		return fmt.Errorf("negative stream accounting")
	}
	total, ok := checkedAdd(accountant.emittedBytes, emitted)
	if !ok || total > accountant.limits.MaxTotalEmittedBytes {
		if !ok {
			total = math.MaxInt64
		}
		return &limitFailure{name: "max_total_emitted_bytes", limit: accountant.limits.MaxTotalEmittedBytes, observed: total}
	}
	if emitted > 0 && compressed == 0 {
		return &limitFailure{name: "max_expansion_ratio", limit: accountant.limits.MaxExpansionRatio, observed: math.MaxInt64}
	}
	if compressed > 0 && ratioExceeded(emitted, compressed, accountant.limits.MaxExpansionRatio) {
		return &limitFailure{
			name: "max_expansion_ratio", limit: accountant.limits.MaxExpansionRatio,
			observed: ceilingRatio(emitted, compressed),
		}
	}
	if accountant.rawBytes == 0 {
		if total > 0 {
			return &limitFailure{name: "max_expansion_ratio", limit: accountant.limits.MaxExpansionRatio, observed: math.MaxInt64}
		}
		return nil
	}
	if ratioExceeded(total, accountant.rawBytes, accountant.limits.MaxExpansionRatio) {
		return &limitFailure{
			name: "max_expansion_ratio", limit: accountant.limits.MaxExpansionRatio,
			observed: ceilingRatio(total, accountant.rawBytes),
		}
	}
	return nil
}

func (accountant *limitAccountant) streamBudget(compressed int64) (streamOutputBudget, error) {
	if compressed < 0 {
		return streamOutputBudget{}, fmt.Errorf("negative compressed stream size")
	}
	if accountant.emittedBytes < 0 || accountant.emittedBytes > accountant.limits.MaxTotalEmittedBytes {
		return streamOutputBudget{}, fmt.Errorf("invalid existing emitted-byte accounting")
	}
	candidates := []streamOutputBudget{
		{
			maximum:   accountant.limits.MaxSingleLeafBytes,
			limitName: "max_single_leaf_bytes", limit: accountant.limits.MaxSingleLeafBytes,
		},
		{
			maximum:   accountant.limits.MaxTotalEmittedBytes - accountant.emittedBytes,
			limitName: "max_total_emitted_bytes", limit: accountant.limits.MaxTotalEmittedBytes,
		},
	}
	streamRatioBudget, ok := checkedMultiply(compressed, accountant.limits.MaxExpansionRatio)
	if !ok {
		streamRatioBudget = math.MaxInt64
	}
	candidates = append(candidates, streamOutputBudget{
		maximum: streamRatioBudget, limitName: "max_expansion_ratio",
		limit: accountant.limits.MaxExpansionRatio,
	})
	aggregateRatioBudget, ok := checkedMultiply(accountant.rawBytes, accountant.limits.MaxExpansionRatio)
	if !ok {
		aggregateRatioBudget = math.MaxInt64
	}
	if aggregateRatioBudget < accountant.emittedBytes {
		aggregateRatioBudget = 0
	} else {
		aggregateRatioBudget -= accountant.emittedBytes
	}
	candidates = append(candidates, streamOutputBudget{
		maximum: aggregateRatioBudget, limitName: "max_expansion_ratio",
		limit: accountant.limits.MaxExpansionRatio,
	})
	selected := candidates[0]
	for _, candidate := range candidates[1:] {
		if candidate.maximum < selected.maximum {
			selected = candidate
		}
	}
	selected.compressed = compressed
	selected.startingEmitted = accountant.emittedBytes
	selected.rawBytes = accountant.rawBytes
	return selected, nil
}

func (budget streamOutputBudget) failure(observedOutput int64) *limitFailure {
	observed := observedOutput
	switch budget.limitName {
	case "max_total_emitted_bytes":
		if total, ok := checkedAdd(budget.startingEmitted, observedOutput); ok {
			observed = total
		} else {
			observed = math.MaxInt64
		}
	case "max_expansion_ratio":
		streamRatio := ceilingRatio(observedOutput, budget.compressed)
		total, ok := checkedAdd(budget.startingEmitted, observedOutput)
		if !ok {
			total = math.MaxInt64
		}
		aggregateRatio := ceilingRatio(total, budget.rawBytes)
		observed = streamRatio
		if aggregateRatio > observed {
			observed = aggregateRatio
		}
	}
	return &limitFailure{name: budget.limitName, limit: budget.limit, observed: observed}
}

func (accountant *limitAccountant) snapshot() TraversalAccounting {
	aggregateRatio := ceilingRatio(accountant.emittedBytes, accountant.rawBytes)
	if accountant.emittedBytes == 0 {
		aggregateRatio = 0
	}
	return TraversalAccounting{
		RawPayloadBytes:         accountant.rawBytes,
		TotalEmittedBytes:       accountant.emittedBytes,
		ContainerCount:          accountant.containerCount,
		EntryCount:              accountant.entryCount,
		MaxObservedArchiveDepth: accountant.maxArchiveDepth,
		MaxObservedLeafBytes:    accountant.maxLeafBytes,
		MaxStreamInputBytes:     accountant.maxStreamInputBytes,
		MaxStreamEmittedBytes:   accountant.maxStreamEmittedBytes,
		MaxStreamExpansionRatio: accountant.maxStreamExpansionRatio,
		AggregateExpansionRatio: aggregateRatio,
	}
}

func bindTraversalAccounting(
	accounting TraversalAccounting,
	rawKind string,
	nodes []ManifestNode,
) (TraversalAccounting, error) {
	manifestedEntries := int64(0)
	if rawKind == "canonical_tree" && len(nodes) > 0 {
		// A captured tree charges its real root as an enumerated filesystem
		// entry. A file payload's root is instead covered by raw-payload
		// accounting and is not an archive entry.
		manifestedEntries = 1
	}
	manifestedEmitted := int64(0)
	manifestedLeafMaximum := int64(0)
	for _, node := range nodes {
		emittedSize, emittedErr := manifestedNodeEmittedSize(node)
		if emittedErr != nil {
			return TraversalAccounting{}, emittedErr
		}
		if node.Parent != "" {
			var ok bool
			manifestedEntries, ok = checkedAdd(manifestedEntries, 1)
			if !ok {
				return TraversalAccounting{}, fmt.Errorf("manifested entry accounting overflows")
			}
		}
		if node.Parent == "" || len(node.ContainerChain) == 0 || !byteBearingNode(node.Kind) || node.SHA256 == "" {
			if node.Parent == "" && node.Kind != NodeRegularFile {
				continue
			}
		} else {
			var ok bool
			manifestedEmitted, ok = checkedAdd(manifestedEmitted, emittedSize)
			if !ok {
				return TraversalAccounting{}, fmt.Errorf("manifested emitted-byte accounting overflows")
			}
		}
		if byteBearingNode(node.Kind) && node.SHA256 != "" &&
			(node.Parent != "" || node.Kind == NodeRegularFile) && emittedSize > manifestedLeafMaximum {
			manifestedLeafMaximum = emittedSize
		}
	}
	if accounting.EntryCount < manifestedEntries {
		return TraversalAccounting{}, fmt.Errorf(
			"entry accounting %d is smaller than manifested entries %d",
			accounting.EntryCount, manifestedEntries,
		)
	}
	if accounting.TotalEmittedBytes < manifestedEmitted {
		return TraversalAccounting{}, fmt.Errorf(
			"emitted-byte accounting %d is smaller than manifested bytes %d",
			accounting.TotalEmittedBytes, manifestedEmitted,
		)
	}
	if accounting.MaxObservedLeafBytes < manifestedLeafMaximum {
		return TraversalAccounting{}, fmt.Errorf(
			"leaf accounting %d is smaller than manifested maximum %d",
			accounting.MaxObservedLeafBytes, manifestedLeafMaximum,
		)
	}
	accounting.ManifestedEntryCount = manifestedEntries
	accounting.UnmanifestedEntryCount = accounting.EntryCount - manifestedEntries
	accounting.ManifestedEmittedBytes = manifestedEmitted
	accounting.UnmanifestedEmittedBytes = accounting.TotalEmittedBytes - manifestedEmitted
	accounting.MaxManifestedLeafBytes = manifestedLeafMaximum
	if accounting.MaxObservedLeafBytes > manifestedLeafMaximum {
		accounting.MaxUnmanifestedLeafBytes = accounting.MaxObservedLeafBytes
	} else {
		accounting.MaxUnmanifestedLeafBytes = 0
	}
	return accounting, nil
}

func byteBearingNode(kind NodeKind) bool {
	return kind == NodeRegularFile || kind == NodeArchive || kind == NodeCompressedStream
}

func ratioExceeded(numerator, denominator, limit int64) bool {
	if denominator <= 0 {
		return numerator > 0
	}
	if numerator <= 0 {
		return false
	}
	if denominator > math.MaxInt64/limit {
		return false
	}
	return numerator > denominator*limit
}

func ceilingRatio(numerator, denominator int64) int64 {
	if denominator <= 0 {
		return math.MaxInt64
	}
	quotient := numerator / denominator
	if numerator%denominator != 0 && quotient < math.MaxInt64 {
		quotient++
	}
	return quotient
}

func checkedAdd(left, right int64) (int64, bool) {
	if left < 0 || right < 0 || left > math.MaxInt64-right {
		return 0, false
	}
	return left + right, true
}

func checkedMultiply(left, right int64) (int64, bool) {
	if left < 0 || right < 0 || left != 0 && right > math.MaxInt64/left {
		return 0, false
	}
	return left * right, true
}
