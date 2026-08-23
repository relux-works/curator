# TASK-260729-2vfvgi: run-windows-kotlin-driver-pair-qualification

## Description
Qualify the selected local and external Kotlin driver pair on a Windows amd64 host, which decision 0010 makes the only tuple that can admit the pair before Linux enters the protocol platform set.

## Scope
Curate a curator-kotlin-bundle-v1 root on Windows per decision 0010 section 3 (official archive plus published checksum, prehydrated dependency closure, JDK inside the root, read-only tree); run the decision 0010 section 12 acceptance test A1-A9 for windows/amd64 with recorded argv and real exit codes; supply the tuple base-installation library allow-list required by the published-artifact gate; record the overlay materialisation mechanism available on Windows; fill primary_relpath, probe, platforms and compatibility only from that evidence.

## Acceptance Criteria
A1-A9 all pass for windows/amd64 with recorded argv and real exit codes, or the tuple is recorded as failed with the exact blocking evidence; every spawned child resolves inside the fingerprinted bundle or the operation-private overlay with the containment control shown to fire; no cmd.exe, powershell.exe, link.exe, cl.exe, lib.exe, vswhere.exe or Visual Studio activation script appears in the process graph; the compile performs no download under airplaneMode plus network denial and leaves the bundle byte-unchanged; the tuple base-installation allow-list is supplied; no registry field is filled by assertion and no platform is claimed without that evidence.
