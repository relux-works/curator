//go:build !windows

package artifactpolicy

import (
	"os"
	"reflect"
)

func regularFileLinkCount(_ *os.File, info os.FileInfo) (uint64, bool, error) {
	value := reflect.ValueOf(info.Sys())
	if !value.IsValid() {
		return 0, false, nil
	}
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return 0, false, nil
		}
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return 0, false, nil
	}
	field := value.FieldByName("Nlink")
	if !field.IsValid() {
		return 0, false, nil
	}
	switch field.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return field.Uint(), true, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Int() < 0 {
			return 0, false, nil
		}
		return uint64(field.Int()), true, nil // #nosec G115 -- the signed value was checked nonnegative immediately above.
	default:
		return 0, false, nil
	}
}
