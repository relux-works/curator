"""Bounded source checks only; no renderer or runtime qualification."""
from collections import Counter
from pathlib import Path
import re, subprocess, urllib.parse

root=Path.cwd()
paths=subprocess.check_output(['git','diff','--name-only'],text=True).splitlines()
links=0; figures=0; private=re.compile(r'(?<![\w-])F-\d+(?!\w)|f-(?:r-name-namespace|wire-register-drift|book-muse-c7|exec-input-closure|base-binding|export-refusal-pins|export-cleanup-authority|workspace-result-modes)|\$[A-Za-z0-9_-]{20,}|/Users/|#swarma-review-3|wonder-\d+|astra-\d+')
for name in paths:
 p=root/name;s=p.read_text();old=subprocess.check_output(['git','show','HEAD:'+name],text=True)
 assert not (Counter(private.findall(s))-Counter(private.findall(old))), f'new private marker: {name}'
 assert p.suffix in ['.md','.puml','.dsl'],f'non-document change: {name}'
 for target in re.findall(r'(?<!!)\[[^\]]+\]\(([^)]+)\)|\[\[([^\s\]]+)(?:[^\]]*)\]\]',s):
  t=(target[0] or target[1]).strip().split()[0].strip('<>')
  if t.startswith(('http:','https:','mailto:','#')):continue
  dest=(p.parent/urllib.parse.unquote(t.split('#')[0])).resolve()
  assert dest.exists(),f'broken local target: {name} {t}'
  links+=1
 if p.suffix=='.puml':
  assert s.count('@startuml')==1 and s.count('@enduml')==1,name
  stack=[];in_note=False
  for line in s.splitlines():
   l=line.strip()
   if l.startswith('note ') and ':' not in l:in_note=True;continue
   if l=='end note':assert in_note,name;in_note=False;continue
   if in_note:continue
   if re.match(r'^(alt|opt|break|loop|par|group|critical)\b',l):stack.append(l)
   elif l=='end':assert stack,(name,l);stack.pop()
  assert not stack and not in_note,(name,stack)
  figures+=1
 # New internal markdown anchors resolve to headings or explicitly preserved old IDs.
 new_targets=set(re.findall(r'(?<!!)\[[^\]]+\]\(([^)]+)\)',s))-set(re.findall(r'(?<!!)\[[^\]]+\]\(([^)]+)\)',old))
 for t in new_targets:
  if t.startswith(('http:','https:','mailto:')) or '#' not in t:continue
  file,anchor=t.split('#',1);q=(p.parent/file) if file else p
  body=q.read_text();anchors=set(re.findall(r'<a id="([^"]+)"',body))
  for h in re.findall(r'^#+ (.*)$',body,re.M):
   anchors.add(re.sub(r'[^\w\- ]','',h.lower()).replace(' ','-'))
  assert anchor in anchors,(name,t)
print(f'checked {len(paths)}/{len(paths)} changed documentation files; {links}/{links} local path targets exist; {figures}/{figures} PlantUML source block pairs balanced; no new private markers')
print('Bound: source/path checks only. External register rows, cross-module status gate, PlantUML/DSL rendering and runtime qualification are not measured.')
