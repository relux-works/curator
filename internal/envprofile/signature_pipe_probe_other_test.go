//go:build !windows

package envprofile

import "os"

func signaturePipeMetadataProbe(path string) error {
	_, err := os.Stat(path)
	return err
}
