# THE ONLY CURRENT INSTRUCTION — research: Android as the base OS for the Swarma host (researcher, read-only, conceptual)

## Question (owner, 2026-10-10)
Would Android (AOSP), stripped of what we do not need including the UI (or not stripped in the first versions), suit as the base operating system of a machine that hosts all of the platform's infrastructure, instead of a classical Linux? Ordinary Linux and macOS would then be supported with weaker fault-tolerance guarantees. Also consider the opposite use: the same platform as special services for agents on Android devices themselves. First a conceptual analysis: how our design maps onto Android, and whether it works as intended.

## Read first
- The Swarma book: relux-works/wiki `book/swarma/ru/` (chapters 02–11), and `session-host/architecture.ru.md`.
- The module specifications: `relux-works/swarma-user-manager`, `swarma-credential-broker`, `swarma-dispatcher`, `swarma-session-host` (`spec/`).
- Our principles are: one OS user per agent; a key keeper holding hardware-backed keys; a broker and a dispatcher behind sockets that identify callers by kernel peer credentials; per-UID egress; kernel sandboxes; operator-only daemon updates; fences and epochs; session runners that survive service restarts.

## Answer
1. **A mapping table.** For each platform component and principle (user manager and launcher; dynamic per-agent UIDs and homes; key keeper; credential broker; dispatcher; session host and session runner with PTYs and app-servers; per-UID firewall and network profiles; sandboxes; supervision and restart; updates, rollback and verified boot; the remote-worker bridge; the Matrix carrier), give the Android mechanism that would carry it, for example:
   - per-app UIDs and isolated processes;
   - SELinux domains and seccomp;
   - Binder caller identity;
   - netd and eBPF per-UID rules;
   - Keystore, KeyMint, StrongBox and key attestation;
   - init services;
   - AVB/dm-verity, A/B updates, APEX and Mainline;
   - lmkd.

   Then state fit, gap or conflict, with sources: AOSP documentation and code, cited by URL and version.
2. **What breaks or costs a lot:**
   - runtime-created UIDs versus install-time package UIDs;
   - no sudo or root helper model;
   - bionic versus glibc for the harnesses and developer toolchains: Claude Code's runtime, the Codex binary, Go, Rust, Node, Python, git, compilers, containers;
   - PTY and terminal behaviour;
   - background execution limits;
   - SSH server and Matrix server hosting;
   - server-class hardware support (x86_64, ARM servers, Cuttlefish virtual devices);
   - the work of building and maintaining a custom AOSP distribution and its security updates.
3. **Alternatives with similar guarantees** on Linux: Bottlerocket, Flatcar or Fedora CoreOS, ChromiumOS/minijail, NixOS, Talos; GrapheneOS as hardened Android. Compare them on the same table: what of Android's guarantees each gives, at what cost.
4. **Android as an agent node:** agents and their services on a user's Android device. What it would take, and what it would be good for.
5. **Verdict:** base OS, agent node, reference design only, or not at all. Give the reasons, the main risks, and a small first experiment that would test the riskiest assumption cheaply, preferably on hosted CI runners and never on our working machines.

## Rules
Read-only research; no execution on this host. Deliver as a `.research/` file in the Change Request and attach it as `android-os-base.md`. Never edit `LOGBOOK.md`. Cite every claim. No secrets, personal paths or host names. Then `task-board handoff <TASK> --role researcher` and END YOUR TURN.
