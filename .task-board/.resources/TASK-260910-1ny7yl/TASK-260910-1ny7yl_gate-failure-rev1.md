# TASK-260910-1ny7yl — hosted gate failure on Change Request revision 1 (run 35261471138)

Extracted by the orchestrator. Five of six lanes are green; `Tests / Python
3.14 on ubuntu-latest` fails ONE test that this task did not write:

```
tests/test_registry.py::test_health_stale_verifier_fails_closed_and_refresh_recovers
tests/test_registry.py:2026  assert client.get("/health").status_code == 200   → 503
```

That is the R2 staleness test (landed in `131952d`) at its "age == bound is
still fresh" step: it captures `base = store._health_last_refresh_monotonic`,
installs a manual clock and expects `/health` 200 at `base + 20.0`. The
background verifier thread keeps running under the monkeypatched module
clock, so a refresh that starts or finishes around that instant can observe
a different `now`/`last_refresh` pair than the assertion assumes; the result
is lane-dependent flakiness, not a product defect of `serve --checkpoint`
(the checkpoint tests all passed on every lane).

Fix inside this revision, harness-only, in that test: stop (or pause) the
background verifier before installing the manual clock so that only the
explicit `refresh_health_verdict()` calls advance the verdict (use the
store's verifier lifecycle API — the thread is created in `store.py`; add a
small `pause_health_verifier()`/`resume` or a constructor flag for tests if
none exists and keep it out of the production path), and read `base` AFTER
the pause. Keep the exact-boundary assertions. Do not widen the bound, do
not add sleeps, do not skip the lane.

Re-run `python -m pytest -q` (with `CURATOR_CONFORMANCE_ROOT`) and hand off
again; the runtime re-runs the hosted gate.
