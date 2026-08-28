package buildcache

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/buildmeta"
	"github.com/relux-works/curator/internal/protocoljson"
)

// cacheRejectionVector is the authoritative rejection record for the cache
// boundary. Portable expected values stay in the suite.
type cacheRejectionVector struct {
	Name     string `json:"name"`
	Boundary string `json:"boundary"`
	Expected struct {
		Result           string `json:"result"`
		Error            string `json:"error"`
		Reuse            bool   `json:"reuse"`
		ArtifactExecuted bool   `json:"artifact_executed"`
	} `json:"expected"`
}

func loadCacheRejections(t *testing.T) map[string]cacheRejectionVector {
	t.Helper()
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	payload, err := os.ReadFile(filepath.Join(root, "vectors", "build-drivers.json")) // #nosec G304 -- explicit conformance input
	if os.IsNotExist(err) {
		t.Skipf("%s publishes no build-drivers vector", root)
	}
	if err != nil {
		t.Fatal(err)
	}
	var vectors struct {
		RejectionCases []cacheRejectionVector `json:"rejection_cases"`
	}
	if err := json.Unmarshal(payload, &vectors); err != nil {
		t.Fatal(err)
	}
	indexed := make(map[string]cacheRejectionVector, len(vectors.RejectionCases))
	for _, testCase := range vectors.RejectionCases {
		indexed[testCase.Name] = testCase
	}
	return indexed
}

// cacheRejection is the Curator half of one cache-boundary mapping.
type cacheRejection struct {
	// status is the stable non-reusable Curator inspection outcome.
	status Status
	// reason is a stable fragment of the Curator diagnostic for this seam.
	reason string
	// mutate corrupts the published entry. Nil means the entry stays intact
	// and expect carries the divergence instead.
	mutate func(t *testing.T, entry entryPaths, publication Publication)
	// expectHash overrides the recorded receipt identity a caller requires.
	expectHash func(buildmeta.ReceiptHash) buildmeta.ReceiptHash
	// publishOnly runs the case against Publish instead of Inspect.
	publishOnly func(t *testing.T, store *Store, publication Publication) error
	// note records why the Curator seam is the published boundary when the
	// Curator vocabulary is coarser than the protocol code.
	note string
}

// entryPaths names the members of one published cache entry.
type entryPaths struct {
	entry    string
	receipt  string
	artifact string
}

// rewriteReceiptField stores a canonical receipt whose stored bytes carry one
// tampered value. The bytes are produced directly with the protocol canonical
// encoder so the divergence survives Curator's own receipt constructor, which
// would otherwise refuse to build an inconsistent document.
func rewriteReceiptField(t *testing.T, entry entryPaths, publication Publication, edit func(map[string]any)) {
	t.Helper()
	decoder := json.NewDecoder(bytes.NewReader(publication.ReceiptBytes))
	decoder.UseNumber()
	var document map[string]any
	if err := decoder.Decode(&document); err != nil {
		t.Fatal(err)
	}
	edit(document)
	payload, err := protocoljson.MarshalCanonical(document)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, entry.receipt, payload, 0o600)
}

// receiptInput returns the mutable input object inside a decoded receipt.
func receiptInput(t *testing.T, document map[string]any) map[string]any {
	t.Helper()
	input, ok := document["input"].(map[string]any)
	if !ok {
		t.Fatalf("stored receipt has no input object: %v", document)
	}
	return input
}

func cacheRejectionMappings() map[string]cacheRejection {
	forgedKey := buildmeta.CacheKey("sha256:" + strings.Repeat("0", 64))
	return map[string]cacheRejection{
		"cache-key-mismatch": {
			status: Corrupt, reason: "receipt cache_key does not match its complete input",
			mutate: func(t *testing.T, entry entryPaths, publication Publication) {
				rewriteReceiptField(t, entry, publication, func(document map[string]any) {
					document["cache_key"] = string(forgedKey)
				})
			},
		},
		"cache-wrong-target": {
			status: Corrupt, reason: "invalid receipt",
			mutate: func(t *testing.T, entry entryPaths, publication Publication) {
				rewriteReceiptField(t, entry, publication, func(document map[string]any) {
					target, ok := receiptInput(t, document)["target"].(map[string]any)
					if !ok {
						t.Fatalf("stored receipt has no target object")
					}
					target["goos"] = "otheros"
				})
			},
		},
		"cache-wrong-toolchain": {
			status: Corrupt, reason: "invalid receipt",
			mutate: func(t *testing.T, entry entryPaths, publication Publication) {
				rewriteReceiptField(t, entry, publication, func(document map[string]any) {
					toolchain, ok := receiptInput(t, document)["toolchain"].(map[string]any)
					if !ok {
						t.Fatalf("stored receipt has no toolchain object")
					}
					toolchain["content_sha256"] = "sha256:" + strings.Repeat("d", 64)
				})
			},
		},
		"cache-wrong-policy": {
			status: Corrupt, reason: "invalid receipt",
			mutate: func(t *testing.T, entry entryPaths, publication Publication) {
				rewriteReceiptField(t, entry, publication, func(document map[string]any) {
					policy, ok := receiptInput(t, document)["policy"].(map[string]any)
					if !ok {
						t.Fatalf("stored receipt has no policy object")
					}
					policy["execution_policy"] = buildmeta.ReservedHardenedExecutionPolicy
				})
			},
		},
		"cache-wrong-build-source": {
			status: Corrupt, reason: "invalid receipt",
			mutate: func(t *testing.T, entry entryPaths, publication Publication) {
				rewriteReceiptField(t, entry, publication, func(document map[string]any) {
					source, ok := receiptInput(t, document)["build_source"].(map[string]any)
					if !ok {
						t.Fatalf("stored receipt has no build_source object")
					}
					source["content_sha256"] = "sha256:" + strings.Repeat("e", 64)
				})
			},
		},
		"receipt-hash-mismatch": {
			status: Corrupt, reason: "receipt hash mismatch",
			expectHash: func(buildmeta.ReceiptHash) buildmeta.ReceiptHash {
				return buildmeta.ReceiptHash("sha256:" + strings.Repeat("0", 64))
			},
		},
		"artifact-hash-mismatch": {
			status: Corrupt, reason: "artifact hash mismatch",
			mutate: func(t *testing.T, entry entryPaths, _ Publication) {
				writeFile(t, entry.artifact, []byte("TAMPERED"), 0o700)
			},
		},
		"artifact-size-mismatch": {
			status: Corrupt, reason: "artifact size mismatch",
			mutate: func(t *testing.T, entry entryPaths, _ Publication) {
				writeFile(t, entry.artifact, []byte("artifact-with-extra-bytes"), 0o700)
			},
		},
		"artifact-path-mismatch": {
			status: Corrupt, reason: "",
			note: "Curator addresses the artifact exclusively by its manager-derived " +
				"relative path, so a differently named member is an inexact entry",
			mutate: func(t *testing.T, entry entryPaths, _ Publication) {
				if err := os.Rename(entry.artifact, filepath.Join(filepath.Dir(entry.artifact), "package-chosen")); err != nil {
					t.Fatal(err)
				}
			},
		},
		"noncanonical-receipt-whitespace": {
			status: Corrupt, reason: "invalid receipt",
			mutate: func(t *testing.T, entry entryPaths, publication Publication) {
				var object map[string]any
				if err := json.Unmarshal(publication.ReceiptBytes, &object); err != nil {
					t.Fatal(err)
				}
				pretty, err := json.MarshalIndent(object, "", "  ")
				if err != nil {
					t.Fatal(err)
				}
				writeFile(t, entry.receipt, pretty, 0o600)
			},
		},
		"noncanonical-receipt-trailing-lf": {
			status: Corrupt, reason: "invalid receipt",
			mutate: func(t *testing.T, entry entryPaths, publication Publication) {
				writeFile(t, entry.receipt, append(append([]byte(nil), publication.ReceiptBytes...), '\n'), 0o600)
			},
		},
		"partial-cache-entry": {
			status: Corrupt, reason: "incomplete",
			mutate: func(t *testing.T, entry entryPaths, _ Publication) {
				if err := os.Remove(entry.receipt); err != nil {
					t.Fatal(err)
				}
			},
		},
		"artifact-link": {
			status: UntrustedProvenance, reason: "",
			note: "a linked or non-regular artifact never reaches the byte comparison",
			mutate: func(t *testing.T, entry entryPaths, _ Publication) {
				target := entry.artifact + ".real"
				if err := os.Rename(entry.artifact, target); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(target, entry.artifact); err != nil {
					t.Skipf("this host cannot create the symbolic link the vector needs: %v", err)
				}
			},
		},
		"artifact-special-file": {
			status: UntrustedProvenance, reason: "",
			note: "a linked or non-regular artifact never reaches the byte comparison",
			mutate: func(t *testing.T, entry entryPaths, _ Publication) {
				if err := os.Remove(entry.artifact); err != nil {
					t.Fatal(err)
				}
				makeCacheSpecialFile(t, entry.artifact)
			},
		},
		"concurrent-publisher-different-bytes": {
			status: Corrupt, reason: "",
			note: "a second publisher of the same key with other bytes is a publication conflict",
			publishOnly: func(t *testing.T, store *Store, publication Publication) error {
				other, _ := testPublication(t, store.Home(), publication.Input, []byte("other-winner-bytes"))
				_, err := store.Publish(other, testHomeLock{})
				return err
			},
		},
	}
}

// TestCacheRejectionClustersMapToStableCuratorOutcomes proves every
// authoritative cache-boundary rejection reaches a stable non-reusable Curator
// outcome and that the cached artifact is never handed back for execution.
func TestCacheRejectionClustersMapToStableCuratorOutcomes(t *testing.T) {
	published := loadCacheRejections(t)
	owned := cacheRejectionMappings()

	names := make([]string, 0, len(published))
	for name, vector := range published {
		// The forged-provenance vector is a boundary case with its own
		// protected-state assertions; it is owned by the untrusted-cache tests.
		if vector.Boundary == "cache" && name != "self-consistent-forged-receipt-outside-protected-state" {
			names = append(names, name)
		}
	}
	sort.Strings(names)

	for _, name := range names {
		vector := published[name]
		mapping, ok := owned[name]
		if !ok {
			t.Errorf("authoritative cache rejection %q has no Curator mapping", name)
			continue
		}
		t.Run(name, func(t *testing.T) {
			if vector.Expected.Result != "reject" || vector.Expected.Reuse || vector.Expected.ArtifactExecuted {
				t.Fatalf("vector %q no longer fails closed: %+v", name, vector.Expected)
			}
			store := newTestStore(t)
			publication, receiptHash := testPublication(t, store.Home(), testInput("tool"), []byte("artifact"))
			publishedResult, err := store.Publish(publication, testHomeLock{})
			if err != nil {
				t.Fatal(err)
			}
			hit := store.Inspect(Expectation{Input: publication.Input, ReceiptHash: receiptHash, Assurance: publication.Assurance})
			if hit.Status != Hit {
				t.Fatalf("published entry did not become a hit: %+v", hit)
			}
			entry := entryPaths{
				entry:    filepath.Dir(filepath.Dir(hit.ArtifactPath)),
				receipt:  filepath.Join(filepath.Dir(hit.ArtifactPath), "..", ReceiptFilename),
				artifact: hit.ArtifactPath,
			}

			if mapping.publishOnly != nil {
				err := mapping.publishOnly(t, store, publication)
				if err == nil {
					t.Fatalf("%s was accepted, want the %s rejection", name, vector.Expected.Error)
				}
				var conflict *ConflictError
				if !errors.As(err, &conflict) {
					t.Fatalf("%s produced %v, want a stable publication conflict", name, err)
				}
				if conflict.Key != publishedResult.CacheKey {
					t.Fatalf("%s conflict names %q", name, conflict.Key)
				}
				return
			}

			if mapping.mutate != nil {
				mapping.mutate(t, entry, publication)
			}
			expected := Expectation{Input: publication.Input, ReceiptHash: receiptHash, Assurance: publication.Assurance}
			if mapping.expectHash != nil {
				expected.ReceiptHash = mapping.expectHash(receiptHash)
			}
			result := store.Inspect(expected)
			if result.Status != mapping.status {
				t.Fatalf("%s status = %q (%s), want %q", name, result.Status, result.Reason, mapping.status)
			}
			if result.DryRunOutcome() == "cache-hit" {
				t.Fatalf("%s remained reusable", name)
			}
			if result.ArtifactPath != "" {
				t.Fatalf("%s returned an artifact path for execution: %s", name, result.ArtifactPath)
			}
			if result.Reason == "" {
				t.Fatalf("%s carries no stable reason", name)
			}
			if mapping.reason != "" && !strings.Contains(result.Reason, mapping.reason) {
				t.Fatalf("%s reason = %q, want it to name %q", name, result.Reason, mapping.reason)
			}
		})
	}

	for name := range owned {
		if _, ok := published[name]; !ok {
			t.Errorf("Curator maps %q, which the authoritative suite no longer publishes", name)
		}
	}
}
