//go:build windows

package godriver

import (
	"testing"

	"golang.org/x/sys/windows"
)

// TestWorkerRejectsMissingPrivateJobHandle drives the real worker entry point
// with the manager-installed job but without the query-only handle that binds
// its evidence to that exact object. The worker must refuse before a compiler
// starts instead of querying an ambient current job.
func TestWorkerRejectsMissingPrivateJobHandle(t *testing.T) {
	scenario := newWorkerScenario(t)
	worker := startRawWorkerFrom(t, scenario.identity.Path, scenario.limits, scenario.probes)
	request := scenario.request
	request.ControlJobHandle = 0
	worker.send(workerMessage{Kind: kindRequest, Nonce: scenario.nonce, Request: &request})
	worker.expectFailure(CodeCapabilityEvidenceInvalid)
	scenario.requireNoCompilerStarted()
}

// TestWorkerRejectsForeignPrivateJobHandle proves that matching limits alone
// do not attest the manager-installed domain. The worker must also belong to
// the exact Job Object whose limits it reports.
func TestWorkerRejectsForeignPrivateJobHandle(t *testing.T) {
	scenario := newWorkerScenario(t)
	worker := startRawWorkerFrom(t, scenario.identity.Path, scenario.limits, scenario.probes)
	if worker.domain.flags == 0 {
		t.Fatal("the test inventory has no Job Object-backed control")
	}

	foreignJob, err := newControlJob(worker.domain.flags, scenario.limits)
	if err != nil {
		t.Fatal(err)
	}
	defer windows.CloseHandle(foreignJob)
	process, err := windows.OpenProcess(windows.PROCESS_DUP_HANDLE, false, uint32(worker.command.Process.Pid))
	if err != nil {
		t.Fatal(err)
	}
	var workerJob windows.Handle
	err = windows.DuplicateHandle(windows.CurrentProcess(), foreignJob, process,
		&workerJob, jobObjectQueryAccess, false, 0)
	_ = windows.CloseHandle(process)
	if err != nil {
		t.Fatal(err)
	}

	request := scenario.request
	request.ControlJobHandle = uint64(workerJob)
	worker.send(workerMessage{Kind: kindRequest, Nonce: scenario.nonce, Request: &request})
	worker.expectFailure(CodeCapabilityEvidenceInvalid)
	scenario.requireNoCompilerStarted()
}
