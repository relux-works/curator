# Curator book tooling

The book is Russian first. Its prose follows the supplied book editorial
conventions, including the five status labels; other repository documentation
keeps its existing English conventions. The [outline](OUTLINE.ru.md) maps the complete chapters, worked CLI
scenarios, glossary and eight original figures. Technical behavior is bounded to the
October 2026 implementation and accepted design, as stated in the chapters.

## Tools

Run from the repository root. The toolchain uses installed binaries and system
fonts; it downloads no fonts or diagram includes.

| Tool | Purpose | Command | Output |
| --- | --- | --- | --- |
| POSIX sh, grep, find, perl | Chapter hygiene and link check: status marks and delivery hedges outside `ru/11-gde-my.md`, dashes as punctuation, guillemets, the editorial blacklist, internal ids, local paths, chapter cross-references, figure links and relative Markdown links | `sh docs/book/check.sh` | One `== rule` section per hit with `file:line`; exit 1 on any hit |
| Diagram kit (Python 3.11+, `docs/book/kit/render.py`) | Render a TOML figure source to SVG in the shared book style and check its fit | `python3 docs/book/kit/render.py docs/book/src/NAME.toml docs/book/diagrams/NAME.svg --report --require-fit tall` (`normal` for a figure linked without `{.tall}`) | Committed `docs/book/diagrams/NAME.svg`; the report line on stdout |
| Graphviz 14.0.4 | Node placement for the kit's `graph` layout; the kit refuses any other version | used by the kit | none of its own |
| rsvg-convert | Optional PNG preview of a figure | add `--png .temp/NAME.png` to the kit command | PNG under `.temp/`, not committed |
| Bash, grep, sed, awk, shasum, cmp | Figure inventory and freshness: `--structural` checks links, sources, renders, manifest rows, box classes and copy hashes; `--render` rerenders every source and compares bytes and fit | `bash docs/book/check-figures.sh` (both passes), `--structural` or `--render` | Diagnostic text and exit status |
| Pandoc 3.x | Markdown, table of contents, book filter (status badges, landscape figures, Russian no-break spaces) and HTML | `docs/book/build.sh` | Temporary HTML in a private directory under `TMPDIR`, removed after each build |
| WeasyPrint 69+ | Print layout using PT Serif, PT Sans and PT Mono | `docs/book/build.sh ru OUTPUT.pdf` | Default: `docs/book/curator-book.ru.pdf`; PDF is not committed |
| Go | Tooling contracts, negative fixtures and narrowing controls | `env -u GIT_DIR -u GIT_WORK_TREE go test ./docs/book -count=1 -v` | Test output; capture under `.temp/` |
| Poppler | Verify PDF metadata and extracted text | `pdfinfo docs/book/curator-book.ru.pdf` and `pdftotext docs/book/curator-book.ru.pdf .temp/book.txt` | Metadata on stdout, extracted text under `.temp/` |

`build.sh` runs `check.sh`, then the structural figure pass, then the render
pass, and only then pandoc and WeasyPrint. `BOOK_SKIP_FIGURE_RENDER=1` skips
only the render pass for a draft and prints a warning. An image WeasyPrint
fails to load fails the build. The PDF replaces the output file only after a
clean run, so every refusal leaves the prior PDF in place.

The diagram kit in [`kit/`](kit/README.md) is a byte-identical copy of the
Swarma book kit; its README records the source repository, path, commit and
the SHA-256 of each file, and a refresh is a plain copy. Figure sources follow
the Swarma diagram style rules (`editorial/diagram-style.ru.md` in
`swarma-docs`): sources name roles, the theme sets colours and sizes.
[The figure manifest](src/MANIFEST.md) lists every figure with its origin, the
SHA-256 of an upstream copy if any, its box (`tall` or `обычное`, matching the
`{.tall}` class of its chapter links) and its chapters. The render pass proves
that each committed SVG is the kit's current output for its source and that
its smallest text stays at least 7 pt in its box; it does not judge the
content, so inspect changed figures during review.

The Go tests pair valid inventories with missing-source, missing-render,
missing-row, box-class-mismatch, mixed-class, orphan, source-without-render
and changed-copy negatives for the structural pass; fresh, stale, wrong-box
and wrong-Graphviz cases for the render pass; a clean fixture and one planted
violation per rule for `check.sh`; and gate-order cases for `build.sh` with
renderer stand-ins. Bounded mutants remove one gate at a time and must be
caught by their witness. Tests that need pandoc, python3 3.11+ or Graphviz
14.0.4 skip with the reason when the tool is absent, so Go success alone does
not replace the real PDF build. The renderer stand-ins establish call order
and prior-output preservation, not real rendering.

The CLI chapter contains all 87 static command paths from `commandRegistry`
in `cmd/curator/command_help.go`, grouped into a generated tree. It excludes
installable umbrella providers from that static inventory. Recheck that
registry when updating the chapter; provider-specific flags belong to the
installed provider help. Runtime exclusion observations are bounded to the
versions and surfaces documented in `docs/qualification/`; they do not certify
a complete strict project adapter.
