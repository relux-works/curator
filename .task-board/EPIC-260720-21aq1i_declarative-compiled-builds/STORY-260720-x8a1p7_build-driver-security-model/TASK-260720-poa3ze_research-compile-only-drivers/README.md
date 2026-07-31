# Research compile-only build drivers

## Description
Analyze the existing protocol and both managers, then design a declarative build-driver contract that preserves the no-package-code-during-install invariant. The recommendation must be implementable without generic shell hooks or package-controlled argument arrays and must support a deterministic first Go driver.

## Scope
Read-only research across /Users/iv/Developer/ReluxWorks/curator-spec at origin/main, /Users/iv/Developer/ReluxWorks/curator at origin/main, and /Users/iv/Developer/Wildberries/cocoaskills at origin/main. Persist conclusions as a board outcome resource and, when broadly useful, under curator/.research/. Do not modify product/spec source files.

## Acceptance Criteria
The report proposes exact JSON examples and semantic validation rules; specifies fixed Go environment and command construction; defines build ordering, dry-run, rollback, cache keys and receipts; analyzes network/module behavior and cgo; identifies all affected protocol artifacts; classifies at least Go, Rust, Zig, Swift, C/C++, Java/Kotlin, .NET, Node/TypeScript, Deno, and Python; and states a clear v1 recommendation.
