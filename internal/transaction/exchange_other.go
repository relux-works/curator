//go:build !darwin && !linux

package transaction

import "fmt"

const exchangeSupported = false

func durableExchange(a, b string) error {
	return fmt.Errorf("atomic name exchange is unavailable")
}
