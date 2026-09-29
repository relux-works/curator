# Review note — TASK-260929-1kze6z naming gate vs binary-patch noise (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review rev1 (base 7f2fb6b8, tree e5318a75, 3 paths, gate green including the Naming gate on a tree that contains the 31gaka binary patch)
against `1kze6z-brief.md`. Verify:
1. No committed file spells either name. Both patterns are assembled from parts.
2. The only exemption is literal/delta/blank/base85-shaped lines inside a `GIT binary patch` block, which ends at the next `diff --git` or
   EOF. Everything else is still scanned. Try to find a bypass:
   - a markdown file containing a line "GIT binary patch" followed by a base85-shaped line that contains the short name (for example the
     short name alone, or the short token glued to base85 characters). Decide whether that is an acceptable residual: base85-shaped
     lines cannot contain spaces or dots, so a real sentence cannot hide there. Say so explicitly;
   - a file whose name does not end in .patch — say whether the exemption applies and whether it should;
   - CRLF line endings.
3. Selftest rows a–e exist and fail for the naming reason. Re-run the two mutants yourself (no skip; no base85 shape check) with real exit
   codes and confirm each is killed.
4. `bash .github/ci/naming-gate.sh` on the candidate tree exits 0. Put the short word on a normal line of a temp .md file: it exits 1.
accept_cr or changes requested with file:line. No LOGBOOK.md.
