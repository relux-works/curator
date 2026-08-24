// Command golden-tool is the multi-module vendored build fixture.
package main

import (
	"fmt"

	"example.test/board"
	"example.test/remoteconfig"
)

func main() { fmt.Println(board.Name(), remoteconfig.Value()) }
