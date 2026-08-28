// Command curator-spec-pin verifies Curator's immutable curator-spec release.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/relux-works/curator/internal/buildrepo"
)

func main() {
	root := flag.String("root", "", "path to the curator-spec repository root")
	revision := flag.String("revision", "", "full curator-spec commit checked out at root")
	flag.Parse()
	if *root == "" || *revision == "" || flag.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "usage: curator-spec-pin --root <curator-spec-root> --revision <full-commit>")
		os.Exit(2)
	}
	if err := buildrepo.VerifyReleasePin(*root, *revision); err != nil {
		fmt.Fprintf(os.Stderr, "curator-spec-pin: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("curator-spec-pin: verified %s at %s\n", buildrepo.SpecReleaseTag, buildrepo.SpecReleaseCommit)
}
