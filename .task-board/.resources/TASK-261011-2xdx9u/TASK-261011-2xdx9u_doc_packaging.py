import re
import subprocess
from pathlib import Path

root = Path.cwd()
paths = [Path(p) for p in subprocess.check_output(['git', 'diff', '--name-only'], text=True).splitlines()]
paths += [Path(p) for p in subprocess.check_output(['git', 'ls-files', '--others', '--exclude-standard'], text=True).splitlines()]
paths += [Path('/tmp/TASK-261011-2xdx9u_alignment.md')]
errors = []
link_count = table_count = 0

def anchor(text):
    return re.sub(r'[^\w\- ]', '', text.lower()).replace(' ', '-')

for path in paths:
    if path.suffix != '.md':
        errors.append(f'non-document candidate: {path.name}')
        continue
    text = path.read_text()
    if not text.endswith('\n'):
        errors.append(f'no terminal newline: {path.name}')
    if '\x00' in text or any(line.rstrip() != line for line in text.splitlines()):
        errors.append(f'NUL or trailing whitespace: {path.name}')
    if re.search(r'\]\(board-resource://|(?:^|[ `\"(])/(?:Users|home)/|\b(?:ivan|evgen)\b|chatgpt\.com/c/|claude\.ai/chat/', text, re.I):
        errors.append(f'bounded prohibited-reference pattern: {path.name}')
    if len(re.findall(r'^```', text, re.M)) % 2:
        errors.append(f'unbalanced fenced code: {path.name}')
    if path.name.startswith('CIP-'):
        for marker in ('## Operative disposition', '**Decided:**', '**Open:**', '**Superseded:**'):
            if marker not in text[:2500]:
                errors.append(f'missing operative marker {marker}: {path.name}')
    for label, target in re.findall(r'\[([^\]\n]+)\]\(([^)\n]+)\)', text):
        if '://' in target or target.startswith('mailto:'):
            continue
        target = target.strip('<>')
        target_path, _, fragment = target.partition('#')
        dest = path.parent / target_path if target_path else path
        if not dest.is_file():
            errors.append(f'broken local link: {path.name}: {target}')
            continue
        if fragment and dest.suffix == '.md':
            headings = [anchor(line.lstrip('#').strip()) for line in dest.read_text().splitlines() if re.match(r'^#+ ', line)]
            if fragment not in headings:
                errors.append(f'broken local anchor: {path.name}: {target}')
        link_count += 1
    # Table-shape checks cover the newly authored tables, not historic tables.
    if path.name in ('platform-contract-disposition.md', 'TASK-261011-2xdx9u_alignment.md'):
        width = None
        for line in text.splitlines():
            if line.startswith('|'):
                cells = len(re.sub(r'`[^`]*`', '', line).split('|')) - 2
                if width is None:
                    width = cells
                    table_count += 1
                elif cells != width:
                    errors.append(f'table width: {path.name}: {line[:70]}')
            else:
                width = None
assert not (root / '.research/TASK-261011-2xdx9u_alignment.md').exists(), 'alignment must remain board-only'
print(f'{len(paths)}/{len(paths)} Markdown files; {link_count} local links/anchors; {table_count} new tables; {len(errors)} findings')
for error in errors:
    print(error)
raise SystemExit(1 if errors else 0)
