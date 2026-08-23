# TASK-260811-3ksxig pnpm patch/layout evidence

Date: 2026-08-23

## Pinned upstream evidence

- pnpm tag inspected: `v10.33.0`.
- `lockfile/settings-checker/src/calcPatchHashes.ts` calls
  `createHexHashFromFile` for every declared patch.
- `crypto/hash/src/index.ts` defines that operation as SHA-256 hex over the
  UTF-8 file text after replacing CRLF with LF.
- `packages/dependency-path/src/index.ts` defines virtual-store filenames by
  replacing unsafe path characters, converting peer/patch parentheses to
  underscores, and using a 32-hex SHA-256 prefix when the 120-character limit
  or mixed-case rule requires hashing.
- `patching/apply-patch/src/index.ts` applies the declared patch to the package
  directory. Curator independently derives an expected inventory with a closed
  unified-diff transform and receipts admitted tarball + admitted patch ->
  expected files before materialization.

Downloaded source probes are stored beside this note:

- `calcPatchHashes.ts`
- `pnpm-crypto-hash.ts`
- `dependency-path.ts`
- `applyPatch.ts`
- upstream fixture lock and patch

## Tool anomaly

The first GitHub tree request failed with exit 1 because unquoted `?recursive=1`
was expanded by zsh (`no matches found`). The retry quoted the URL and exited 0.

