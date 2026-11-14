// Command check-store-compatibility checks the exact historical store replay.
// Its successful result is a migration property, not v1 store conformance.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/relux-works/curator/internal/storecompatibility"
)

func main() {
	input := flag.String("input", "", "complete go test -json historical replay")
	dir := flag.String("dir", "", "same-head CI artifact directory")
	commit := flag.String("commit", "", "expected exact source head")
	flag.Parse()
	if *dir != "" {
		if *input != "" || flag.NArg() != 0 {
			fmt.Fprintln(os.Stderr, "input and artifact modes are exclusive")
			os.Exit(2)
		}
		report, err := storecompatibility.Admit(*dir, *commit)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := json.NewEncoder(os.Stdout).Encode(report); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if *input == "" || flag.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "usage: check-store-compatibility -input replay.jsonl")
		os.Exit(2)
	}
	f, err := os.Open(*input)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer func() { _ = f.Close() }()
	report, err := storecompatibility.Check(f)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := json.NewEncoder(os.Stdout).Encode(report); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
