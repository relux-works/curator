# Review note — BUG-260916-3aco9f install/reinstall --use/--takeover matrix (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review against `3aco9f-brief.md`. Candidate: cmd/curator/profile_install_matrix_test.go + internal/registry/registry_test.go (the latter is
unexpected for this bug — name why it changed; a flake fix must be justified and must not weaken an assertion). Verify: the matrix covers
every install shape × {first, same-source reinstall, changed-source reinstall} × {--use, --takeover, neither} through the CLI production entry;
any shape found broken was fixed in production (if only tests changed, every shape already passed — confirm by reproducing two rows and a
mutant that drops the activation step for one shape). Gate green. accept_cr or changes requested with file:line. No LOGBOOK.md.
