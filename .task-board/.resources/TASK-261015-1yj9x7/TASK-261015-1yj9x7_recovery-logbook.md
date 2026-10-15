# TASK-261015-1yj9x7 recovery logbook

The first Change Request validation failed in the hosted Windows platform-case gate because TestGitToolRejectsSymlinkedBrokers printed an unclassified skip reason. Hosted native Windows go test exited 0; the platform-case gate exited 1. The full configured gate therefore failed and is not reported as green.

Recovery changes only the skip message in internal/buildrepo/admission_test.go to "controlled broker symlink fixture is exercised on unix runners". This accurately uses the existing platform-control class without changing CI policy or production admission. A synthetic stream through the production platform-case gate reproduces the old exit 1 and the candidate exit 0. Native Windows was not rerun by this recovery session; the Change Request flow reruns the configured hosted validation.

Current direct checks: full buildrepo tests, lint, build, formatting, suppression gate and scope audit exit 0. Both field-specific symlink-admission overlay mutants exit 1 in their named negative subtests. Full gate self-test evidence is recorded separately after its process exits.

The recovery diff stays within internal/buildrepo; the candidate remains uncommitted. No control-root logbook file was edited. This task-scoped logbook resource carries the record through the Change Request flow.
