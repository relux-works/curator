package godriver

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestQualifiedGoFamilies(t *testing.T) {
	if got, want := TestedFamilies(), []string{"1.25", "1.26", "1.27"}; !reflect.DeepEqual(got, want) {
		t.Errorf("TestedFamilies = %v, want %v", got, want)
	}
	for _, version := range []string{"1.25.5", "1.26.0", "1.27.1"} {
		t.Run(version, func(t *testing.T) {
			versionLine := "go version go" + version + " " + runtime.GOOS + "/" + runtime.GOARCH
			root := stubGOROOT(t, stubScript{Version: versionLine + "\n"})
			writeTestFile(t, filepath.Join(root, "VERSION"), []byte("go"+version+"\n"), 0o644)
			session, err := Establish(context.Background(), Config{PrivateBase: t.TempDir(), GOROOT: root})
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = session.Close() })
			identity := session.Toolchain()
			if identity.GoVersion != versionLine || identity.Algorithm != buildmetaAlgorithm || !strings.HasPrefix(identity.ContentSHA256, "sha256:") {
				t.Fatalf("toolchain identity = %+v", identity)
			}
			if err := session.VerifyToolchain(context.Background()); err != nil {
				t.Fatal(err)
			}
			writeTestFile(t, filepath.Join(root, "VERSION"), []byte("go"+version+" changed\n"), 0o644)
			if err := session.VerifyToolchain(context.Background()); DiagnosticCode(err) != "toolchain_mutated" {
				t.Fatalf("mutated %s identity accepted: %v", version, err)
			}
		})
	}
}

func TestUnknownGoFamiliesRemainRefused(t *testing.T) {
	for _, version := range []string{"1.24.9", "1.28.0", "1.28rc1", "1.99.0"} {
		t.Run(version, func(t *testing.T) {
			root := stubGOROOT(t, stubScript{Version: "go version go" + version + " " + runtime.GOOS + "/" + runtime.GOARCH + "\n"})
			base := t.TempDir()
			session, err := Establish(context.Background(), Config{PrivateBase: base, GOROOT: root})
			if session != nil || DiagnosticCode(err) != "unsupported_go_family" {
				if session != nil {
					_ = session.Close()
				}
				t.Fatalf("unknown family %s admitted: %v", version, err)
			}
			entries, err := os.ReadDir(base)
			if err != nil || len(entries) != 0 {
				t.Fatalf("refused probe state remains: %v, %v", entries, err)
			}
		})
	}
}
