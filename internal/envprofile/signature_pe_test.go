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
// MSYS and Cygwin executables on every test host. Unknown and malformed images
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
	for _, value := range []any{file, optional, section} {
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
