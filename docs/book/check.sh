#!/bin/sh
# Hygiene and link check for the Curator book chapters (docs/book/ru/*.md).
#
#   sh docs/book/check.sh
#
# Adapted from the Swarma book check (scripts/check.sh in swarma-docs). Fails
# on:
#  - status markers ⟦...⟧ (also split or unmatched) and delivery hedges outside
#    the state chapter 11-gde-my.md;
#  - the prose blacklist of the editorial guide: a dash used as punctuation
#    (any em dash, and any en dash except in a numeric range whose left side
#    is a digit: 0–2, L0–L2), guillemets, antithesis, filler openers and
#    promotional adjectives;
#  - internal identifiers (TASK-, BUG-, STORY-, EPIC-, RUN-, "Decision NNNN",
#    D-NNN, PR #N) and local paths (/Users/, /home/, ~/);
#  - a chapter without a level-1 heading;
#  - broken links: a cross-reference such as глава "X", главы "X", главе "X",
#    главу "X", главой "X" or the plural главах "X", "Y" и "Z" whose title is
#    not the `# ` title of a chapter; a ../diagrams/ link to a missing file; a
#    relative Markdown link that does not resolve. Fenced code blocks are
#    skipped by the link check, and a reference may wrap across lines.
# Prints every hit with its file and line.
#
# A hedge is "пока", "ещё не" (also written "еще не"), "в v0" or
# "не реализован..."; other wordings of the same meaning pass. The checks are
# lexical, so the final judgement stays with the reviewer. The checker pins a
# UTF-8 locale whatever the caller's environment, and refuses to run without
# one. BOOK_CHECK_ROOT points it at another book directory (used by tests).
set -u

utf8=
for loc in C.UTF-8 en_US.UTF-8 C.utf8 en_US.utf8; do
  if printf 'Пока ё\n' | LC_ALL=$loc grep -qiE '^пока [ё]$' 2>/dev/null; then
    utf8=$loc
    break
  fi
done
[ -n "$utf8" ] || { echo "check.sh: no UTF-8 locale available (tried C.UTF-8, en_US.UTF-8); refusing" >&2; exit 2; }
LC_ALL=$utf8
export LC_ALL

root="${BOOK_CHECK_ROOT:-$(cd "$(dirname "$0")" && pwd)}"
cd "$root" || exit 2
state="ru/11-gde-my.md"

files=$(find ru -type f -name '*.md' 2>/dev/null | sort)
[ -n "$files" ] || { echo "check.sh: no chapters in $root/ru"; exit 2; }
other=$(printf '%s\n' "$files" | grep -vxF "$state")

fail=0
hit() {
  # $1 label, $2 extended regex, $3 extra grep flags ("" or i), $4 files
  [ -n "$4" ] || return 0
  out=$(grep -Hn${3}E -- "$2" $4 2>/dev/null)
  if [ -n "$out" ]; then
    echo "== $1"
    printf '%s\n' "$out"
    fail=1
  fi
}

# Status and delivery state belong to the state chapter only.
hit "status marker" '[⟦⟧]' "" "$other"
hit "delivery hedge" '(^|[^[:alpha:]])(пока|[её]щ[её][[:space:]]+не|в[[:space:]]+v0|не[[:space:]]+реализован[[:alpha:]]*)($|[^[:alpha:]])' i "$other"

# Prose blacklist (editorial guide, "Чёрный список").
hit "dash in prose" '—|(^|[^0-9])–|–($|[^[:alnum:]])' "" "$files"
hit "guillemets" '[«»]' "" "$files"
hit "antithesis" '(не[[:space:]]+просто[[:space:]].*,[[:space:]]+а[[:space:]])|(это[[:space:]]+не[[:space:]].*,[[:space:]]+а[[:space:]])|(речь[[:space:]]+не[[:space:]]+о[[:space:]])' i "$files"
hit "filler opener" '(стоит[[:space:]]+отметить|важно[[:space:]]+понимать|давайте[[:space:]]+разбер|в[[:space:]]+современном[[:space:]]+мире)' i "$files"
hit "marketing adjective" '(^|[^[:alpha:]])(мощн[[:alpha:]]*|бесшовн[[:alpha:]]*)($|[^[:alpha:]])' i "$files"

# Internal identifiers and local paths.
hit "board, run or review identifier" '(^|[^A-Za-z0-9])(TASK|BUG|STORY|EPIC|RUN)-[0-9A-Za-z]' "" "$files"
hit "decision number" '(Decision[[:space:]]+[0-9]{2,}|(^|[^A-Za-z0-9-])D-[0-9]{3}([^0-9]|$))' "" "$files"
hit "pull request number" 'PR[[:space:]]*#[0-9]+' "" "$files"
hit "local path" '/Users/|/home/[A-Za-z0-9._-]+|~/' "" "$files"

# Every chapter starts with one level-1 heading (after front matter in 00).
for f in ru/[0-9][0-9]*.md; do
  [ -e "$f" ] || continue
  if ! grep -qE '^# ' "$f"; then
    echo "== chapter without a level-1 heading"
    echo "$f"
    fail=1
  fi
done

# Links: chapter cross-references, figures and relative Markdown links.
out=$(perl -CSDA -Mutf8 -e '
  use strict; use warnings;
  my @files = @ARGV;
  my (%titles, %text);
  for my $f (@files) {
    open my $fh, "<:encoding(UTF-8)", $f or die "$f: $!";
    my $fence = 0; my @lines;
    while (my $l = <$fh>) {
      if ($l =~ /^\s*(```|~~~)/) { $fence = !$fence; push @lines, "\n"; next; }
      if ($fence) { push @lines, "\n"; next; }
      push @lines, $l;
      if ($l =~ /^# (.+?)\s*$/) { (my $t = $1) =~ s/\s+/ /g; $titles{$t} = 1; }
    }
    $text{$f} = join "", @lines;
  }
  my $q = qr/"([^"]+)"/;
  for my $f (@files) {
    my $s = $text{$f};
    my $line = sub { 1 + (substr($s, 0, $_[0]) =~ tr/\n//) };
    while ($s =~ /(?<![\p{L}])[Гг]лав(?:а|ы|е|у|ой|ах|ам|ами)\s+("[^"]+"(?:\s*(?:,|\bи\b)\s*"[^"]+")*)/g) {
      my ($list, $pos) = ($1, $-[0]);
      while ($list =~ /$q/g) {
        (my $t = $1) =~ s/\s+/ /g;
        print "chapter reference\t$f:" . $line->($pos) . ": no chapter titled \"$t\"\n" unless $titles{$t};
      }
    }
    while ($s =~ /\]\(([^)\s]+)(?:\s+"[^"]*")?\)/g) {
      my ($target, $pos) = ($1, $-[0]);
      next if $target =~ m{^(?:[a-z][a-z0-9+.-]*:|#)}i;
      (my $path = $target) =~ s/#.*//;
      next if $path eq "";
      (my $dir = $f) =~ s{/[^/]*$}{};
      next if -e "$dir/$path";
      my $kind = $path =~ m{^\.\./diagrams/} ? "missing figure" : "broken link";
      print "$kind\t$f:" . $line->($pos) . ": $target\n";
    }
  }
' $files 2>&1)
if [ -n "$out" ]; then
  printf '%s\n' "$out" | sort -t"$(printf '\t')" -k1,1 -s | awk -F'\t' '
    $1 != last { print "== " $1; last = $1 }
    { print $2 }'
  fail=1
fi

[ "$fail" -eq 0 ] && echo "check.sh: ok ($(printf '%s\n' "$files" | wc -l | tr -d ' ') files)"
exit "$fail"
