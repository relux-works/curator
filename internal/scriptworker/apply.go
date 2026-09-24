package scriptworker

// applyScriptControls validates the request inventory closed-form and
// applies and confirms each installable control inside the worker. It
// returns the applied control names in inventory order, from which the
// worker builds the record it returns with its acknowledgement.
//
// The parent probed availability once, before the worker launched; the
// worker trusts that list structurally (closed names, known availability
// values) and proves installation: every installable control is confirmed
// in effect here before the readiness proof, except network isolation,
// whose flags install at the interpreter spawn and whose confirmation
// refuses with a failure frame instead of a result when the spawn does not
// isolate, and except the Landlock controls, whose ruleset construction
// is confirmed here while the domain itself is enforced at the spawn on
// the locked spawn thread. A control that cannot be confirmed refuses
// here, before the interpreter starts, with the capability-evidence
// diagnostic — never with the pre-launch control-unavailable diagnostic,
// which only the parent before the launch can produce — except a
// Landlock control, whose construction failure refuses with the
// worker-protocol diagnostic: only the parent's evidence gate verdicts
// probe contradictions, and an installable control is never reported
// `unavailable`, which would contradict the probe that found it present.
func applyScriptControls(request *workerRequest) ([]string, error) {
	switch request.InventoryPlatform {
	case ScriptPlatformLinux, ScriptPlatformMacOS, ScriptPlatformWindows:
	default:
		return nil, diagnostic(CodeWorkerProtocolInvalid,
			"request inventory platform %q is not a known platform", request.InventoryPlatform)
	}
	if len(request.Inventory) != len(scriptInventoryOrder) {
		return nil, diagnostic(CodeWorkerProtocolInvalid,
			"request inventory carries %d controls, want exactly one per inventory control", len(request.Inventory))
	}
	seen := make(map[string]bool, len(request.Inventory))
	installable := make(map[string]bool, len(request.Inventory))
	for _, input := range request.Inventory {
		if !inScriptInventory(input.Name) {
			return nil, diagnostic(CodeWorkerProtocolInvalid,
				"request inventory names control %q outside the inventory", input.Name)
		}
		if seen[input.Name] {
			return nil, diagnostic(CodeWorkerProtocolInvalid,
				"request inventory duplicates control %q", input.Name)
		}
		seen[input.Name] = true
		switch input.Availability {
		case ScriptAvailabilityAvailable, ScriptAvailabilityHostConditional, ScriptAvailabilityUnavailable:
		default:
			return nil, diagnostic(CodeWorkerProtocolInvalid,
				"request inventory control %q reports availability %q", input.Name, input.Availability)
		}
		installable[input.Name] = input.Present &&
			(input.Availability == ScriptAvailabilityAvailable ||
				input.Availability == ScriptAvailabilityHostConditional)
	}
	var applied []string
	confirm := func(name string, confirmFunc func() error) error {
		if !installable[name] {
			return nil
		}
		if err := confirmFunc(); err != nil {
			return err
		}
		applied = append(applied, name)
		return nil
	}
	if err := confirm(ScriptControlDescendantDomainTermination, func() error {
		return confirmScriptTermination(request)
	}); err != nil {
		return nil, err
	}
	if err := confirmScriptAggregateLimits(request, installable); err != nil {
		return nil, err
	}
	for _, name := range []string{ScriptControlActiveProcessCountLimit, ScriptControlAggregateMemoryLimit} {
		if installable[name] {
			applied = append(applied, name)
		}
	}
	if err := confirm(ScriptControlPerFileSizeLimit, func() error {
		return confirmScriptFileSize(request)
	}); err != nil {
		return nil, err
	}
	if err := confirm(ScriptControlInheritedHandleRestriction, confirmScriptHandles); err != nil {
		return nil, err
	}
	execDenial := installable[ScriptControlDescendantExecDenial]
	writeConfinement := installable[ScriptControlFilesystemWriteConfinement]
	if execDenial || writeConfinement {
		// Construction only: the exact ruleset the interpreter spawn
		// will enforce is built and discarded without restricting the
		// caller. Each installable control is confirmed from its own
		// grants, and an installable control whose enforcement cannot
		// be constructed refuses here — a derived path that cannot be
		// ruled is an apply failure of its own control, never a status
		// flip: the record may only report `applied` for a control the
		// probe found present. The worker never manufactures a
		// capability-evidence diagnostic for an apply error; only the
		// parent's evidence gate verdicts probe contradictions.
		execErr, writeErr := confirmScriptLandlock(execDenial, writeConfinement, landlockGrantsForRequest(request))
		// The refusal detail carries the construction cause: a grant
		// that cannot be ruled names its own offending path (quoted,
		// like every other path a diagnostic reports — all grants are
		// manager-derived absolute clean paths), so the operator
		// learns which derived member to fix. Only this invocation's
		// failure detail crosses the session boundary; the wrapped
		// cause stays local to the worker.
		if execDenial && execErr != nil {
			return nil, diagnosticErr(CodeWorkerProtocolInvalid, execErr,
				"cannot install inventory control %q: %s", ScriptControlDescendantExecDenial, execErr)
		}
		if writeConfinement && writeErr != nil {
			return nil, diagnosticErr(CodeWorkerProtocolInvalid, writeErr,
				"cannot install inventory control %q: %s", ScriptControlFilesystemWriteConfinement, writeErr)
		}
		if execDenial {
			applied = append(applied, ScriptControlDescendantExecDenial)
		}
		if writeConfinement {
			applied = append(applied, ScriptControlFilesystemWriteConfinement)
		}
	}
	if installable[ScriptControlNetworkIsolationDomain] {
		// Installed by the interpreter spawn itself (see
		// runInterpreter): the flags are fixed here, and the spawn
		// confirms the fresh namespace before returning a result.
		if err := checkScriptNetNS(request); err != nil {
			return nil, err
		}
		applied = append(applied, ScriptControlNetworkIsolationDomain)
	}
	ordered := make([]string, 0, len(applied))
	appliedSet := make(map[string]bool, len(applied))
	for _, name := range applied {
		appliedSet[name] = true
	}
	for _, name := range scriptInventoryOrder {
		if appliedSet[name] {
			ordered = append(ordered, name)
		}
	}
	return ordered, nil
}

// scriptProbesFromInput converts the request inventory to probe records so
// the worker builds its record with the same constructor the parent uses.
func scriptProbesFromInput(inputs []ScriptInventoryInput) []ScriptControlProbe {
	probes := make([]ScriptControlProbe, 0, len(inputs))
	for _, input := range inputs {
		probes = append(probes, ScriptControlProbe{
			Name: input.Name, Availability: input.Availability,
			Present: input.Present, ProbedAt: ScriptProbeTiming,
		})
	}
	return probes
}
