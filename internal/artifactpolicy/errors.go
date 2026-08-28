package artifactpolicy

import "errors"

func errorAs(err error, target any) bool {
	return errors.As(err, target)
}

func errorIs(err, target error) bool {
	return errors.Is(err, target)
}
