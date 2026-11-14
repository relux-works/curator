package envprofile

import (
	"bytes"
	"debug/pe"
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Actual PE imports, rather than injected library names, distinguish native,
// MSYS and Cygwin executables on every test host. Unknown-DLL and malformed-header/import fixtures
// cannot acquire a known runtime. These fixtures do not execute Windows code.
func testSignaturePEFlavours(t *testing.T) {
	t.Helper()
	for _, tc := range []struct{ name, library, want string }{
		{"pe-native", "KERNEL32.dll", "native"},
		{"pe-msys", "msys-2.0.dll", "msys"},
		{"pe-cygwin", "cygwin1.dll", "cygwin"},
		{"pe-unknown", "fixture.dll", "unknown"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := signaturePEFixture(t, tc.library)
			got, err := signatureExecutableFlavour(path)
			if err != nil || got != tc.want {
				t.Fatalf("PE runtime=%q err=%v want=%q", got, err, tc.want)
			}
		})
	}
	t.Run("pe-malformed-imports", func(t *testing.T) {
		path := signaturePEFixture(t, "KERNEL32.dll")
		image, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		// The import descriptor points beyond the actual section data.
		binary.LittleEndian.PutUint32(image[512:], 0x1fff)
		if err := os.WriteFile(path, image, 0700); err != nil {
			t.Fatal(err)
		}
		if _, err := signatureExecutableFlavour(path); err == nil || !strings.Contains(err.Error(), "gpg-path-flavour-unknown:") {
			t.Fatalf("malformed PE imports did not produce named refusal: %v", err)
		}
	})
	t.Run("pe-malformed", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "invalid.exe")
		if err := os.WriteFile(path, []byte("not a PE executable"), 0700); err != nil {
			t.Fatal(err)
		}
		if _, err := signatureExecutableFlavour(path); err == nil || !strings.Contains(err.Error(), "gpg-path-flavour-unknown:") {
			t.Fatalf("malformed PE did not produce named refusal: %v", err)
		}
	})
}

func signaturePEFixture(t *testing.T, library string) string {
	t.Helper()
	return signaturePEFixtureWidth(t, library, false)
}

func signaturePEFixtureWidth(t *testing.T, library string, is32 bool) string {
	t.Helper()
	var header bytes.Buffer
	dos := make([]byte, 64)
	copy(dos, "MZ")
	binary.LittleEndian.PutUint32(dos[60:], 64)
	header.Write(dos)
	header.WriteString("PE\x00\x00")
	file := pe.FileHeader{Machine: pe.IMAGE_FILE_MACHINE_AMD64, NumberOfSections: 1, SizeOfOptionalHeader: 240}
	optional := pe.OptionalHeader64{Magic: 0x20b, NumberOfRvaAndSizes: 16}
	optional.DataDirectory[pe.IMAGE_DIRECTORY_ENTRY_IMPORT] = pe.DataDirectory{VirtualAddress: 0x1000, Size: 40}
	section := pe.SectionHeader32{VirtualSize: 512, VirtualAddress: 0x1000, SizeOfRawData: 512, PointerToRawData: 512}
	copy(section.Name[:], ".idata")
	values := []any{file, optional, section}
	if is32 {
		file.Machine, file.SizeOfOptionalHeader = pe.IMAGE_FILE_MACHINE_I386, 224
		header32 := pe.OptionalHeader32{Magic: 0x10b, NumberOfRvaAndSizes: 16}
		header32.DataDirectory[pe.IMAGE_DIRECTORY_ENTRY_IMPORT] = optional.DataDirectory[pe.IMAGE_DIRECTORY_ENTRY_IMPORT]
		values = []any{file, header32, section}
	}
	for _, value := range values {
		if err := binary.Write(&header, binary.LittleEndian, value); err != nil {
			t.Fatal(err)
		}
	}
	image := make([]byte, 1024)
	copy(image, header.Bytes())
	imports := image[512:]
	binary.LittleEndian.PutUint32(imports[0:], 0x1080)  // OriginalFirstThunk
	binary.LittleEndian.PutUint32(imports[12:], 0x1060) // DLL name
	binary.LittleEndian.PutUint32(imports[16:], 0x1080) // FirstThunk
	copy(imports[0x60:], library+"\x00")
	binary.LittleEndian.PutUint64(imports[0x80:], 0x10a0)
	copy(imports[0xa2:], "fixture_import\x00") // Hint word precedes the name.
	path := filepath.Join(t.TempDir(), "fixture.exe")
	if err := os.WriteFile(path, image, 0700); err != nil {
		t.Fatal(err)
	}
	return path
}

// Malformed name RVAs/terminators and IATs fail closed, with the original
// native image beside each plant. Ordinal-only images remain unqualified;
// mixed named/ordinal imports retain the ordinal DLL's runtime identity.
func TestSignaturePEImportValidation(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func([]byte)
	}{
		{"symbol-rva", func(b []byte) { binary.LittleEndian.PutUint64(b[0x280:], 0x3000) }},
		{"symbol-reserved-bits", func(b []byte) {
			// Keep all RVAs mapped above bit 31; only the named import's
			// reserved bit violates PE32+, rather than an incidental EOF.
			binary.LittleEndian.PutUint32(b[64+4+20+112+pe.IMAGE_DIRECTORY_ENTRY_IMPORT*8:], 0x80001000)
			binary.LittleEndian.PutUint32(b[64+4+20+240+12:], 0x80001000)
			binary.LittleEndian.PutUint32(b[0x200:], 0x80001080)
			binary.LittleEndian.PutUint32(b[0x20c:], 0x80001060)
			binary.LittleEndian.PutUint32(b[0x210:], 0x80001080)
			binary.LittleEndian.PutUint64(b[0x280:], 0x800010a0)
		}},
		{"symbol-unterminated", func(b []byte) {
			for i := 0x2a2; i < len(b); i++ {
				b[i] = 'x'
			}
		}},
		{"dll-rva", func(b []byte) { binary.LittleEndian.PutUint32(b[0x20c:], 0x3000) }},
		{"dll-unterminated", func(b []byte) {
			for i := 0x260; i < len(b); i++ {
				b[i] = 'x'
			}
		}},
		{"iat-rva", func(b []byte) { binary.LittleEndian.PutUint32(b[0x210:], 0x3000) }},
		{"ordinal-only", func(b []byte) { binary.LittleEndian.PutUint64(b[0x280:], uint64(1)<<63|1) }},
		{"ordinal-reserved-bits", func(b []byte) { binary.LittleEndian.PutUint64(b[0x280:], uint64(1)<<63|1<<20|1) }},
		{"lookup-unterminated", func(b []byte) {
			binary.LittleEndian.PutUint32(b[0x200:], 0x11f8)
			binary.LittleEndian.PutUint32(b[0x210:], 0x11f8)
			binary.LittleEndian.PutUint64(b[0x3f8:], 0x10a0)
		}},
		{"directory-unterminated", func(b []byte) { binary.LittleEndian.PutUint32(b[64+4+20+112+pe.IMAGE_DIRECTORY_ENTRY_IMPORT*8+4:], 20) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := signaturePEFixture(t, "KERNEL32.dll")
			if got, err := signatureExecutableFlavour(path); err != nil || got != "native" {
				t.Fatalf("native PE positive control: %q %v", got, err)
			}
			image, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			tc.change(image)
			if err := os.WriteFile(path, image, 0700); err != nil {
				t.Fatal(err)
			}
			if got, err := signatureExecutableFlavour(path); err == nil || !strings.Contains(err.Error(), "gpg-path-flavour-unknown:") {
				t.Fatalf("unsafe PE %s admitted as %q: %v", tc.name, got, err)
			}
		})
	}
	for _, dll := range []string{"msys-2.0.dll", "cygwin1.dll"} {
		t.Run("mixed-ordinal-"+dll, func(t *testing.T) {
			path := signaturePEFixture(t, "KERNEL32.dll")
			image, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			binary.LittleEndian.PutUint32(image[64+4+20+112+pe.IMAGE_DIRECTORY_ENTRY_IMPORT*8+4:], 60)
			binary.LittleEndian.PutUint32(image[0x214:], 0x10f0)
			binary.LittleEndian.PutUint32(image[0x220:], 0x10d0)
			binary.LittleEndian.PutUint32(image[0x224:], 0x10f0)
			copy(image[0x2d0:], dll+"\x00")
			binary.LittleEndian.PutUint64(image[0x2f0:], uint64(1)<<63|1)
			if err := os.WriteFile(path, image, 0700); err != nil {
				t.Fatal(err)
			}
			want := "msys"
			if dll == "cygwin1.dll" {
				want = "cygwin"
			}
			if got, err := signatureExecutableFlavour(path); err != nil || got != want {
				t.Fatalf("ordinal runtime lost: %q %v want %s", got, err, want)
			}
		})
	}
}

// PE32 uses 4-byte thunks, including ordinal flag bit 31. Valid native imports
// pass and invalid names/ordinal-only inventories fail as on PE32+.
func TestSignaturePE32Imports(t *testing.T) {
	for _, plant := range []string{"none", "name-rva", "ordinal-only"} {
		t.Run(plant, func(t *testing.T) {
			path := signaturePEFixtureWidth(t, "KERNEL32.dll", true)
			if got, err := signatureExecutableFlavour(path); err != nil || got != "native" {
				t.Fatalf("PE32 positive control: %q %v", got, err)
			}
			image, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			switch plant {
			case "name-rva":
				binary.LittleEndian.PutUint32(image[0x280:], 0x3000)
			case "ordinal-only":
				binary.LittleEndian.PutUint32(image[0x280:], uint32(1)<<31|1)
			}
			if err := os.WriteFile(path, image, 0700); err != nil {
				t.Fatal(err)
			}
			got, err := signatureExecutableFlavour(path)
			if plant == "none" {
				if err != nil || got != "native" {
					t.Fatalf("valid PE32: %q %v", got, err)
				}
			} else if err == nil || !strings.Contains(err.Error(), "gpg-path-flavour-unknown:") {
				t.Fatalf("unsafe PE32 %s admitted: %q %v", plant, got, err)
			}
		})
	}
}
