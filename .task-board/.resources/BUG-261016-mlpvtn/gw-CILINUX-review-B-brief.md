# Review B brief: BUG-261016-mlpvtn, self-hosted Linux CI isolation

You decide accept or reject for the Change Request on BUG-261016-mlpvtn. Read the
producer brief (gw-CILINUX-brief.md) and the bug notes.

Check, with evidence:

1. Every job the LINUX_SELF_HOSTED expression can route to the self-hosted Linux
   runners sets job-private GOCACHE, GOMODCACHE and GOPATH before its first Go
   command, and setup-go caching cannot restore or save a shared per-user cache
   there. Walk every job; a missed job is a reject.
2. The concurrency group never cancels pushes to main or gate/**.
3. Fork pull requests still route to the hosted image; hosted macOS/Windows
   routing, Race routing, SPEC_PIN and PNPM_PIN contracts are unchanged.
4. The control plant is real: the new check fails on the pre-fix workflow for the
   intended reason and passes on the fixed one.
5. The remote gate ran on the self-hosted Linux runners (runner names in the run)
   and was green.
6. No host names, IPs, personal names or local paths in committed text.

Sweep the whole diff, not the first hit. Record the surface table in your verdict.
