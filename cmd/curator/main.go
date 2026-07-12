// Command curator is the agent environment manager CLI.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/relux-works/curator/internal/version"
)

const usage = `curator: agent environment manager

Usage:
  curator --version
  curator <command> [arguments]

Commands are added phase by phase; see docs/implementation-plan.md.
`

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	flags := flag.NewFlagSet("curator", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.Usage = func() { fmt.Fprint(os.Stderr, usage) }
	showVersion := flags.Bool("version", false, "print the curator version and exit")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if *showVersion {
		fmt.Println("curator " + version.String())
		return 0
	}
	if flags.NArg() == 0 {
		flags.Usage()
		return 2
	}
	fmt.Fprintf(os.Stderr, "curator: unknown command %q\n", flags.Arg(0))
	return 2
}
