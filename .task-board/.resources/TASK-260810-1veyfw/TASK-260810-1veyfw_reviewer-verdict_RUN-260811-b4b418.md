# TASK-260810-1veyfw reviewer verdict

Run: RUN-260811-b4b418
Goal: GOAL-260811-e7967b revision 1
Resolved scope: TASK-260810-1veyfw
Verdict: CHANGES REQUESTED
Route: analysis

## Acceptance assessment

The submitted outcome is strong for the surfaces it includes, but it does not satisfy the criterion that a cited evidence matrix record each current implementation or CLI surface. Two current surfaces discovered by the producer are reduced to an uncited paragraph at report lines 153-157 and are absent from the revision ledger and full evidence matrix.

PASS: Go is correctly classified as the implemented Curator baseline. Pinned Curator source confirms schemas 1-7 and Go-only go-v1 / go-repository-v1 driver vocabulary.

PASS: The Node/TypeScript versus Python relationship is explicitly and correctly resolved as protocol-level rather than implementation-level. CocoaSkills at 7f04ae1141c9f1f39f9320e8bb0ca5ad231abf5f declares itself an independent Python implementation, exposes only the csk Python entry point, and has no Node manifest, lock, JavaScript, or TypeScript source.

PASS: Kotlin is kept outside the active investigation. The outcome records only its deferred boundary, including Java 21, Gradle 8.13 network bootstrap without distribution SHA-256, and the tracked 45,633-byte wrapper JAR.

FAIL: Complete current-surface inventory and evidence closure are missing for the following surfaces.

### 1. skill-currency-exchange is a real mixed Go and Swift skill-facing CLI

Authoritative installed-state evidence:
- csk global status reports skill-currency-exchange tag v2.1.1 at c29210aa6eb4cc0f64f307fa30561ac80feb6b3b as up-to-date.
- /Users/iv/.cocoaskills/global/skills/skill-currency-exchange/.csk-install.json binds that commit and installs SKILL.md plus scripts/build.sh and scripts/install.sh.
- The pinned SKILL.md mandates the exchange CLI and documents exchange spawning exchange-scraper.

Authoritative repository evidence at https://github.com/relux-works/skill-currency-exchange/tree/c29210aa6eb4cc0f64f307fa30561ac80feb6b3b:
- exchange/go.mod is Go 1.25.5 and has go.sum, a sizable transitive module graph, and absolute replacements to /Users/alexis/src/relux-works/skill-agent-facing-api/agentquery and /Users/alexis/src/relux-works/skill-go-testing-tools/tuitestkit. Those replacements are undeclared cross-repository closure and portability edges.
- exchange-scraper/Package.swift is Swift tools 6.0, macOS 13, and uses swift-argument-parser from 1.5.0.
- git ls-tree at the pinned commit contains no Package.resolved. The local generated resolver file is not authoritative.
- scripts/install.sh builds the Swift release executable and the Go executable, copies both to the user bin directory, and launches exchange for verification.

This is precisely the mixed-language package and build-order case requested by the task. Calling it context-only does not justify excluding it: the report already treats undocumented and unavailable-on-PATH commands such as appraise as repo-facing surfaces. The same inclusion rule must apply here.

### 2. telegram-telethon is a current Python CLI surface with artifact drift

Authoritative installed-state evidence:
- csk status --all reports the tgiv project surface telegram-telethon at b9a76b01e7ce211c1d0e707f97b231ee7b817d41 with content-drift.
- /Users/iv/Developer/IV/tgiv/.agents/skills/telegram-telethon/.csk-install.json binds the source repository git@github.com:ivanopcode/telegram-telethon.git and the pinned commit.
- Its SKILL.md invokes tg-telethon and auth helper commands. scripts/bootstrap.sh creates a Python venv, installs requirements with pip, and writes three command shims; scripts/tg-telethon launches the venv interpreter.
- The current installed tree contains scripts/__pycache__/telegram_telethon.cpython-314.pyc, a precompiled bytecode payload not present in the marker file list. The installed copy also lacks the requirements input referenced by bootstrap, so the authoritative pinned source must be inspected before dependency and runtime claims can be completed.

The submitted report mentions an undeclared Telegram helper with content drift but gives no repository/path citation, language/package-manager classification, launch entry, dependency shape, runtime requirement, integrity metadata, or precompiled-payload disposition.

## Required corrections

1. State a reproducible estate inclusion rule covering declared Curator commands, shipped repo-facing CLIs invoked by skill instructions or launchers, CocoaSkills project/global surfaces, and external system commands as a separately classified group.
2. Add skill-currency-exchange at c29210aa6eb4cc0f64f307fa30561ac80feb6b3b to the authoritative revision ledger and full evidence matrix. Record both products, both build systems, missing tracked Swift lock, Go absolute replace edges, build order and subprocess boundary, toolchain/runtime requirements, and generated or precompiled payload observations.
3. Inspect telegram-telethon at b9a76b01e7ce211c1d0e707f97b231ee7b817d41 and add it to the matrix, or document a scope rule that excludes it and apply that same rule consistently to every other repo-facing surface. Record the observed pyc drift either way.
4. Re-run the active-root and registered-project scan after applying that rule, cite the exact evidence, and update conclusions and recommendations. In particular, reconsider whether currency-exchange is a better real mixed Go/Swift migration input than an unmentioned estate footnote.
5. Update the existing task-scoped research resource with update_resource rather than creating a duplicate, then return the task for another reviewer cycle.

## Independent checks performed

- Board outcome and .research file SHA-256 both equal 6f513a5c177df122d8ee65f0c8b0ab726ce9b32180869b93e28e7be56085a150.
- Pinned local git objects were read for Curator, Curator Protocol, CocoaSkills, and skill-project-management.
- Primary GitHub tree and raw-file reads verified the SwiftPM manifest, the absence of a tracked Package.resolved at bd59caaf4bb712f35d4b8b73141ce28999cc13cb, and the Android Gradle/JAR facts at b79449cc3e1767680b069ae314c850a1d93c6f99.
- skill-project-management at 8dc0b71490214fe5ead6bf9cfde9574df084fd91 does track the reported 14,501,282-byte Mach-O blob 3cabcb35efb5ba6079cb33ae4754abd52bf5fee9.
- csk version, list, global status, and project status commands all exited 0.

No product code was modified. This is ordinary research rework, not a stop-the-line boundary.