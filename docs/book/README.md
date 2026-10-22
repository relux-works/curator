# Curator book tooling

The book is Russian first. Its prose follows the supplied book editorial
conventions, including the five status labels; other repository documentation
keeps its existing English conventions. The [outline](OUTLINE.ru.md) maps the complete chapters, worked CLI
scenarios, glossary and nine original figures. Technical behavior is bounded to the
October 2026 implementation and accepted design, as stated in the chapters.

## Tools

Run from the repository root. The toolchain uses installed binaries and system
fonts; it downloads no fonts or diagram includes.

| Tool | Purpose | Command | Output |
| --- | --- | --- | --- |
| Pandoc 3.x | Markdown, table of contents, status filter and HTML | `docs/book/build.sh` | Temporary HTML under `docs/book/.temp/`, removed after each build |
| WeasyPrint 69+ | Print layout using PT Serif, PT Sans and PT Mono | `docs/book/build.sh ru OUTPUT.pdf` | Default: `docs/book/curator-book.ru.pdf`; PDF is not committed |
| PlantUML and Graphviz | Render original diagram sources | `PLANTUML_LIMIT_SIZE=16384 plantuml -charset UTF-8 -tpng -o "$PWD/docs/book/diagrams" docs/book/src/*.puml` | Committed `docs/book/diagrams/*.png` |
| Bash, grep, sed, awk, shasum | Check references, source/render inventory and source hashes | `docs/book/check-figures.sh` | Diagnostic text and exit status |
| Go | Tooling contracts and narrowing controls | `env -u GIT_DIR -u GIT_WORK_TREE go test ./docs/book -count=1 -v` | Test output; capture under `.temp/` |
| Poppler | Verify PDF metadata and extracted text | `pdfinfo docs/book/curator-book.ru.pdf` and `pdftotext docs/book/curator-book.ru.pdf .temp/book.txt` | Metadata on stdout, extracted text under `.temp/` |

Update source hashes in [the manifest](src/MANIFEST.md) whenever a diagram source
changes, then rerender and run the checker. SHA-256 binds sources, not PNG
pixels: a successful inventory check cannot prove visual correctness or that a
stale render was regenerated. Inspect changed figures during review.

The build runs the figure checker before rendering. The Lua filter refuses
unknown complete status markers, including a second adjacent marker. Real
pandoc tests cover adjacent valid and multiword labels, later unknown labels,
and rejection before the PDF renderer can replace prior output. A bounded
control removes only the adjacent-marker split and requires the planted
unknown-second-marker build to expose that regression.
The tooling tests pair valid inventories with
missing-render, missing-source, missing-row, changed-hash and orphan negatives;
bounded controls verify selected gates fail for their intended witnesses.
Pandoc filter tests skip if pandoc is absent, so Go success alone does not replace
the real PDF build. The renderer stand-ins used by build-order tests establish
call ordering and prior-output preservation, not real rendering.

The CLI chapter contains all 72 static command paths from `commandRegistry`
in `cmd/curator/command_help.go`, grouped into a generated tree. It excludes
installable umbrella providers from that static inventory. Recheck that
registry when updating the chapter; provider-specific flags belong to the
installed provider help. Runtime exclusion observations are bounded to the
versions and surfaces documented in `docs/qualification/`; they do not certify
a complete strict project adapter.
