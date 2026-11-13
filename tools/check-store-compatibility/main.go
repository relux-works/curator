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
	flag.Parse()
	if *input == "" || flag.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "usage: check-store-compatibility -input replay.jsonl")
		os.Exit(2)
	}
	f, err := os.Open(*input)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer f.Close()
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
