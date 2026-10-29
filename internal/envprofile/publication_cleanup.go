package envprofile

import (
	"errors"
	"fmt"
)

// Cleanup may fail after a commit or after a refusal. Preserve the original
// outcome and identify cleanup explicitly; never report a fictitious rollback.
func finishPublicationCleanup(result *error, closeResource func() error) {
	if err := closeResource(); err != nil {
		*result = errors.Join(*result, fmt.Errorf("publication resource cleanup: %w", err))
	}
}
