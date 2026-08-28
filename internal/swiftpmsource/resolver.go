package swiftpmsource

import (
	"context"
	"encoding/json"
	"reflect"
	"sort"
	"strings"
	"sync"

	"github.com/relux-works/curator/internal/artifactpolicy"
	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/closuregraph"
)

type resolvingBroker interface {
	AcquisitionBroker
	ResolvePinWithEvidence(context.Context, ManifestDependency) (Pin, GitVerificationEvidence, error)
}

// BrokeredResolver is the production no-package-code resolver. Only the Git
// broker may contact origins; every selected package tree is admitted before
// its manifest executor starts, and the result is one deterministic v3 lock.
type BrokeredResolver struct {
	Store                *closureexec.CaptureStore
	Policy               *artifactpolicy.Service
	Evaluator            ManifestEvaluator
	Broker               resolvingBroker
	Toolchain            Toolchain
	Destination          Destination
	CausalHead           string
	ProcessStartObserver func(ManifestPermit)
	mu                   sync.Mutex
	issued               map[closuregraph.ID]ResolutionResult
}

// Resolve implements LockResolver.
func (resolver *BrokeredResolver) Resolve(ctx context.Context, root string, permit ResolutionPermit, rootManifest Manifest) (ResolutionResult, error) {
	if resolver == nil || resolver.Store == nil || resolver.Policy == nil || resolver.Evaluator == nil || resolver.Broker == nil || resolver.CausalHead == "" || permit.input.Tree == nil {
		return ResolutionResult{}, fail(CodeDerivationUnauthorized, "brokered resolver authority is incomplete")
	}
	if permit.Network != "broker-only" || permit.AlgorithmID != "swiftpm-brokered-resolution-v1" || !permit.ID.Valid() || !permit.C0CheckpointID.Valid() || permit.ToolchainFingerprint != resolver.Toolchain.Git.Fingerprint {
		return ResolutionResult{}, fail(CodeDerivationUnauthorized, "brokered resolver permit is invalid")
	}
	config := Config{Store: resolver.Store, Policy: resolver.Policy, Evaluator: resolver.Evaluator, Broker: resolver.Broker, MirrorVerifier: resolverMirrorVerifier{}, Toolchain: resolver.Toolchain, Destination: resolver.Destination, CausalHead: resolver.CausalHead, ProcessStartObserver: resolver.ProcessStartObserver}
	rootFiles, rootInventory, err := inventoryTree(root)
	if err != nil {
		return ResolutionResult{}, err
	}
	rootDigest, err := manifestDigest(rootManifest)
	if err != nil {
		return ResolutionResult{}, err
	}
	rootReceiptID, err := permit.input.Receipt.ID()
	if err != nil {
		return ResolutionResult{}, err
	}
	packages := []PackageEvidence{{Identity: strings.ToLower(rootManifest.PackageName), Kind: SourcePath, Origin: "workspace:root", Manifest: rootManifest, ManifestDigest: rootDigest, SourceInventoryDigest: rootInventory, IntakeReceiptID: rootReceiptID, input: permit.input}}
	_ = rootFiles
	packageIndexes := map[string]int{packages[0].Identity: 0}
	rootArtifactID := permit.input.Receipt.ArtifactManifestID
	if err = captureLocalDependencies(ctx, config, permit.C0CheckpointID, root, packages[0], rootArtifactID, permit.input, &packages, packageIndexes); err != nil {
		return ResolutionResult{}, err
	}
	byIdentity := map[string]Pin{}
	queuedPackages := 0
	journal := []closuregraph.ID{}
	gitPermits := []closuregraph.ID{}
	gitReceipts := []closuregraph.ID{}
	for index := 1; index < len(packages); index++ {
		journal = append(journal, packages[index].IntakeReceiptID, packages[index].ManifestPermitID, packages[index].ManifestReceiptID)
	}
	for queuedPackages < len(packages) {
		pkg := packages[queuedPackages]
		queuedPackages++
		for _, dependency := range pkg.Manifest.Dependencies {
			if dependency.Kind == SourcePath {
				continue
			}
			if existing, seen := byIdentity[dependency.Identity]; seen {
				if existing.Kind != dependency.Kind || existing.CanonicalLocation != dependency.Location || !requirementMatchesPin(dependency.Requirement, existing) {
					return ResolutionResult{}, failFields(CodeResolvedFileOutOfDate, map[string]string{"identity": dependency.Identity}, "source-control constraints do not share one immutable pin")
				}
				continue
			}
			if err = recheckTool(ctx, resolver.Toolchain, resolver.Toolchain.Git); err != nil {
				return ResolutionResult{}, err
			}
			pin, pinJournal, resolveErr := resolver.Broker.ResolvePinWithEvidence(ctx, dependency)
			if resolveErr != nil {
				return ResolutionResult{}, resolveErr
			}
			journal = append(journal, pinJournal.PermitIDs...)
			journal = append(journal, pinJournal.ReceiptIDs...)
			gitPermits = append(gitPermits, pinJournal.PermitIDs...)
			gitReceipts = append(gitReceipts, pinJournal.ReceiptIDs...)
			snapshot, acquireErr := resolver.Broker.Acquire(ctx, pin)
			if acquireErr != nil {
				return ResolutionResult{}, failFields(CodeDependencyMirrorMissing, map[string]string{"identity": pin.Identity}, "brokered candidate acquisition failed: %v", acquireErr)
			}
			if err = validateSnapshot(pin, snapshot); err != nil {
				return ResolutionResult{}, err
			}
			if verifier, ok := resolver.Broker.(acquisitionEvidenceVerifier); ok {
				if err = verifier.VerifySnapshot(pin, snapshot); err != nil {
					return ResolutionResult{}, err
				}
			}
			journal = append(journal, snapshot.BrokerPermitIDs...)
			journal = append(journal, snapshot.BrokerProcessReceiptIDs...)
			gitPermits = append(gitPermits, snapshot.BrokerPermitIDs...)
			gitReceipts = append(gitReceipts, snapshot.BrokerProcessReceiptIDs...)
			input, artifactID, inventoryID, files, admitErr := admitTree(ctx, config, pin.Identity, snapshot.Root, pin.CanonicalLocation, pin.Revision, "")
			if admitErr != nil {
				return ResolutionResult{}, admitErr
			}
			evidence, evaluateErr := evaluateAdmitted(ctx, config, permit.C0CheckpointID, pin.Identity, input, files, inventoryID)
			if evaluateErr != nil {
				return ResolutionResult{}, evaluateErr
			}
			if strings.ToLower(evidence.Manifest.PackageName) != pin.Identity {
				return ResolutionResult{}, failFields(CodeDependencyPinMismatch, map[string]string{"identity": pin.Identity}, "resolved package identity differs from manifest")
			}
			evidence.Identity, evidence.Origin, evidence.Revision, evidence.GitTree, evidence.Kind, evidence.ArtifactManifestID, evidence.BrokerReceiptID = pin.Identity, pin.CanonicalLocation, pin.Revision, strings.ToLower(snapshot.GitTree), pin.Kind, artifactID, snapshot.BrokerReceiptID
			evidence.BrokerPermitIDs = append([]closuregraph.ID(nil), snapshot.BrokerPermitIDs...)
			evidence.BrokerProcessReceiptIDs = append([]closuregraph.ID(nil), snapshot.BrokerProcessReceiptIDs...)
			byIdentity[pin.Identity] = pin
			packageIndexes[pin.Identity] = len(packages)
			packages = append(packages, evidence)
			protectedRoot := mustProtectedPath(input)
			beforeLocals := len(packages)
			if err = captureLocalDependencies(ctx, config, permit.C0CheckpointID, protectedRoot, evidence, artifactID, input, &packages, packageIndexes); err != nil {
				return ResolutionResult{}, err
			}
			journal = append(journal, snapshot.BrokerReceiptID, evidence.IntakeReceiptID, evidence.ManifestPermitID, evidence.ManifestReceiptID)
			for index := beforeLocals; index < len(packages); index++ {
				journal = append(journal, packages[index].IntakeReceiptID, packages[index].ManifestPermitID, packages[index].ManifestReceiptID)
			}
		}
	}
	pins := make([]Pin, 0, len(byIdentity))
	for _, pin := range byIdentity {
		pins = append(pins, pin)
	}
	sortPins(pins)
	lockBytes, err := marshalResolvedV3(pins)
	if err != nil {
		return ResolutionResult{}, err
	}
	lock, err := ParseResolved(lockBytes)
	if err != nil {
		return ResolutionResult{}, err
	}
	journal = sortedIDs(journal)
	gitPermits, gitReceipts = sortGitEvidence(gitPermits, gitReceipts)
	receiptID, err := closuregraph.DomainID("swiftpm-brokered-resolution-receipt-v1", map[string]any{"algorithm_id": permit.AlgorithmID, "git_permit_ids": idsAny(gitPermits), "git_receipt_ids": idsAny(gitReceipts), "intake_receipt_id": string(permit.IntakeReceiptID), "journal_ids": idsAny(journal), "lock_digest": string(lock.Digest), "permit_id": string(permit.ID), "pin_count": int64(len(pins)), "toolchain_fingerprint": string(permit.ToolchainFingerprint)})
	if err != nil {
		return ResolutionResult{}, err
	}
	derivations := []closuregraph.ID{}
	for _, pkg := range packages[1:] {
		derivations = append(derivations, pkg.ManifestReceiptID)
	}
	derivations = append(derivations, gitReceipts...)
	result := ResolutionResult{Lock: lockBytes, ReceiptID: receiptID, JournalEntryIDs: journal, DerivationReceiptIDs: sortedIDs(derivations), GitPermitIDs: gitPermits, GitReceiptIDs: gitReceipts}
	resolver.mu.Lock()
	if resolver.issued == nil {
		resolver.issued = map[closuregraph.ID]ResolutionResult{}
	}
	resolver.issued[receiptID] = result
	resolver.mu.Unlock()
	return result, nil
}

func sortGitEvidence(permitIDs, receiptIDs []closuregraph.ID) ([]closuregraph.ID, []closuregraph.ID) {
	type pair struct{ permit, receipt closuregraph.ID }
	pairs := make([]pair, len(permitIDs))
	for index := range permitIDs {
		pairs[index] = pair{permit: permitIDs[index], receipt: receiptIDs[index]}
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].permit < pairs[j].permit })
	permits := make([]closuregraph.ID, len(pairs))
	receipts := make([]closuregraph.ID, len(pairs))
	for index, value := range pairs {
		permits[index], receipts[index] = value.permit, value.receipt
	}
	return permits, receipts
}

// VerifyResult proves that the generated lock and journal were issued by this
// resolver for the exact committed in-process broker algorithm permit.
func (resolver *BrokeredResolver) VerifyResult(permit ResolutionPermit, result ResolutionResult) error {
	if resolver == nil || !result.ReceiptID.Valid() {
		return fail(CodeDerivationUnauthorized, "brokered resolution receipt is absent")
	}
	resolver.mu.Lock()
	issued, ok := resolver.issued[result.ReceiptID]
	resolver.mu.Unlock()
	if !ok || !reflect.DeepEqual(issued, result) {
		return fail(CodeDerivationUnauthorized, "brokered resolution receipt was not issued by this resolver")
	}
	lock, err := ParseResolved(result.Lock)
	if err != nil {
		return err
	}
	want, err := closuregraph.DomainID("swiftpm-brokered-resolution-receipt-v1", map[string]any{"algorithm_id": permit.AlgorithmID, "git_permit_ids": idsAny(result.GitPermitIDs), "git_receipt_ids": idsAny(result.GitReceiptIDs), "intake_receipt_id": string(permit.IntakeReceiptID), "journal_ids": idsAny(result.JournalEntryIDs), "lock_digest": string(lock.Digest), "permit_id": string(permit.ID), "pin_count": int64(len(lock.Pins)), "toolchain_fingerprint": string(permit.ToolchainFingerprint)})
	if err != nil || want != result.ReceiptID {
		return fail(CodeDerivationUnauthorized, "brokered resolution receipt differs from its exact permit or inputs")
	}
	if broker, ok := resolver.Broker.(*GitBroker); !ok || broker.authority == nil || broker.authority.verify(result.GitPermitIDs, result.GitReceiptIDs) != nil {
		return fail(CodeDerivationUnauthorized, "brokered resolution Git receipts are not authority-issued")
	}
	return nil
}

type resolverMirrorVerifier struct{}

func (resolverMirrorVerifier) Verify(context.Context, string, Pin, Snapshot) (GitVerificationEvidence, error) {
	return GitVerificationEvidence{}, nil
}

type generatedResolved struct {
	Version int                    `json:"version"`
	Pins    []generatedResolvedPin `json:"pins"`
}

type generatedResolvedPin struct {
	Identity string                 `json:"identity"`
	Kind     SourceKind             `json:"kind"`
	Location string                 `json:"location"`
	State    generatedResolvedState `json:"state"`
}

type generatedResolvedState struct {
	Revision string `json:"revision"`
	Version  string `json:"version,omitempty"`
	Branch   string `json:"branch,omitempty"`
}

func marshalResolvedV3(pins []Pin) ([]byte, error) {
	records := make([]generatedResolvedPin, len(pins))
	for index, pin := range pins {
		records[index] = generatedResolvedPin{Identity: pin.Identity, Kind: pin.Kind, Location: pin.CanonicalLocation, State: generatedResolvedState{Revision: pin.Revision, Version: pin.Version, Branch: pin.Branch}}
	}
	return json.Marshal(generatedResolved{Version: 3, Pins: records})
}

func sortPins(pins []Pin) {
	for index := 1; index < len(pins); index++ {
		for cursor := index; cursor > 0 && pins[cursor].Identity < pins[cursor-1].Identity; cursor-- {
			pins[cursor], pins[cursor-1] = pins[cursor-1], pins[cursor]
		}
	}
}
