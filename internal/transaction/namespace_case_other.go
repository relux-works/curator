//go:build !darwin && !windows

package transaction

func namespaceCaseInsensitive(string) (bool, error) {
	return false, nil
}

func namespaceNormalizationInsensitive(string) (bool, error) {
	return false, nil
}
