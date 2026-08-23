//go:build windows

package artifactpolicy

import (
	"os"

	"golang.org/x/sys/windows"
)

func regularFileLinkCount(file *os.File, _ os.FileInfo) (uint64, bool, error) {
	var information windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(windows.Handle(file.Fd()), &information); err != nil {
		return 0, false, err
	}
	return uint64(information.NumberOfLinks), true, nil
}
