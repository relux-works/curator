# Brief: bring CIP-0002 (project context in managed launches) to "ready for acceptance" (curator-spec)

**Repository:** curator-spec. **Branch:** `cip02-ready-for-acceptance`, from `origin/main`. The owner reviews the text with the orchestrator before it becomes Accepted. Your output is the candidate text, with status left at **Draft** and a clearly marked "Proposed for acceptance" disposition block.

## Owner decisions to integrate (all recorded by the orchestrator, verbatim in substance)

**2026-10-10:** already in the CIP ("Operator input"). Keep it.

**2026-10-18:** answers to the eight open questions:
1. Home: option A, a persistent home per checkout × profile × environment. That home is what the platform now calls a **Werkstatt**. Claude login goes through setup-token.
2. Repository as an overlay in the Werkstatt home, plus a sparse Werkstatt worktree outside the repository tree; trusted mode first.
3. Permissions: default deny. We manage all harness permissions ourselves, programmatically. Agents change them only through grants. Project-shipped permissions never apply.
4. Native harness memory off by default; an explicit opt-in parameter.
5. Busy home update: wait or refuse.
6. PATH: D-PATH accepted, with the checked destination transform; skills and CLIs are per Werkstatt.
7. Adapter wave: the first wave is now **Claude, Codex and Muse** (owner, 2026-10-21; this replaces "Pi first, then Claude and Codex, Muse later"). OpenCode and Pi come later; all stay in the design.
8. R-TRANSFER: a foreign candidate is shown as an inert diff.

**2026-10-21:**
- MCP server declarations inline in the project Skillfile are accepted as a project contribution under this CIP, approved per item and bound to the digest of the normalized `server` object. Cross-reference the new MCP-in-Skillfile CIP draft (branch `cip-mcp-in-project-skillfile`; refer to it by title if it has no number on main yet).
- The interim hosted route `launch-delegated/1` (CIP-0016 amendment) is accepted. The project-context launch must state how it composes on that interim route and on the later exact-plan route.

## Close the open items in the "Operative disposition"

- **Home architecture:** decided A. Write the stable-home and lease semantics as they now stand, including the busy-update refusal.
- **Certified adapter delivery:** a separately versioned, optional capability per environment. Advertise the certified surfaces explicitly. An unsupported required item gets a typed refusal.
- **Closed context, lease and transform grammar:** freeze the **smallest executable slice** for the first wave (implementation-plan leaf 1). Name the closed members, refusal codes and fragment-v4 vectors needed for: a registered project, an approved context or skill item, Claude, Codex and Muse source exclusion, inline MCP as an item, and one launch. Defer everything else, listed explicitly.

## Also

- Update "Open questions for the operator" so it lists only what is genuinely still open after these decisions, each with a recommendation.
- Update the implementation-plan table for the new first wave (Claude, Codex, Muse) and the Werkstatt naming.
- Keep the prose rules, links and the validator green. Add no implementation claims.

## Hand-off

Results file: a summary of the decisions integrated, the list of remaining owner questions, and the files changed. Branch pushed, signed.
