#!/usr/bin/env bash
# Build the Curator book with pandoc 3.x and WeasyPrint 69+.
# Usage: docs/book/build.sh [ru [OUTPUT.pdf]]
set -euo pipefail
here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
lang="${1:-ru}"
src="$here/$lang"
out="${2:-$here/curator-book.$lang.pdf}"
for tool in pandoc weasyprint; do
  command -v "$tool" >/dev/null || { echo "build.sh: $tool not found" >&2; exit 1; }
done
shopt -s nullglob
chapters=("$src"/[0-9][0-9]-*.md)
if (( ${#chapters[@]} == 0 )); then
  echo "build.sh: no chapters in $src" >&2
  exit 1
fi
"$here/check-figures.sh"
# Keep temporary HTML within the book tree, independently of caller TMPDIR.
mkdir -p "$here/.temp"
tmp="$(mktemp -d "$here/.temp/build.XXXXXX")"
trap 'rm -rf "$tmp"' EXIT
pandoc "${chapters[@]}" \
  --from markdown --to html5 --standalone \
  --template "$here/template.html" --lua-filter "$here/filters/book.lua" \
  --toc --toc-depth=2 --section-divs --syntax-highlighting=none \
  --number-sections --number-offset=-1 --metadata lang="$lang" \
  --output "$tmp/book.html"
weasyprint --base-url "$src/" --stylesheet "$here/style.css" "$tmp/book.html" "$out"
echo "$out"
