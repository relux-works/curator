package swiftpmsource

import "context"

// Manager is the production adapter entry point. It deliberately owns only
// resolution/source closure through C4; C-family validation and compilation
// are downstream boundaries.
type Manager struct{ config Config }

// NewManager binds the concrete evaluator, resolver, broker, verifier,
// toolchain, destination, and shared trust services once.
func NewManager(config Config) (*Manager, error) {
	if config.Store == nil || config.Policy == nil || config.Evaluator == nil || config.Broker == nil || config.MirrorVerifier == nil || config.OfflineRunner == nil || config.CausalHead == "" {
		return nil, fail(CodeDerivationUnauthorized, "SwiftPM manager authority is incomplete")
	}
	if err := validateToolchain(config.Toolchain); err != nil {
		return nil, err
	}
	if err := validateDestination(config.Destination); err != nil {
		return nil, err
	}
	return &Manager{config: config}, nil
}

// CaptureAndClose invokes the complete production capture path.
func (manager *Manager) CaptureAndClose(ctx context.Context, request Request) (*Capture, error) {
	if manager == nil {
		return nil, fail(CodeDerivationUnauthorized, "SwiftPM manager is absent")
	}
	return CaptureAndClose(ctx, manager.config, request)
}
