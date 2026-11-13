package envprofile

import (
	"bytes"
	"fmt"
)

// splitGitSignedPayload preserves the exact unsigned bytes covered by Git's
// embedded commit header or annotated-tag signature. It never normalizes LF.
func splitGitSignedPayload(objectType string, payload []byte) (data, signature []byte, err error) {
	if objectType == "tag" {
		offset := -1
		for _, armor := range [][]byte{[]byte("-----BEGIN SSH SIGNATURE-----"), []byte("-----BEGIN PGP SIGNATURE-----")} {
			if i := bytes.LastIndex(payload, armor); i > offset {
				offset = i
			}
		}
		if offset < 0 {
			return nil, nil, fmt.Errorf("tag signature armor is absent")
		}
		return payload[:offset], payload[offset:], nil
	}
	if objectType != "commit" {
		return nil, nil, fmt.Errorf("unsupported signed object type")
	}
	header, body, ok := bytes.Cut(payload, []byte("\n\n"))
	if !ok {
		return nil, nil, fmt.Errorf("commit header is incomplete")
	}
	collecting, found := false, false
	for _, line := range bytes.Split(header, []byte("\n")) {
		if bytes.HasPrefix(line, []byte("gpgsig ")) || bytes.HasPrefix(line, []byte("gpgsig-sha256 ")) {
			if found {
				return nil, nil, fmt.Errorf("multiple commit signature headers require explicit verification support")
			}
			found = true
			collecting = true
			_, value, _ := bytes.Cut(line, []byte(" "))
			signature = append(signature, value...)
			signature = append(signature, '\n')
			continue
		}
		if collecting && bytes.HasPrefix(line, []byte(" ")) {
			signature = append(signature, line[1:]...)
			signature = append(signature, '\n')
			continue
		}
		collecting = false
		data = append(data, line...)
		data = append(data, '\n')
	}
	if !found {
		return nil, nil, fmt.Errorf("commit signature header is absent")
	}
	data = append(data, '\n')
	data = append(data, body...)
	return data, signature, nil
}
