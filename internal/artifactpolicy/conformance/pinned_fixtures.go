package conformance

import (
	"crypto/sha256"
	_ "embed"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
)

// PinnedFixtureProvenance is immutable generation evidence for an accepted
// exact byte vector. The encoded fixtures are conformance evidence only: they
// are never dependency inputs, toolchain components, or publishable outputs.
type PinnedFixtureProvenance struct {
	ID            string
	Toolchain     string
	Command       string
	SourceSHA256  string
	PayloadSHA256 string
	Size          int64
}

//go:embed testdata/gnu-dynamic-pie.b64
var gnuDynamicPIEBase64 string

//go:embed testdata/gnu-static-pie.b64
var gnuStaticPIEBase64 string

//go:embed testdata/gnu-shared-object.b64
var gnuSharedObjectBase64 string

//go:embed testdata/Fixture.class.b64
var jvmFixtureBase64 string

var pinnedFixtureEvidence = []PinnedFixtureProvenance{
	{
		ID: "gnu-dynamic-pie", Toolchain: "x86_64-elf-gcc (GCC) 15.2.0; GNU ld 2.45.1",
		Command:       "/opt/homebrew/bin/x86_64-elf-gcc -x c - -nostdlib -fPIE -pie -Wl,-e,_start -Wl,--dynamic-linker,/lib64/ld-linux-x86-64.so.2 -Wl,--build-id=none -o gnu-dynamic-pie < testdata/gnu-pie.c",
		SourceSHA256:  "sha256:f2b500bdb3c726b045833ba4e7725b1d90e9698f65aaa1adb4d88921665c524f",
		PayloadSHA256: "sha256:395f0d7c9f0a63867c9794e041957a6526f9b59c0e482ea2e126e216649020ae", Size: 1808,
	},
	{
		ID: "gnu-static-pie", Toolchain: "x86_64-elf-gcc (GCC) 15.2.0; GNU ld 2.45.1",
		Command:       "/opt/homebrew/bin/x86_64-elf-gcc -x c - -nostdlib -fPIE -static-pie -Wl,-pie,--no-dynamic-linker,-e,_start,--build-id=none -o gnu-static-pie < testdata/gnu-pie.c",
		SourceSHA256:  "sha256:f2b500bdb3c726b045833ba4e7725b1d90e9698f65aaa1adb4d88921665c524f",
		PayloadSHA256: "sha256:f50542baf6de08499f4fede5cef8cc7ee9057b3bc21d7851575263e338561f33", Size: 1592,
	},
	{
		ID: "gnu-shared-object", Toolchain: "x86_64-elf-gcc (GCC) 15.2.0; GNU ld 2.45.1",
		Command:       "/opt/homebrew/bin/x86_64-elf-gcc -x c - -nostdlib -fPIC -shared -Wl,-shared,-soname,libcase.so,--build-id=none -o gnu-shared-object < testdata/gnu-shared.c",
		SourceSHA256:  "sha256:ae1ef2e7bc27e600a2d348cd0af539ef6ff295d222fb72a87876a4406b4bc103",
		PayloadSHA256: "sha256:d37f603e6c9b1143a6565558b7ad921169d22e5fc1c89542656ae8c1a37d6d5b", Size: 1560,
	},
	{
		ID: "jvm-fixture-class", Toolchain: "OpenJDK javac 26.0.1; --release 17",
		Command:       "javac --release 17 -g:none -d fixture-output testdata/Fixture.java",
		SourceSHA256:  "sha256:049d7f4013bd6cb33760a20ac5a0cdbc317deb967270e33b0d4a1bb510a5499f",
		PayloadSHA256: "sha256:d817d28c88f8507aed732737090a36baeb46ac2aca445dfe8c913db66bcd8ccf", Size: 165,
	},
}

// PinnedFixtureEvidence returns a copy of the immutable exact-vector
// provenance in stable vector order.
func PinnedFixtureEvidence() []PinnedFixtureProvenance {
	return append([]PinnedFixtureProvenance(nil), pinnedFixtureEvidence...)
}

// GNUDynamicPIE returns the accepted C01a GNU -fPIE/-pie byte vector.
func GNUDynamicPIE() []byte { return decodePinnedFixture("gnu-dynamic-pie", gnuDynamicPIEBase64) }

// GNUStaticPIE returns the accepted C01b GNU -fPIE/-static-pie byte vector.
func GNUStaticPIE() []byte { return decodePinnedFixture("gnu-static-pie", gnuStaticPIEBase64) }

// GNUSharedObject returns the accepted C01c GNU -fPIC/-shared SONAME vector.
func GNUSharedObject() []byte { return decodePinnedFixture("gnu-shared-object", gnuSharedObjectBase64) }

// JVMClass returns the accepted C07 javac-produced class byte vector.
func JVMClass() []byte { return decodePinnedFixture("jvm-fixture-class", jvmFixtureBase64) }

func decodePinnedFixture(id, encoded string) []byte {
	payload, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encoded))
	if err != nil {
		panic(fmt.Sprintf("decode pinned fixture %s: %v", id, err))
	}
	for _, evidence := range pinnedFixtureEvidence {
		if evidence.ID != id {
			continue
		}
		digest := sha256.Sum256(payload)
		identity := "sha256:" + hex.EncodeToString(digest[:])
		if int64(len(payload)) != evidence.Size || identity != evidence.PayloadSHA256 {
			panic(fmt.Sprintf("pinned fixture %s identity drift", id))
		}
		return payload
	}
	panic("unknown pinned fixture " + id)
}
