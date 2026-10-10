"""Document packaging and rejection-coverage check; no runtime tests/builds."""
import argparse, json, pathlib, re
p=argparse.ArgumentParser();p.add_argument('document');p.add_argument('--narrow-retire',action='store_true');a=p.parse_args()
s=pathlib.Path(a.document).read_text()
# Narrowing mutant preserves the row and citations but excludes delete from its canonical request set.
if a.narrow_retire:
 old='Current closed `user.retire.home` request admits `archive` / `delete`;'
 assert old in s
 s=s.replace(old,'Current closed `user.retire.home` request admits `archive` only;',1)
refs=dict(re.findall(r'^\| ([A-Z][A-Z0-9]*) \| \[[^\]]+\]\((https://github.com/[^)]+)\) \|$',s,re.M));reverse={v:k for k,v in refs.items()}
rows={m[0]:m[1] for m in re.findall(r'^\| ([NWILPCGRQXDV]\d+) ([^|]+\|[^\n]+)$',s,re.M)}
assert len(rows)==52,len(rows)
assert len(s.encode())<81920
expected=None;tables=0
for n,l in enumerate(s.splitlines(),1):
 if l.startswith('|'):
  count=len(re.split(r'(?<!\\)\|',l))-2
  if expected is None:expected=count;tables+=1
  assert count==expected,(n,count,expected)
 else:expected=None
links=re.findall(r'\[([^\]]+)\]\((https://github.com/[^)]+)\)',s);ranges=0
for label,url in links:
 base,_,anchor=url.partition('#');assert base in reverse,url
 assert re.search(r'/blob/[a-f0-9]{40}/',base),url
 if anchor:
  m=re.fullmatch(r'L(\d+)-L(\d+)',anchor);assert m,url
  lo,hi=map(int,m.groups());lines=pathlib.Path('.temp/contract-sources',reverse[base]+'.md').read_text().splitlines()
  assert 1<=lo<=hi<=len(lines),url
  assert label==f'{reverse[base]}:L{lo}–L{hi}',label
  ranges+=1
for pat in [r'/Users/',r'/home/',r'file://',r'BEGIN .*PRIVATE KEY',r'gh[pousr]_[A-Za-z0-9]{20}',r'\x00']:
 assert not re.search(pat,s,re.I),pat
assert all(l==l.rstrip() for l in s.splitlines())
# Named regression check: all four rejected coverage surfaces need substantive boundaries,
# owner/propagation cells and pinned citations, not merely the audit IDs.
requirements={
 'P7':['A03.9','`archive` / `delete`','operator-only','dispatcher for its own agents','installer for configured named services','§4.4.1','OWNER DECISION NEEDED','B03'],
 'P8':['A07.5','Qualified-version-range','successful probe','same runtime family','same managed home','no transport conversion','Specified architecture target','B07'],
 'P9':['A07.8','Lane 1 projection: private','Lane 2 consumer: separate private','Lane 3 module: public','invented fixtures','proves neither private extraction nor SH2','SHR'],
 'P10':['A07.10','supersedes private-fork','tagged public module','immutable tag','operator maintenance','no package moved','B12']}
passed=0
for key,terms in requirements.items():
 for term in terms:
  assert term.lower() in rows[key].lower(),f'rev2_rejection_regression {key}: missing {term}'
 assert '/blob/' in rows[key] and '#L' in rows[key],key
 passed+=1
print(f'rev2_rejection_regression: {passed}/4 rejection surfaces covered')
print(f'packaging: 52/52 unique rows, {ranges}/{ranges} pinned ranges in bounds, {len(links)}/{len(links)} known pinned links, {tables} consistent tables, {len(s.encode())}/81920 bytes; bounded hygiene/whitespace pass')
print('Bound: textual coverage/range checks; semantic accuracy requires source review. No runtime behavior or product gate tested.')
