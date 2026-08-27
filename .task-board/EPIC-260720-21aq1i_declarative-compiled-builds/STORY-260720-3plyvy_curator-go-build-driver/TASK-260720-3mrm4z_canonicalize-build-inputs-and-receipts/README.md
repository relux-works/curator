# Canonicalize build inputs and receipts

## Description
Define Curator internal models for the exact go-v1 logical input, target, toolchain, fixed policy, artifact, receipt, cache key, and receipt hash. Provide strict CCJ-1 encoding and decoding independently of filesystem cache behavior.

## Scope
Extend internal/protocoljson with reusable strict CCJ-1 canonical encoding and byte-equality support, migrate internal/registry to consume it without changing registry vectors, and create internal/buildmeta for the logical go-v1 input, target, toolchain, fixed policy, artifact, receipt, cache-key, and receipt-hash models and codecs. Store no absolute paths or timestamps. Readers reject duplicate keys, non-integer numbers, unknown fields, unsupported versions, noncanonical whitespace, a BOM, and a trailing newline; writers emit exact CCJ-1 bytes. internal/buildcache filesystem ownership is reserved for TASK-260720-3pwg2w. This task does not inspect ownership, publish files, run Go, or write install markers.

## Acceptance Criteria
The authoritative cache-input canonical bytes derive the exact rc.4 cache key and the canonical receipt derives its exact receipt_sha256; the full expected input is compared, including build_root, source_dir, source identity, target tuning, toolchain identity, and directive policy; artifact path, size, and hash constraints are strict and platform-derived; pretty-printed or otherwise noncanonical receipt bytes fail even when their parsed values match; receipt hashes are documented and tested as consistency metadata rather than provenance; existing registry canonicalization vectors remain green.
