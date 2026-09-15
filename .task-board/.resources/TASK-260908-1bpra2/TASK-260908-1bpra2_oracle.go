package main

import (
 "fmt"
 "os"
 "github.com/relux-works/curator/internal/contextpkg"
)
func main() {
 failed := false
 for _, path := range os.Args[1:] {
  b, err := os.ReadFile(path)
  if err == nil { _, err = contextpkg.ParseMCP(b) }
  if err != nil { fmt.Printf("%s: REJECT %v\n", path, err); failed = true
  } else { fmt.Printf("%s: ACCEPT\n", path) }
 }
 if failed { os.Exit(1) }
}
