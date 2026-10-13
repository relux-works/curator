# Wave 5 (2026-10-13): worker code transfer, the untrusted class and abuse controls. Owner decisions and orchestrator resolutions

The owner decided the items below on 2026-10-13. They take precedence over any earlier text. The wave-3, wave-4 and wave-4b shared names (`gw4-decisions.md`, `gw4b-decisions.md`) still apply unchanged.

## Owner decisions (operator, 2026-10-13): apply exactly

1. **One path for code out of an isolated account.**
   - The worker never has network access, push access or merge-request access.
   - Its code leaves the account only as data that the platform takes.
   - It becomes a commit only in our repository.
   - It is reviewed as that exact commit.
   - It lands only through the board's existing path.
   - Rework goes back to the worker with the reviewer's findings, on top of the last candidate.
   - The steps are R-TRANSFER below. The owner accepted them as proposed.
2. **Owners of the path.**
   - Materialization and export belong to the dispatcher's workspace contract. They run as approved executors under the child account.
   - The board only accepts the result as a resource.
   - Import, candidate, review and integration belong to the board's change-request machinery, and the orchestrator executes them.
3. **Internet is a tool of the agent loop, not network access of the account.** The owner proposed this as the design direction to work out.
   - In our environment an agent account has no internet. For a remote workplace, nothing exists besides the SSH connection.
   - A remote worker that needs the internet reaches it through its own agent loop on the donor side. One example is an MCP server we ship, which receives the donor's own search token in a way that the project can never obtain.
   - See R-NET-TOOL.
3a. **The design works without web access** (owner, 2026-10-13, later the same night): «я думаю мы должны иметь дизайн который работает и без веб поиска, где задача для модели несет потенциально нужные ресурсы на исполнение (свежие документации и тд) досутп в интернет это потенциальное опасно тк смешиваются контексты (что то попадает в конеткст из среды, что то из интерента) так что мы должны уметь работать когда цикл донора ходит только в эгндпоинты провайдера и к нам, а у нас интернет отключен. поддерживать возможность поиска нужно и описать в дизайне нужно, тк это может быть полезно когда удаленный агент это часть доверенного круга.»
   - By default, the donor's loop reaches only its model-provider endpoints and our host, and our environment has no internet.
   - Tasks carry the resources they need.
   - Web search is described in the design and supported only as an opt-in for the trusted circle.
   - See R-NET-TOOL items 4, 4a and 4b.
4. **No builds or tests in agent accounts.**
   - Agents must not run calls that can load the machine. Type checking stays allowed. Compilation, transpilation, test suites and similar work go only through a controlled verification phase in isolated ephemeral runners on another machine.
   - A watcher monitors agent processes.
   - See R-EXEC, R-VERIFY and R-WATCH.
5. **Standing agents.**
   - A standing agent with an agent loop on our host is trusted only.
   - Agents from outside may also be long-lived, as a standing workplace. They can never start their own agent loop or agent environment on our host.
   - The operating-system services they may use are a whitelist for their task.
   - See R-STANDING and R-EXEC.
6. **The independent infrastructure review of 2026-10-13 is accepted, items 1–11:**
   - range deny;
   - untrusted class;
   - precedence check;
   - proxy requirements;
   - Linux ceilings;
   - kernel baseline;
   - node-wide emergency stop;
   - per-UID telemetry;
   - harness-configuration persistence;
   - socket reachability;
   - no shared services beside untrusted work.
7. **No operating-system restriction at platform level.** The platform itself must not exclude macOS for any class. Whether macOS has critical blockers is worked out by qualification: admission follows each host's published, qualified capability matrix. See R-CEILINGS.

## Orchestrator resolutions: shared names

### R-TRANSFER: the code transfer path

1. **Materialize.**
   - The caller supplies the admitted exact base B as a bounded canonical archive of B's tree with the start request. B is a commit of the caller's repository.
   - The dispatcher relays the archive and journals its digest. The dispatcher holds no repository.
   - An approved executor, design label `swarma-workspace-materialize`, runs as the child UID before the work execution. It writes a standalone repository into the child's private home: freshly written local configuration, no hooks, no remotes, no credential helpers.
   - The launch plan records B, the digest of B's file manifest (path to blob id), the task's **write scope** and the exec profile (R-EXEC). The plan digest covers them.
   - Tests and other validators of the result stay outside the write scope unless the task is about them: a worker must not be able to weaken what judges its own work.
   - Dependencies for light checks come only from read-only caches seeded by the platform; nothing is downloaded inside the account.
   - The archive format and the exact plan fields are unpublished.
2. **Work.** The worker's own `.git`, commits and history are never used by the platform.
3. **Round end.**
   - The worker reports done: a board outcome for a local agent, a signed final envelope for a remote worker.
   - The platform confirms that no process of the account remains, using the user manager's account-wide stop scope.
   - For a workplace, the round's workspace is write-fenced: the project-side executor refuses further tool calls into it.
4. **Export.**
   - An approved executor, design label `swarma-workspace-export`, runs as the child UID only after confirmed absence. It has a fixed environment and no network.
   - It walks the round's workspace without following links, skips every `.git` component, hashes files and compares them with the manifest from the plan.
   - It emits only the change set: added or modified regular files (content and executable bit), symlinks as link text, and deleted paths.
   - Caps: file count, per-file bytes and total bytes. Binary files have their own cap.
   - These refuse the whole export, with no partial result: special files; any path with a `.git` component however spelled; case-fold or Unicode-normalisation collisions; gitlinks or submodules; paths escaping the root; an exceeded cap.
   - The dispatcher journals the digests and returns the result only to the original caller principal, the requester of the run. The result kind has the design label `swarma-dispatcher workspace-result/1`; its schema and retention are unpublished.
   - **Export precedes retirement.** A workspace-capable allocation carries an export cleanup obligation. Home deletion waits until the export is recorded or the operator waives it by an operator act (for example after compromise). It is never silently dropped.
5. **Candidate.**
   - The orchestrator runs the board's deterministic import. This is task-board lane work; the command name is unpublished.
   - The import builds the tree on base B in a scratch index, with hooks disabled and no attribute filters or drivers.
   - **Hard refusals:** any path with a `.git` component; any path outside the write scope; any change under `.task-board/`.
   - **Flagged, review-required paths:**
     - CI: `.github/**`, `.gitlab-ci.yml`, `.circleci/**`;
     - git behaviour: `.gitattributes`, `.gitmodules`;
     - hooks and auto-run: `.githooks/**`, `.husky/**`, `.pre-commit-config.yaml`, `lefthook.yml`, `.envrc`, `.vscode/**`;
     - agent instructions and configuration: `AGENTS.md` and `CLAUDE.md` in any directory, `.claude/**`, `.codex/**`, `.agents/**`, `.mcp.json`;
     - every symlink.
   - The result is candidate commit C, signed by the repository identity, with trailers naming the worker identity, the run and the result digest.
   - C's tree must equal the tree recomputed from B plus the change set, so anyone can re-check the import.
6. **Review.**
   - A trusted reviewer reviews exactly C. The reviewer is not of the untrusted class and is not the producing identity.
   - Acceptance binds to C (board kernel rule K1). It is never transferred to another candidate by patch equality.
   - Flagged paths need explicit acknowledgement, hunk by hunk, in the verdict.
   - Executing C's code (tests, builds, generators) happens only in R-VERIFY.
   - Before acceptance, a candidate from an untrusted-class worker is never pushed where CI would run it with secrets.
7. **Rework.**
   - A changes-requested verdict puts its findings on the board, and the next round starts with base C, which is an admitted commit of ours.
   - A remote worker gets a signed task letter. Its donor-side session continues, and the workspace is materialized afresh from C. The previous round's workspace is removed after its export is recorded.
   - A local agent gets a new run from C, with the findings in its task snapshot.
   - The next export yields C2 on top of C.
   - Questions travel as ordinary letters or outcomes. An intermediate export may be requested at a turn boundary under the same fence.
8. **Integration.**
   - The accepted exact head integrates by signed fast-forward (K1/K3).
   - If the base has moved, the board's ordinary base refresh applies, with re-review.

Reference: the remote-workplace protocol v2.1 (RWP), §8 change extraction and §10 orchestrator obligations. Its gates are adopted. The only difference is who exports: an approved executor under the account, instead of the harness reading a container volume.

R-TRANSFER closes dispatcher open question 22 and user manager open question 15, except for the formats it names as unpublished.

### R-UNTRUSTED: the untrusted admission class

1. **Who is in it.**
   - Remote workers (donors and freelancers) come first. Later, any identity the operator classifies as untrusted.
   - Every workplace allocation is untrusted-class, unless the operator admits that remote worker to the trusted circle by an operator act.
   - Trusted-circle membership enables only donor-side opt-ins such as web search (R-NET-TOOL 4b). It never relaxes a boundary on our host: zero egress, the exec profile, the transfer path and review all stay.
2. **Where the class is recorded.** At registration or admission. It is never inferred from an account name.
3. **What admission on a host requires:**
   - egress per R-NET-TOOL: zero egress for workplaces, and never `direct`;
   - ceilings per R-CEILINGS that meet the class minimum;
   - an exec profile per R-EXEC;
   - no subscription binding or lease ever: the broker refuses them;
   - no agent loop on our host (R-STANDING);
   - no shared services on the same host (R-SHARED).
4. **Where it is refused.** Slice 0 and user-manager v0 refuse the class. The refusal code is unpublished.

### R-NET-TOOL: internet as a tool of the agent loop

1. **No general internet for agent or workplace accounts.**
2. **Workplaces have zero egress.**
   - There is no proxy destination, and loopback is denied.
   - Linux: a per-execution network namespace without external interfaces, plus address families restricted to Unix sockets.
   - macOS: a sandbox profile that denies network, plus the per-UID deny.
   - Both mechanisms need qualification.
3. **Local trusted agents** reach only their model-provider endpoints, through the bound profile proxy with vendor names in the proxy allowlist.
4. **Default: no web tool for any agent.**
   - The design works without web access: the donor's loop reaches only its model-provider endpoints and our host, and our environment has no internet.
   - No search or fetch tool is offered by default. The CIP-0008 lockdown keeps provider-side search disabled, and no donor-side web server runs.
   - Rationale: web content mixes an outside context into the model's context next to the environment's own, which invites injection.
4a. **Task resources.**
   - A task carries the resources it may need: fresh documentation, API references, specifications, examples.
   - The trusted side gathers them and attaches them to the task as board resources. For a remote worker they also travel as attachments of the signed task letter.
   - The platform materializes them read-only next to the workspace, with their digests in the plan. They are data, never instructions.
   - If the worker needs more, it asks through an ordinary letter or outcome. The trusted side fetches, checks and attaches. The worker never fetches.
   - Caps and formats are unpublished.
4b. **Opt-in web for the trusted circle only.**
   - The trusted circle is our own agents, plus remote workers that the operator admits to it by an operator act at admission.
   - A web tool is granted explicitly per identity, never by default, and never to the untrusted class.
   - The options below are design options with unpublished formats:
   - **(a) Provider-side search.** The harness's server-executed search, where a qualified tuple shows that it runs at the provider and adds no local fetch. For remote workers this needs a CIP-0008 tuple change.
   - **(b) Donor-side MCP server for remote workers,** shipped by the platform (design label `swarma-web`):
     - It runs under its own donor-side identity in variant B, or under its own sandbox profile in variant C. It never runs in the harness identity, whose egress is the provider allowlist only, and never in the bridge identity.
     - Its egress is the search provider's API endpoints only.
     - It holds the donor's own search key in that identity (or in the donor's Keychain or a 0600 file in variant A).
     - The key never crosses the channel, is never part of a `swarma-rb/1` frame or package, and the project has no method to request it.
     - It calls search-provider endpoints only, never an arbitrary fetch from the donor's address.
     - The donor consents at `init`.
     - Results reach only the model at the donor, as tool results.
   - **(c) Platform web-tool service for local agents,** reached over a Unix socket. Its key lives in a service account, never in an agent account. It has an allowlist, rate limits and logs.
5. **Dependency downloads happen only in R-VERIFY runners,** under their own controlled egress.
6. **No `direct` and no `external-http-proxy` on a platform host.** Neither can be bound to any account, trusted or untrusted:
   - `direct` is unbounded;
   - an external proxy cannot show served-policy evidence.
   Both stay for operator machines without the platform, as do cooperative profiles. This is an orchestrator resolution derived from owner decision 3, proposed by the network-profiles lane.
7. **One proxy listener per account generation,** bound to its profile, so caps and logs are attributable. The account's row admits only its own port. This is the network-profiles lane's N15 revision; it is referenced here, not specified.
8. **Proxy requirements from the review** apply to the model-provider proxy and to the runners' dependency egress:
   - host and port allowlists, with no mail ports;
   - a private and metadata dial guard;
   - per-identity rate, connection and daily-volume caps;
   - logs.
   - RWP's egress service is a candidate implementation.
   - The proxies belong to the network-profiles owner (N15). They are referenced here, not specified.

### R-EXEC: exec profile per task class

1. **An allowlist of executables** that an agent or workplace account may run, per task class. It is enforced by the operating system, never by PATH alone:
   - Linux: a per-execution unit where all paths are non-executable except the allowlist;
   - macOS: a sandbox profile that restricts process execution to the allowlist.
   - Writable areas are never executable.
2. **Excluded for every agent class:**
   - compilers, linkers, transpilers and bundlers;
   - test runners;
   - package managers and installers;
   - container and VM tools;
   - network tools;
   - harness and agent binaries. No agent loop or agent environment starts from an agent or workplace account; only the executor of the admitted plan starts the harness.
3. **Allowed by profile:**
   - editing tools, version control for local inspection, and search;
   - light checks such as type checking, per toolchain qualification.
   - Where a light check and a heavy action share one executable, only the protected wrapper in the manager dispatcher directory (D-PATH) may invoke it, with a fixed subcommand allowlist. Qualification proves that the heavy form refuses.
   - Interpreters are allowed only where the profile needs them. Ceilings remain the hard bound.
4. **The dispatcher job class `agent-headless-build` is withdrawn** for agent accounts. Builds run only in R-VERIFY.

### R-VERIFY: the verification phase

1. Building and testing a candidate, and any execution of worker-produced code, run only in isolated ephemeral runners:
   - on a machine other than the agents' host;
   - without secrets;
   - with dependency egress per R-NET-TOOL.
2. Results return to the board as data, in the testing state.
3. An agent role such as the tester orchestrates the phase and reads the results; it never runs the code itself.
4. The runner pool, the job contract and the results format are unpublished. They belong to the board and infrastructure lanes.

### R-CEILINGS: resource ceilings and the host capability matrix

1. **The capability matrix.**
   - Each host publishes a qualified capability matrix per OS release. For CPU, memory, process count, IO, disk, temporary areas, process visibility and network, it states whether the ceiling is kernel-enforced (hard), watcher-enforced (soft) or absent.
   - Admission classes declare minimums, and the dispatcher admits a class on a host only when the host's matrix meets them.
   - No operating system is excluded by name.
2. **Untrusted-class minimums:**
   - a hard process-count ceiling;
   - enforced zero or provider-only egress;
   - an exec profile;
   - isolated temporary areas;
   - hard memory, CPU and disk ceilings where the host provides them.
   - Whether soft rows are acceptable on a host that lacks hard ones is recorded per host by an operator act. This is an open owner decision.
3. **Linux reference mechanism.** The root-created, non-delegated per-execution unit (the R-FATE option) is **selected** for agent and workplace executions:
   - CPU, memory, task and IO limits;
   - a private temporary directory, no new privileges, a read-only system;
   - a per-UID disk quota on the agents filesystem;
   - `KillUserProcesses=yes` and `/proc` with `hidepid=2`;
   - the exec allowlist (R-EXEC), the network namespace and address families (R-NET-TOOL);
   - a CPU-time budget per execution.
4. **macOS reference mechanism:**
   - a sandbox profile for exec, files and network;
   - a per-user process-count limit and per-process CPU limits;
   - the background scheduling class for agent processes, which throttles CPU and IO;
   - an APFS volume quota for the agents root (aggregate, not per account);
   - the watcher for memory and per-account disk (soft).
   - Known gaps to qualify: no public hard per-account memory ceiling, no per-user disk quota on APFS, and a deprecated, though widely used, sandbox profile interface. They are recorded as soft or absent rows; they do not exclude macOS.
5. **Kernel baseline rows** for hosts that admit the untrusted class (Linux):
   - unprivileged user namespaces, io_uring and unprivileged eBPF disabled;
   - `kernel.yama.ptrace_scope=2`;
   - timely sudo patching.
   - Host infrastructure installs them; the platform verifies them.

### R-WATCH: the usage watcher

1. **Telemetry per UID:** CPU, memory, process count, home disk usage, network counters and refused-connection counters, including mail ports.
2. **The sampler** is read-only and has the least privilege the operating system requires. It runs as root only where the OS requires that, and holds no other authority.
3. **Thresholds** are set per admission class.
4. **On a hard threshold** the watcher requests automatic narrowing of that one account only: freeze its egress, and a journaled stop through the dispatcher (`exec.stop`). It never kills without a journal (R-FATE).
5. **Alerts** go to the operator over the phone path. The node-wide stop stays manual (R-ALLSTOP).
6. Thresholds, the sampler interface and the alert format are unpublished.

### R-ALLSTOP: the node-wide emergency stop

1. **The act.** One operator narrowing act: phone path with authentication and one confirmation, or the terminal fallback. Like every operator act, it is refused from inside agent sessions.
2. **What it does:**
   - the dispatcher sets hold, which closes admission;
   - the user manager flips the agent and workplace UID range to deny-all (R-RANGE-DENY);
   - every agent and workplace generation stops, with confirmed absence (account-wide scope, journaled);
   - the broker freezes all leases and renewals. Tokens already delivered keep their vendor lifetime, and the text says so;
   - the session host closes delivery and snapshots bound boards.
3. **Recovery.** Recovery is explicit, never an automatic restart. Each service journals its part. A partial or unknown result keeps admission closed.
4. The act and record formats are unpublished.

### R-RANGE-DENY: the static range rule

1. **Scope.** A static default deny for the agent and workplace UID range (30000–59999). The installer places it with the helper-owned parent anchor or table, before any agent account exists, independent of per-account `fw.apply`.
2. **From v0.** The range rule admits only the platform's proxy listener or listeners. Per-account narrowing to the bound profile, or to zero egress, arrives with v1 `fw.apply`. The untrusted class requires v1 regardless.
3. **Linux:** an nftables `meta skuid` range rule with counters and logging that covers every protocol.
4. **macOS:** pf `user` range rules plus the launcher-applied sandbox for traffic other than TCP and UDP.
5. **Precedence.**
   - Verification proves that no earlier `quick` rule, in any anchor including host ingress anchors, can match agent-range traffic on any interface, including loopback and tunnel interfaces.
   - The proof is an active probe from a dedicated test UID, run at install, at every service start and on the verification schedule.
   - A failed probe is an incident: enforced readiness drops.

### R-STANDING: standing placements and harness surfaces

1. **Who may stand.** A standing session with an agent loop on our host is admissible for trusted-class identities only. An untrusted standing placement is a workplace only; its loop runs at the donor, and no harness binary is executable from it (R-EXEC).
2. **Harness surfaces of trusted standing agents.** Settings, hooks, MCP configuration and instruction files come from the plan.
   - The executor verifies them by digest before every launch, restart and native resume, or supplies them from a root-owned location outside the home.
   - A drifted surface refuses the launch and is reported. The refusal code is unpublished.
   - Repair is an explicit re-materialization, never silent acceptance.

### R-SOCKETS and R-SHARED: qualification and deployment rows

1. **R-SOCKETS.** The set of Unix-domain sockets that an agent or workplace UID can connect to must equal the declared platform set.
   - Container runtimes, the tailnet daemon, other users' session daemons and every other socket must be unreachable.
   - A probe from a test UID checks this on each host.
2. **R-SHARED.** A host that admits the untrusted class runs no shared services: no message server, ingress or reverse proxy, media server, or operator session daemon. Platform services themselves are not shared services.

## Rules for every producer of this wave

1. Public repositories: no first names, host names, local absolute paths, session links or board-resource links.
2. Docs only. Do not run tests or builds.
3. Never invent wire fields, refusal codes or numeric values beyond the names above. Where a format is unpublished, say so and add or extend an open question ("would need …").
4. Remove superseded text, and update every figure that shows changed behaviour.
5. Add a revision note at the top in the existing style.
