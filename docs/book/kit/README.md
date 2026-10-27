# Vendored diagram kit

These files are byte-identical copies of the Swarma book diagram kit. Do not
edit them here: a refresh is a plain copy from the source repository followed
by an update of the table below, `check-figures.sh` and a review of every
rerendered figure.

- Source repository: `swarma-docs`
- Source path: `scripts/diagram-kit/`
- Source commit: `c423a055f83c88ea12f0f0da5f3c031ac07b30e9`
- Style rules: `editorial/diagram-style.ru.md` in the same repository

| File | Purpose | SHA-256 |
| --- | --- | --- |
| `render.py` | TOML diagram source to SVG, geometry lint, `--report` and `--require-fit` | `aa7526f5f0d44509b1380810f105555b529a21d3b24e7ba6d1be41c95161d87c` |
| `theme.toml` | Role colours, fonts, sizes, spacing and print budget | `f7b082e1ed86100507e13481c3d3895460d5170c791799a6b627d2036f03facd` |
| `metrics.json` | PT Sans and PT Mono advance widths used to measure text | `91374cb8142d625a45280e050e0531b3f54e7e089d3266e9d4983040e6a6fe82` |

`render.py` loads `theme.toml` and `metrics.json` from its own directory and
needs nothing else at runtime besides Python 3.11+ (`tomllib`). The `graph`
layout needs Graphviz `dot` at the version pinned in `GRAPHVIZ_PIN`
(14.0.4); `--png` needs `rsvg-convert`. The source repository's kit tests,
fixtures, probes and metrics generator are not vendored.

Verify the copies:

```sh
shasum -a 256 docs/book/kit/render.py docs/book/kit/theme.toml docs/book/kit/metrics.json
```
