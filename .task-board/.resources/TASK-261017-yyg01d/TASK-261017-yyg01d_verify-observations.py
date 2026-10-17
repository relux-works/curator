from pathlib import Path
import json
base=Path(__file__).resolve().parent
primary=(base/'sandbox-05/sandbox.log').read_text()
supplement=(base/'sandbox-06/sandbox.log').read_text()
def snap(log,label):
    text=log.split(label,1)[1].split('\n',1)[1]
    return json.JSONDecoder().raw_decode(text)[0]
def digest(snapshot,path):
    entry=snapshot[path]
    if 'link' in entry:
        assert entry['resolves']
        return snapshot[entry['link'].split('/home/',1)[1]]['sha256']
    return entry['sha256']
before=snap(primary,'E BEFORE')
partial=snap(primary,'E AFTER failed use')
recovered=snap(primary,'E AFTER recovery attempt')
takeover=snap(primary,'C native AFTER takeover')
claude='.claude/CLAUDE.md'
assert digest(before,claude)==digest(partial,claude)==digest(recovered,claude)
for path in ['.codex/AGENTS.md','.config/opencode/AGENTS.md','.pi/AGENTS.md']:
    assert digest(before,path)!=digest(partial,path)
    assert digest(before,path)==digest(recovered,path)
    assert digest(partial,path)==digest(takeover,path)
assert digest(partial,claude)!=digest(takeover,claude)
old=snap(supplement,'C selected BEFORE'); new=snap(supplement,'C selected AFTER')
assert 'skills/sample/private-note.md' in old and 'skills/sample/private-note.md' not in new
assert digest(old,'skills/sample/SKILL.md')!=digest(new,'skills/sample/SKILL.md')
old=snap(supplement,'C adapter BEFORE'); new=snap(supplement,'C adapter AFTER')
assert digest(old,'skills/sample/SKILL.md')==digest(new,'skills/sample/SKILL.md')
assert not snap(supplement,'C adapter canonical AFTER')
print('PASS: E Claude unchanged on refusal; 3/3 other roots changed, persisted, and restored by explicit retry; takeover converged.')
print('PASS: C selected canonical collision loses manual notes and replaces SKILL.md; native collision refuses with unchanged bytes and no canonical publication.')
