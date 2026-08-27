//go:build windows

package transaction

func namespaceCaseInsensitive(string) (bool, error) {
	return true, nil
}

func namespaceNormalizationInsensitive(string) (bool, error) {
	return false, nil
}
