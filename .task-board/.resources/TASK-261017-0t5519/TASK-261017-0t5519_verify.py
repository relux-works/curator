"""Check derived timing tables against the original nine hosted event streams."""
import csv
import json
import sys
from datetime import datetime
from pathlib import Path

ROOT = Path(__file__).parent

def read_events(path):
    return [json.loads(line) for line in path.read_text().splitlines()]

def validate_lifecycle(events):
    starts = {(e['Package'], e.get('Test', '')) for e in events if e['Action'] in ('start', 'run')}
    terminals = {(e['Package'], e.get('Test', '')) for e in events if e['Action'] in ('pass', 'skip', 'fail')}
    assert starts == terminals, 'missing_terminal'
    assert not any(e['Action'] == 'fail' for e in events), 'failed_event'

def validate_csv(rows, events, run, platform):
    prefix = f'{run}_{platform}_'
    indexed = {(e['Package'].removeprefix('github.com/relux-works/curator/'), e.get('Test', '')): e for e in events if e['Action'] in ('pass', 'skip', 'fail')}
    starts = {(e['Package'].removeprefix('github.com/relux-works/curator/'), e.get('Test', '')): e for e in events if e['Action'] in ('start', 'run')}
    found = set()
    for row in rows:
        key = row['package'], row['test']
        event = indexed.get(key)
        if event is None:
            assert row[prefix+'status'] == 'absent', 'false_presence'
        else:
            found.add(key)
            assert row[prefix+'status'] == event['Action'], 'status_mismatch'
            assert float(row[prefix+'elapsed_s']) == event.get('Elapsed', 0), 'elapsed_mismatch'
            parse = lambda s: datetime.fromisoformat(s.replace('Z', '+00:00'))
            wall = (parse(event['Time']) - parse(starts[key]['Time'])).total_seconds()
            assert abs(float(row[prefix+'wall_s']) - wall) < 0.000002, 'wall_mismatch'
    assert found == set(indexed), 'missing_csv_row'

rows = []
for category in ['packages', 'tests']:
    with (ROOT / f'TASK-261017-0t5519_{category}.csv').open() as file:
        rows.extend(csv.DictReader(file))

events_by_lane = {}
for item in json.loads((ROOT/'TASK-261017-0t5519_evidence-manifest.json').read_text()):
    run, platform = item['run'], item['platform']
    events = read_events(ROOT/f'run-{run}'/f'test-evidence-{platform}'/'test/go-test.json')
    validate_lifecycle(events)
    validate_csv(rows, events, run, platform)
    events_by_lane[run, platform] = events

mode = sys.argv[1] if len(sys.argv)>1 else 'green'
if mode == 'missing-terminal':
    # A truncated stream must not be accepted as a complete timing sample.
    events = events_by_lane[39042501825, 'windows-latest']
    validate_lifecycle([e for e in events if not (e['Action']=='pass' and e.get('Package','').endswith('/cmd/curator') and not e.get('Test'))])
elif mode == 'wrong-timing':
    # A fabricated CSV duration must disagree with the independent hosted terminal event.
    changed = [dict(row) for row in rows]
    row = next(row for row in changed if row['package']=='cmd/curator' and not row['test'])
    row['39042501825_windows-latest_elapsed_s'] = '1.0'
    validate_csv(changed, events_by_lane[39042501825, 'windows-latest'], 39042501825, 'windows-latest')
else:
    print(f'PASS: {len(events_by_lane)} complete green event streams; {len(rows)} table rows reconciled with original statuses and elapsed values.')
