# Repository admission limits

Local and network repository admission share `proveRepository`. The manager's
raw-object policy implements profile §11.5 with the following expansion limits:

- `MaxObjects` bounds unique objects read by Git's batch object reader.
- `MaxObjectBytes` bounds each raw object before allocation.
- `MaxExpandedBytes` independently bounds the sum of cached unique raw-object
  bytes, the sum of emitted file-content bytes, and the canonical snapshot size.
  Every file path is charged, including aliases of the same blob or tree.
- `MaxFiles` bounds emitted files. `MaxTreeEntries` bounds all visited expanded
  tree entries, including directories and entries reached through tree aliases.
  Its default is 400,000, independently of the 200,000-file default.

File content and canonical framing are reserved before copying a blob or
appending its file record. Framing includes the frozen 24-byte header, each
record's type byte, two eight-byte lengths, UTF-8 path bytes, and content bytes.
Canonical storage is allocated once at its reserved size; its framing and digest
remain unchanged for repositories within the limits. The cache, copied contents,
and canonical storage therefore each have a byte bound, with file and entry
limits bounding metadata and the path-length/depth limits bounding path storage.

The walk checks cancellation at tree entry and before each child entry, including
cache hits. Exceeding a budget returns `build_repository_incomplete_source` from
both admission paths. Positive limit overrides are honored; omitted limits use
`DefaultLimits`.
