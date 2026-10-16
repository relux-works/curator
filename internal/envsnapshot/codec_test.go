package envsnapshot

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/protocoljson"
)

const corpus = "../crossconformance/testdata/managed-home-snapshot-v1"

func read(t *testing.T, p string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(corpus, p))
	if err != nil {
		t.Fatal(err)
	}
	return b
}
func canonical(t *testing.T, b []byte) []byte {
	t.Helper()
	var v any
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	if err := dec.Decode(&v); err != nil {
		t.Fatal(err)
	}
	out, err := protocoljson.MarshalCanonical(v)
	if err != nil {
		t.Fatal(err)
	}
	return append(out, '\n')
}
func requireReason(t *testing.T, err error, reason string) {
	t.Helper()
	var refusal *Error
	if !errors.As(err, &refusal) || refusal.Code != Mismatch || (reason != "" && refusal.Reason != reason) {
		t.Fatalf("refusal = %v, want %s/%s", err, Mismatch, reason)
	}
}
func manifest(t *testing.T) Manifest {
	t.Helper()
	m, err := DecodeManifest(read(t, "golden/manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	return m
}
func descriptor(t *testing.T) Descriptor {
	t.Helper()
	d, err := DecodeDescriptor(read(t, "golden/launch.json"))
	if err != nil {
		t.Fatal(err)
	}
	return d
}

// All indexed fixed schema cases drive the public codecs after canonical wire
// framing; their original validity expectations are independent of this code.
func TestPinnedSchemaCases(t *testing.T) {
	var index []struct {
		Schema, Instance string
		Valid            bool
	}
	if err := json.Unmarshal(read(t, "index.json"), &index); err != nil {
		t.Fatal(err)
	}
	if len(index) != 58 {
		t.Fatalf("schema inventory changed: %d", len(index))
	}
	for _, c := range index {
		t.Run(c.Instance, func(t *testing.T) {
			raw := read(t, c.Instance)
			var domain any
			decoder := json.NewDecoder(bytes.NewReader(raw))
			decoder.UseNumber()
			if err := decoder.Decode(&domain); err != nil {
				t.Fatal(err)
			}
			wire, ccjErr := protocoljson.MarshalCanonical(domain)
			if ccjErr != nil {
				if c.Valid {
					t.Fatal(ccjErr)
				}
				// Invalid numeric schema vectors (e.g. 1.5) are also outside
				// CCJ-1. Drive the codec with compact JSON and require refusal.
				var compact bytes.Buffer
				if err := json.Compact(&compact, raw); err != nil {
					t.Fatal(err)
				}
				wire = compact.Bytes()
			}
			b := append(wire, '\n')
			var err error
			if c.Schema == "managed-home-snapshot-v1.schema.json" {
				_, err = DecodeManifest(b)
			} else {
				_, err = DecodeDescriptor(b)
			}
			if (err == nil) != c.Valid {
				t.Fatalf("valid=%v error=%v", c.Valid, err)
			}
			if !c.Valid {
				requireReason(t, err, "")
			}
		})
	}
}

// Exact metadata bytes and LF-inclusive digests match published vectors;
// malformed frames, duplicates, unknown fields and wrong domains refuse.
func TestMetadataFramingAndHashDomains(t *testing.T) {
	var pins map[string]string
	if err := json.Unmarshal(read(t, "golden/digests.json"), &pins); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"manifest.json", "launch.json"} {
		t.Run(name, func(t *testing.T) {
			raw := read(t, "golden/"+name)
			var encoded []byte
			var err error
			decode := func(b []byte) error {
				if name == "manifest.json" {
					_, e := DecodeManifest(b)
					return e
				}
				_, e := DecodeDescriptor(b)
				return e
			}
			if name == "manifest.json" {
				encoded, err = EncodeManifest(manifest(t))
			} else {
				encoded, err = EncodeDescriptor(descriptor(t))
			}
			if err != nil || !bytes.Equal(encoded, raw) || Digest(raw) != pins[name] {
				t.Fatalf("pinned encoding/digest mismatch: %v", err)
			}
			for label, b := range map[string][]byte{"no LF": raw[:len(raw)-1], "two LF": append(bytes.Clone(raw), '\n'), "CRLF": append(bytes.Clone(raw[:len(raw)-1]), '\r', '\n'), "space": append([]byte(" "), raw...), "BOM": append([]byte{239, 187, 191}, raw...), "duplicate": []byte("{\"descriptor\":\"snapshot-launch-v1\",\"descriptor\":\"snapshot-launch-v1\"}\n"), "unknown": []byte(strings.Replace(string(raw), "{", "{\"_unknown\":true,", 1))} {
				t.Run(label, func(t *testing.T) {
					requireReason(t, decode(b), "")
				})
			}
			if Digest(raw) == Digest(raw[:len(raw)-1]) {
				t.Fatal("LF excluded from metadata digest")
			}
		})
	}
	fragment := []byte(`{"env":{"CODEX_HOME":"/fixture/a"},"fragment":"launch-env-fragment-v2"}`)
	got, err := FragmentDigest(fragment)
	if err != nil || got != Digest(fragment) {
		t.Fatal(got, err)
	}
	_, err = FragmentDigest(append(bytes.Clone(fragment), '\n'))
	requireReason(t, err, "fragment-canonical")
}

// Semantic corpus metadata uses public codecs and the same semantic boundary
// as Verify; tree-byte cases are exercised through Verify in TestPinnedTrees.
func TestPinnedMetadataSemantics(t *testing.T) {
	var vectors struct {
		Cases []struct{ Name, Manifest, Launch, Expected string }
	}
	if err := json.Unmarshal(read(t, "vectors.json"), &vectors); err != nil {
		t.Fatal(err)
	}
	for _, v := range vectors.Cases {
		t.Run(v.Name, func(t *testing.T) {
			m, err := DecodeManifest(read(t, v.Manifest))
			if err == nil {
				var d Descriptor
				d, err = DecodeDescriptor(read(t, v.Launch))
				if err == nil {
					_, err = validateMetadata(m, d, read(t, v.Launch))
				}
			}
			switch v.Expected {
			case "entry-set", "file-identity", "accept":
				if err != nil {
					t.Fatal(err)
				}
			default:
				requireReason(t, err, v.Expected)
			}
		})
	}
}

// Schema source pins cover the two frozen schemas; all corpus files retain
// their independent manifest hashes rather than being regenerated locally.
func TestCorpusPins(t *testing.T) {
	var pins struct {
		Files []struct{ Path, SHA256 string }
	}
	if err := json.Unmarshal(read(t, "manifest.json"), &pins); err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, p := range pins.Files {
		var b []byte
		var err error
		switch {
		case strings.HasPrefix(p.Path, "conformance/managed-home-snapshot-v1/"):
			b = read(t, strings.TrimPrefix(p.Path, "conformance/managed-home-snapshot-v1/"))
		case strings.HasPrefix(p.Path, "schemas/managed-home-snapshot-v1/"):
			b, err = schemas.ReadFile(p.Path)
			if err != nil {
				t.Fatal(err)
			}
		default:
			continue
		}
		if "sha256:"+Digest(b) != p.SHA256 {
			t.Fatalf("pin mismatch: %s", p.Path)
		}
		count++
	}
	if count != 95 {
		t.Fatalf("pin count = %d", count)
	}
}

// Encoding typed models preserves Unicode scalar values and refuses invalid
// UTF-8 instead of accepting encoding/json's replacement-rune normalization.
func TestTypedEncodingRejectsInvalidUTF8(t *testing.T) {
	m := manifest(t)
	m.Files[0].Path = "home/é.txt"
	if _, err := EncodeManifest(m); err != nil {
		t.Fatal(err)
	}
	m.Files[0].Path = "home/" + string([]byte{0xff}) + ".txt"
	_, err := EncodeManifest(m)
	requireReason(t, err, "metadata-canonical")
	d := descriptor(t)
	d.Env[string([]byte{0xff})] = "home"
	_, err = EncodeDescriptor(d)
	requireReason(t, err, "metadata-canonical")
}
