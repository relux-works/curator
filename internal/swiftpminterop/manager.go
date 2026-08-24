package swiftpminterop

import (
	"context"

	"github.com/relux-works/curator/internal/swiftpmsource"
)

// Manager is the production entry point for the interop validation stage. It
// owns only C-family target, module, header, system, and boundary closure:
// compilation, linking, and output publication stay downstream.
type Manager struct{ config Config }

// NewManager binds the exact selected C-family drivers, SDK, sysroot, system
// libraries, destination profile, and assurance mode.
func NewManager(config Config) (*Manager, error) {
	if config.Recheck == nil || config.Profile.ID == "" {
		return nil, fail(CodeDerivationUnauthorized, "SwiftPM interop authority is incomplete")
	}
	return &Manager{config: config}, nil
}

// Close validates and republishes one accepted SwiftPM source closure.
func (manager *Manager) Close(ctx context.Context, capture *swiftpmsource.Capture) (*Result, error) {
	if manager == nil {
		return nil, fail(CodeDerivationUnauthorized, "SwiftPM interop manager is absent")
	}
	return Close(ctx, manager.config, capture)
}
