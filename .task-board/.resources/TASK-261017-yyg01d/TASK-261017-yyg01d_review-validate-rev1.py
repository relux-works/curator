from pathlib import Path
import json, re, subprocess
b=Path(__file__).resolve().parent
s=(b/'sandbox-05/sandbox.log').read_text(); t=(b/'sandbox-06/sandbox.log').read_text()
def snap(text,label):
 return json.JSONDecoder().raw_decode(text.split(label,1)[1].split('\n',1)[1])[0]
def result_rows(text):
 return [(block.splitlines()[0],int(re.search(r'^exit=(\d+)$',block,re.M)[1])) for block in text.split('$ ')[1:]]
rows=result_rows(s)+result_rows(t)
assert len(rows)==36, len(rows)
assert sum(code!=0 for cmd,code in rows)==4
assert s.count('HOME CHECK: private, mode 0700')==len(result_rows(s))
assert t.count('HOME CHECK: private, mode 0700')==len(result_rows(t))
a=snap(s,'A BEFORE managed'); after=snap(s,'A AFTER managed')
for path in ['skills/unmanaged-skill/SKILL.md','skills/unmanaged-notes/notes.md','skills/loose.txt']:
 assert path in a and path not in after
assert 'skills/sample/SKILL.md' in after
assert not snap(s,'A AFTER native')['skills/manual-link']['resolves']
print('PASS 36 CLI/git commands each guarded by private-HOME assertion; four intended nonzero exits; A three deletions and dangling link confirmed.')
# Control plant: make the captured native refusal lose its bytes. The observation verifier must reject it.
p=b/'control-plant'; (p/'sandbox-05').mkdir(parents=True); (p/'sandbox-06').mkdir()
(p/'sandbox-05/sandbox.log').write_text(s)
head,tail=t.split('C adapter AFTER',1)
tail=re.sub(r'"sha256": "[a-f0-9]{64}"','"sha256": "'+'0'*64+'"',tail,count=1)
(p/'sandbox-06/sandbox.log').write_text(head+'C adapter AFTER'+tail)
(p/'verify.py').write_bytes((b/'verify-observations.py').read_bytes())
r=subprocess.run(['python3',str(p/'verify.py')],capture_output=True,text=True)
assert r.returncode==1 and 'AssertionError' in r.stderr and "digest(old,'skills/sample/SKILL.md')==digest(new,'skills/sample/SKILL.md')" in r.stderr
print('PASS control plant rejected altered native preservation hash at the intended equality assertion, exit=1.')
# Inspect declared research payloads, excluding system-produced spawn transcripts.
for f in (b.parent/'resources/TASK-261017-yyg01d').iterdir():
 text=f.read_text()
 assert not re.search(r'/Users/|/home/(?![.\'\"]|\s)[A-Za-z0-9_-]+/|ivmbp|ivan@',text), f.name
print('PASS public research resources contain no checked machine identity/path patterns; manual inspection also performed.')
