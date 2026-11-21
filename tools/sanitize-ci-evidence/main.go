// Package main provides the command-line entry point for sanitizing CI evidence.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/relux-works/curator/internal/ciartifact"
)

func main() {
	source := flag.String("source", "", "source evidence directory")
	destination := flag.String("destination", "", "new sanitized evidence directory")
	flag.Parse()
	if *source == "" || *destination == "" || flag.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "source and destination directories are required")
		os.Exit(2)
	}
	counts, err := ciartifact.SanitizeTree(*source, *destination)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not create sanitized evidence copy: %v\n", err)
		os.Exit(1)
	}
	encoded, err := json.Marshal(counts)
	if err != nil {
		fmt.Fprintln(os.Stderr, "could not encode sanitizer totals")
		os.Exit(1)
	}
	fmt.Println(string(encoded))
}
