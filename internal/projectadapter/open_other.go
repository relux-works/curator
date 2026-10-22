//go:build !darwin && !linux

package projectadapter

import (
	"fmt"
	"os"
)

func openGenerated(*os.Root, string) (*os.File, error) {
	return nil, fmt.Errorf("project_context_unsupported: native materialization read boundary is unqualified on this platform")
}
