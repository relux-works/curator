import collections
import csv
import datetime as dt
import hashlib
import json
from pathlib import Path
import statistics

ROOT = Path(__file__).parent
RUNS = [39029649557, 39029714957, 39042501825]
PLATFORMS = ['windows-latest', 'macos-latest', 'ubuntu-latest']
PREFIX = 'github.com/relux-works/curator/'

def timestamp(s):
    return dt.datetime.fromisoformat(s.replace('Z', '+00:00'))

data = {}
manifest = []
summary = []
for run in RUNS:
    metadata = json.loads((ROOT / f'run-{run}.json').read_text())
    artifacts = json.loads((ROOT / f'artifacts-{run}.json').read_text())['artifacts']
    for platform in PLATFORMS:
        job = next(j for j in metadata['jobs'] if j['name'] == f'Test ({platform})')
        assert job['conclusion'] == 'success'
        artifact_name = f'test-evidence-{platform}'
        artifact = next(a for a in artifacts if a['name'] == artifact_name)
        directory = ROOT / f'run-{run}' / artifact_name / 'test'
        source = directory / 'go-test.json'
        payload = source.read_bytes()
        events = [json.loads(line) for line in payload.splitlines() if line.startswith(b'{')]
        starts = {}
        terminals = {}
        pauses = set()
        conts = {}
        actions = collections.Counter(e['Action'] for e in events)
        assert actions['fail'] == 0
        assert not (directory / 'plan-deferred.txt').read_text().strip()
        for event in events:
            key = (event['Package'].removeprefix(PREFIX), event.get('Test', ''))
            when = timestamp(event['Time'])
            if event['Action'] in ('start', 'run'):
                assert key not in starts, key
                starts[key] = when
            elif event['Action'] == 'pause':
                pauses.add(key)
            elif event['Action'] == 'cont':
                conts[key] = when
            elif event['Action'] in ('pass', 'skip', 'fail'):
                assert key in starts and key not in terminals, key
                terminals[key] = dict(elapsed=event.get('Elapsed', 0),
                                      wall=(when-starts[key]).total_seconds(),
                                      start=event['Time'] if key not in starts else starts[key].isoformat(),
                                      end=event['Time'], action=event['Action'],
                                      parallel=key in pauses,
                                      cont=conts.get(key).isoformat() if key in conts else None)
        assert set(starts) == set(terminals), set(starts)-set(terminals)
        data[run, platform] = terminals
        packages = {k[0]: v for k, v in terminals.items() if not k[1]}
        top = {k: v for k, v in terminals.items() if k[1] and '/' not in k[1]}
        cli = {k: v for k, v in top.items() if k[0] == 'cmd/curator'}
        steps = []
        for step in job['steps']:
            if step['startedAt'] and step['completedAt']:
                steps.append(dict(name=step['name'], seconds=(timestamp(step['completedAt'])-timestamp(step['startedAt'])).total_seconds(), conclusion=step['conclusion']))
        summary.append(dict(run=run, platform=platform, sha=metadata['headSha'], job=job['databaseId'],
            seconds=(timestamp(job['completedAt'])-timestamp(job['startedAt'])).total_seconds(),
            steps=steps, counts=dict(actions), packages=len(packages), top_tests=len(top),
            cli_serial=sum(v['elapsed'] for v in cli.values() if not v['parallel']),
            cli_parallel=sum(v['elapsed'] for v in cli.values() if v['parallel']),
            cli_serial_count=sum(not v['parallel'] for v in cli.values()),
            cli_parallel_count=sum(v['parallel'] for v in cli.values()),
            cli_first_cont=min(v['cont'] for v in cli.values() if v['cont']),
            cli_package=packages['cmd/curator'],
            event_span=(max(timestamp(e['Time']) for e in events)-min(timestamp(e['Time']) for e in events)).total_seconds()))
        manifest.append(dict(run=run, platform=platform, sha=metadata['headSha'],job=job['databaseId'],
             artifact_id=artifact['id'], artifact_name=artifact_name, stream_sha256=hashlib.sha256(payload).hexdigest(), stream_bytes=len(payload),
             stream_lines=len(payload.splitlines()),parsed_events=len(events),non_json_lines=len(payload.splitlines())-len(events)))

columns = [(r,p) for r in RUNS for p in PLATFORMS]
keys = sorted(set().union(*(d.keys() for d in data.values())))
for category in ['packages','tests']:
    with (ROOT / f'TASK-261017-0t5519_{category}.csv').open('w', newline='') as f:
        writer = csv.writer(f)
        writer.writerow(['package','test']+[f'{r}_{p}_{metric}' for r,p in columns for metric in ['status','elapsed_s','wall_s']])
        for key in keys:
            if bool(key[1]) != (category=='tests'):
                continue
            cells = []
            for column in columns:
                value = data[column].get(key)
                cells.extend([value['action'],value['elapsed'],round(value['wall'],6)] if value else ['absent','',''])
            writer.writerow(list(key)+cells)
(ROOT/'summary.json').write_text(json.dumps(summary,indent=2))
(ROOT/'TASK-261017-0t5519_evidence-manifest.json').write_text(json.dumps(manifest,indent=2))
(ROOT/'timing-data.json').write_text(json.dumps({f'{r}/{p}': [{'package':k[0],'test':k[1],**v} for k,v in values.items()] for (r,p),values in data.items()}))
for row in summary:
    print(row['run'],row['platform'], 'job',row['seconds'], 'stream',round(row['event_span'],3), 'cli',row['cli_package']['elapsed'], 'serial',round(row['cli_serial'],3),'parallel',round(row['cli_parallel'],3),'counts',row['cli_serial_count'],row['cli_parallel_count'])
    print([(s['name'],s['seconds']) for s in row['steps'] if s['seconds']>=10])
print('All nine streams: paired starts/terminals, no fail actions, no deferred packages.')
