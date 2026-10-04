"""Task-scoped rc.3 document validator and regression tests; no repository edits."""
from pathlib import Path
import argparse
import hashlib
import json
import re
import subprocess
import unittest

TASK = 'TASK-261004-31fcvu'
RC2 = '## 0.15.0-rc.2 - 2026-09-26\n'
RC3 = '## 0.15.0-rc.3 - 2026-10-04\n'
HISTORICAL = '### Shipped in 0.15.0-rc.2 but not recorded in its notes\n'
EXPLANATION = ('These entries were listed under Unreleased when 0.15.0-rc.2 was tagged; '
               'the changes are part of 0.15.0-rc.2.')

def git(*args):
    return subprocess.check_output(['git', *args]).decode()

BASE = git('show', 'origin/main:CHANGELOG.md')
TAG = git('show', 'v0.15.0-rc.2:CHANGELOG.md')
BODY = BASE.split('## Unreleased\n', 1)[1].split(RC2, 1)[0]
BLOCKS = [b.rstrip() for b in re.findall(r'^- .*?(?=^-[ ]|^### |\Z)', BODY, re.M | re.S)]
# Five explicit consolidations required by the original brief, not deletion exceptions.
# Entry 2's obsolete rc.13 pin statement must be replaced rather than copied verbatim.
CONSOLIDATED = {
    1: ('Fixed', '- Marker v3/v4 readers', ('external repository identities', 'substitution kinds', 'effective revision widths')),
    2: ('Changed', '- CI conformance pin advances', ('1.0.0-rc.14', '43bf0a2506d5c354a73bbc3ea4623d4653db10c7',
                                               '6f832d813efc768ea154a7d5076b512ab4be6aa9409d92e11469d21ea9bc69f5')),
    3: ('Security', '- Windows executable resolution', ('platform ownership', 'every', 'component-store hard-link origin', 'captured System32 exception')),
    4: ('Fixed', '- Unix HTTPS askpass', ('secret only after accepting the password', 'broken-pipe', 'refusal paths')),
    21: ('Fixed', '- Project install materializes', ('non-git product roots', 'hygiene', 'unexpected Git failures')),
}
RESTORED = [(n, b) for n, b in enumerate(BLOCKS, 1) if n not in CONSOLIDATED]
INTERNAL_LABEL = re.search(r'outside the probed paths \((\S+) runs', BASE).group(1)


def rc3(text):
    return text.split(RC3, 1)[1].split(RC2, 1)[0]


def missing_original_entries(text, narrow_release_search=False):
    """Preservation gate: released-section twin or verbatim rc.3, never tag Unreleased.

    The five explicitly consolidated entries are validated separately below. No old
    historical entry may use those transformations, a partial match, or tag presence.
    """
    released = BASE[BASE.index(RC2):] + TAG[TAG.index(RC2):]
    if narrow_release_search:
        # Negative-evidence mutant: narrow refusal to entries absent from the whole tag.
        released = TAG
    return [n for n, b in RESTORED if b not in released and b not in rc3(text)]


def private_wording(text):
    body = rc3(text)
    if ARGS.narrow_public_scan:
        body = body.split(HISTORICAL, 1)[0]
    return INTERNAL_LABEL in body or bool(re.search(
        r'/Users/[^\s`]+|/home/[^\s`]+|https?://[^/\s]+\.local\b|\bmacmini[-\w]*', body, re.I))


def validate(text):
    prefix = '# Changelog\n\nAll notable implementation changes are recorded here.\n\n## Unreleased\n\n'
    assert text.startswith(prefix + RC3), 'Fresh empty Unreleased/date'
    assert text.count(RC3) == 1
    assert text[text.index(RC2):].encode() == BASE[BASE.index(RC2):].encode(), 'Earlier release bytes changed'
    section = rc3(text)
    assert not missing_original_entries(text), 'Missing original Unreleased entries: ' + str(missing_original_entries(text))
    assert re.findall(r'^### (.+)$', section, re.M) == [
        'Added', 'Changed', 'Fixed', 'Security', 'Known issues', HISTORICAL.strip()[4:]]
    expected = HISTORICAL + '\n' + EXPLANATION + '\n\n' + '\n\n'.join(b for n, b in RESTORED) + '\n\n'
    assert section.split(HISTORICAL, 1)[1] == expected.split(HISTORICAL, 1)[1], 'Historical entries changed or reordered'
    assert len(BLOCKS) == 32 and len(RESTORED) == 27
    tag_unreleased = TAG.split('## Unreleased\n', 1)[1].split(RC2, 1)[0]
    assert all(b in tag_unreleased for n, b in RESTORED)
    assert not any(b in BASE[BASE.index(RC2):] or b in TAG[TAG.index(RC2):] for n, b in RESTORED)
    fresh = section.split(HISTORICAL, 1)[0]
    for n, (group, lead, tokens) in CONSOLIDATED.items():
        group_body = fresh.split('### ' + group + '\n', 1)[1].split('\n### ', 1)[0]
        assert group_body.count(lead) == 1, f'Entry {n} lost or duplicated'
        entry = next(b for b in re.findall(r'^- .*?(?=^-[ ]|\Z)', group_body, re.M | re.S) if b.startswith(lead))
        assert all(token in ' '.join(entry.split()) for token in tokens), f'Entry {n} incomplete consolidation'
    assert 'keep rc.13 pinned' not in section
    for token in ['curator#100','curator-spec#121','launch-env-fragment-v3','launch-env-fragment-v2',
                  'Decision 0018','env migrate','env unmanage --restore-backups','curator-content-v1',
                  'v2 writer remains off','Codex seed revision A','security_posture_permissive',
                  'approximately 21,000','No root cause','deferred to rc.4','TASK-261003-1uzji7',
                  'atomic v1→v2','N1–N4','not fixed in rc.3','issues/106','STORY-261004-3oognx',
                  'B3 cache-prune PRs are excluded','explicit self-hosted','runner label']:
        assert token in fresh, token
    assert not private_wording(text), 'Internal runner identifier or private path in rc.3'
    ci = git('show', 'origin/main:.github/workflows/ci.yml')
    for token in CONSOLIDATED[2][2][1:]:
        assert token in ci
    assert re.search(r'EnableV2Writers\s*=\s*false', git('show', 'origin/main:internal/hashing/hashing.go'))
    git('cat-file', '-e', 'origin/main:docs/security-audit-2026-10-inline.md')
    assert git('rev-parse', 'HEAD') == git('rev-parse', 'origin/main')
    assert git('diff', '--name-only').splitlines() == ['CHANGELOG.md']
    assert git('diff', '--cached', '--name-only') == ''
    assert git('diff', 'HEAD', '--', 'LOGBOOK.md') == ''
    assert git('status', '--porcelain', '--untracked-files=all').splitlines() == [' M CHANGELOG.md']
    ledger = json.loads(Path(__file__).with_name(TASK + '_map.json').read_text())
    subjects = dict(row.split('\t', 1) for row in git('log', 'v0.15.0-rc.2..origin/main', '--format=%H\t%s').splitlines())
    assert set(subjects) == {r[0] for r in ledger['rows'] + ledger['administrative']}
    assert ledger['base'] == git('rev-parse', 'origin/main').strip()
    for oid, subject, *mapping in ledger['rows'] + ledger['administrative']:
        assert subject == subjects[oid].replace(INTERNAL_LABEL, '[internal runner identifier omitted]')
        assert ledger['subject_sha256'][oid] == hashlib.sha256(subjects[oid].encode()).hexdigest()
    for oid, subject in ledger['administrative']:
        paths = git('diff-tree', '--no-commit-id', '--name-only', '-r', oid).splitlines()
        assert paths and all(p.startswith('.task-board/') for p in paths), 'Non-board administrative commit'
    for n, (row, block) in enumerate(zip(ledger['unreleased'], BLOCKS), 1):
        assert row['entry'] == n and row['text'] == block
        assert row['action'] == ('represented' if n in CONSOLIDATED else 'restored_verbatim')
    assert len(ledger['unreleased']) == 32
    print(f"PASS: {len(ledger['rows'])}/{len(ledger['rows'])} subject mappings; {len(ledger['administrative'])}/{len(ledger['administrative'])} board-only commits; 32/32 entries (27 verbatim rc.2 carryovers, 5 explicit consolidations); release bytes, pin, v1 writes, known issues, public wording and single-file scope.")


class ReleaseNotesRegression(unittest.TestCase):
    def test_prior_unreleased_requires_released_twin_or_verbatim_rc3(self):
        """P1: each of the 27 deletions must be rejected even if present in tag Unreleased."""
        good = Path('CHANGELOG.md').read_text()
        self.assertEqual(missing_original_entries(good), [])
        admitted = []
        for n, block in RESTORED:
            candidate = good.replace(block + '\n\n', '', 1)
            if missing_original_entries(candidate, ARGS.narrow_release_search) != [n]:
                admitted.append(n)
        self.assertEqual(admitted, [], 'Preservation gate admitted omitted entries')
        print('P1 preservation regression: 27/27 single-entry omission mutants rejected')

    def test_prior_unreleased_in_ambient_unreleased_is_not_released(self):
        good = Path('CHANGELOG.md').read_text()
        n, block = RESTORED[0]
        candidate = good.replace(block + '\n\n', '', 1).replace('## Unreleased\n\n', '## Unreleased\n\n' + block + '\n\n', 1)
        self.assertEqual(missing_original_entries(candidate, ARGS.narrow_release_search), [n])

    def test_internal_runner_identifier_rejected_in_rc3(self):
        """P2: reject internal labels in both current prose and the restored historical subsection."""
        good = Path('CHANGELOG.md').read_text()
        self.assertFalse(private_wording(good))
        injected = good.replace(HISTORICAL, HISTORICAL + '\n' + INTERNAL_LABEL + '\n', 1)
        self.assertTrue(private_wording(injected))
        original = good.replace('select the explicit self-hosted\n  runner label',
                                'use the explicit ' + INTERNAL_LABEL + '\n  runner label', 1)
        self.assertTrue(private_wording(original))


if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('--candidate', default='CHANGELOG.md')
    parser.add_argument('--regression', action='store_true')
    parser.add_argument('--narrow-release-search', action='store_true')
    parser.add_argument('--narrow-public-scan', action='store_true')
    ARGS = parser.parse_args()
    if ARGS.regression:
        unittest.main(argv=['rc3-regression'], verbosity=2)
    else:
        validate(Path(ARGS.candidate).read_text())
