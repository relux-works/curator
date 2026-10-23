//go:build !unix

package main

import (
	"os"

	"github.com/relux-works/curator/internal/godriver"
)

func runNativeBuildWorker() int { return godriver.RunWorker(os.Stdin, os.Stdout) }
