package main

import (
	_ "embed"
	"fmt"

	"example.test/vendored/message"
)

//go:embed message.txt
var prefix string

func main() {
	fmt.Print(prefix + message.Value())
}
