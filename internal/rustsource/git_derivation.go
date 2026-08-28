package rustsource

import (
	"encoding/base64"
	"sort"

	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/protocoljson"
)

const gitDerivationSchemaID = "rust-git-projection-v1"

type gitDerivationPayload struct {
	SchemaID                 string              `json:"schema_id"`
	Mode                     ProjectionMode      `json:"mode"`
	Selected                 []string            `json:"selected"`
	NormalizerInputs         []string            `json:"normalizer_inputs"`
	NormalizerID             string              `json:"normalizer_id"`
	NormalizedManifestBase64 string              `json:"normalized_manifest_base64"`
	Commit                   string              `json:"commit"`
	Tree                     string              `json:"tree"`
	PackagePath              string              `json:"package_path"`
	Include                  []string            `json:"include"`
	ManifestTracked          bool                `json:"manifest_tracked"`
	Submodules               []SubmoduleEvidence `json:"submodules"`
}

// BindGitDerivation accepts only canonical evidence causally issued by the
// protected derivation executor and bound to its exact output bytes.
func bindGitDerivation(executor *closureexec.Executor, receipt closureexec.DerivationReceipt, payload []byte) (gitDerivation, error) {
	if executor == nil {
		return gitDerivation{}, fail(CodeGitIdentityInvalid, "Git derivation executor is absent", nil)
	}
	if err := executor.VerifyIssuedDerivationReceipt(receipt); err != nil {
		return gitDerivation{}, err
	}
	if receipt.InvocationSubtype != closureexec.DerivationManifest || len(receipt.Outputs) != 1 || receipt.Outputs[0].Path != "rust-git-projection-v1.json" || receipt.Outputs[0].SchemaID != gitDerivationSchemaID || string(receipt.Outputs[0].SHA256) != "sha256:"+digest(payload) || receipt.Outputs[0].Size != int64(len(payload)) {
		return gitDerivation{}, fail(CodeGitIdentityInvalid, "Git derivation receipt does not bind projection bytes", nil)
	}
	var decoded gitDerivationPayload
	if err := protocoljson.UnmarshalCanonical(payload, &decoded); err != nil {
		return gitDerivation{}, fail(CodeGitIdentityInvalid, "Git derivation payload is not canonical", nil)
	}
	if decoded.SchemaID != gitDerivationSchemaID || decoded.NormalizerID != NormalizerID {
		return gitDerivation{}, fail(CodeVendorTransformUnsupported, "Git derivation implementation is unsupported", nil)
	}
	manifest, err := base64.StdEncoding.Strict().DecodeString(decoded.NormalizedManifestBase64)
	if err != nil || len(manifest) == 0 {
		return gitDerivation{}, fail(CodeGitIdentityInvalid, "normalized Git manifest bytes are invalid", nil)
	}
	selected, unique := sortedUnique(decoded.Selected)
	if !unique || len(selected) == 0 {
		return gitDerivation{}, fail(CodeGitIdentityInvalid, "Git selected set is empty or duplicated", nil)
	}
	inputs, inputUnique := sortedUnique(decoded.NormalizerInputs)
	if !inputUnique || len(inputs) == 0 {
		return gitDerivation{}, fail(CodeGitIdentityInvalid, "Git normalizer inputs are empty or duplicated", nil)
	}
	receiptID, err := receipt.ID()
	if err != nil {
		return gitDerivation{}, err
	}
	include, includeUnique := sortedUnique(decoded.Include)
	if !includeUnique {
		return gitDerivation{}, fail(CodeGitIdentityInvalid, "Git include declarations are duplicated", nil)
	}
	submodules, submoduleErr := canonicalSubmodules(decoded.Submodules)
	if submoduleErr != nil {
		return gitDerivation{}, submoduleErr
	}
	derivation := gitDerivation{
		mode:               decoded.Mode,
		selected:           selected,
		normalizerInputs:   inputs,
		normalizerID:       decoded.NormalizerID,
		normalizedManifest: manifest,
		receiptID:          string(receiptID),
		commit:             decoded.Commit,
		tree:               decoded.Tree,
		packagePath:        decoded.PackagePath,
		include:            include,
		manifestTracked:    decoded.ManifestTracked,
		submodules:         submodules,
	}
	derivation.seal, err = gitDerivationSeal(derivation)
	if err != nil {
		return gitDerivation{}, fail(CodeGitIdentityInvalid, "Git derivation seal cannot be computed", nil)
	}
	return derivation, nil
}

func gitDerivationSeal(derivation gitDerivation) (string, error) {
	payload, err := protocoljson.MarshalCanonical(map[string]any{
		"commit":                     derivation.commit,
		"include":                    derivation.include,
		"manifest_tracked":           derivation.manifestTracked,
		"mode":                       string(derivation.mode),
		"normalized_manifest_base64": base64.StdEncoding.EncodeToString(derivation.normalizedManifest),
		"normalizer_id":              derivation.normalizerID,
		"normalizer_inputs":          derivation.normalizerInputs,
		"package_path":               derivation.packagePath,
		"receipt_id":                 derivation.receiptID,
		"selected":                   derivation.selected,
		"submodules":                 submoduleValues(derivation.submodules),
		"tree":                       derivation.tree,
	})
	if err != nil {
		return "", err
	}
	return digest(payload), nil
}

func canonicalSubmodules(values []SubmoduleEvidence) ([]SubmoduleEvidence, error) {
	result := append([]SubmoduleEvidence(nil), values...)
	sort.Slice(result, func(i, j int) bool { return result[i].Path < result[j].Path })
	for index, value := range result {
		if !safeRelative(value.Path) || !validLowerHex(value.Gitlink, 40) || value.Gitlink != value.Commit || !validLowerHex(value.Commit, 40) || !validLowerHex(value.TreeDigest, 40) || (index > 0 && result[index-1].Path == value.Path) {
			return nil, fail(CodeGitIdentityInvalid, "submodule derivation evidence is invalid", map[string]string{"path": value.Path})
		}
	}
	return result, nil
}

func submoduleValues(values []SubmoduleEvidence) []any {
	result := make([]any, len(values))
	for index, value := range values {
		result[index] = map[string]any{"commit": value.Commit, "gitlink": value.Gitlink, "path": value.Path, "tree_digest": value.TreeDigest}
	}
	return result
}
