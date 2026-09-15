import copy, json, subprocess, re
from pathlib import Path
from jsonschema import Draft202012Validator
from referencing import Registry, Resource
schemas={}; resources=[]
for d in ('schemas/v1','schemas/draft-sources-v1'):
 for p in Path(d).glob('*.json'):
  s=json.loads(p.read_text()); schemas[str(p)]=s; resources.append((s['$id'],Resource.from_contents(s)))
r=Registry().with_resources(resources)
def valid(p,o): return Draft202012Validator(schemas[p],registry=r).is_valid(o)
marker='schemas/draft-sources-v1/install-marker-v5.schema.json'
base=json.loads(Path('conformance/draft-sources-v1/schema-cases/install-marker-v5/valid.json').read_text())
network={'kind':'network-git','repository':'example.org/review','commit':{'object_format':'sha1','hex':'a'*40},'directory':'.'}
configured={'kind':'configured-git','source':'review','commit':{'object_format':'sha1','hex':'a'*40},'directory':'.'}
checks=0
for package in (network,configured):
 for field,value in [('attestation',{'registry':'trusted','status':'audited'}),('substituted','operator-development')]:
  obj=copy.deepcopy(base);obj['package']=package;obj[field]=value
  assert valid(marker,obj),(package,field);checks+=1
  obj['package']=base['package'];assert not valid(marker,obj);checks+=1
 for bad in ({},{'registry':'trusted'},{'registry':'trusted','status':'forged'}, {'registry':'trusted','status':'audited','key_id':True}):
  obj=copy.deepcopy(base);obj['package']=package;obj['attestation']=bad
  assert not valid(marker,obj),bad;checks+=1
print(f'Independent marker regression checks: {checks}/{checks}')
# Require one endpoint and no more than two, then narrow only the overflow refusal.
p='schemas/draft-sources-v1/source-policy-v1.schema.json'
case=json.loads(Path('conformance/draft-sources-v1/schema-cases/source-policy-v1/invalid-too-many.json').read_text())
assert not valid(p,case)
m=copy.deepcopy(schemas[p]);m['properties']['repositories']['additionalProperties']['properties']['endpoints']['maxItems']=3
assert Draft202012Validator(m,registry=r).is_valid(case)
for n in (0,1,2):
 obj=copy.deepcopy(case)
 for value in obj['repositories'].values(): value['endpoints']=value['endpoints'][:n]
 assert valid(p,obj)==(n>0),n
print('Endpoint cardinality: 0 rejected; 1 and 2 admitted; 3 rejected; narrowed >3 refusal mutant detected')
p='schemas/draft-sources-v1/skillfile-v2.schema.json'
base={'schema_version':2,'sources':{'src':{'path':'.'}},'skills':[{'name':'review','from':'src','directory':'.'}]}
assert valid(p,base)
for extra in ({'git':'https://example.org/kit.git'},{'tag':'v1'},{'source':'review'},{'include':['*']}):
 obj=copy.deepcopy(base);obj['skills'][0].update(extra);assert not valid(p,obj)
for path in ('..','../review','x/../review','/review','x./review','x /review','x\\review'):
 obj=copy.deepcopy(base);obj['skills'][0]['directory']=path;assert not valid(p,obj),path
print('Independent selector attacks: 11/11 rejected; root selector admitted')
# Validate actual author guide examples, including contextual fragments.
examples=re.findall(r'```json\n(.*?)\n```',Path('docs/skillfile-sources.md').read_text(),re.S)
for block in examples:
 obj=json.loads(block)
 if 'repositories' in obj: schema='schemas/draft-sources-v1/source-policy-v1.schema.json'
 else:
  schema=p
  if 'schema_version' not in obj:
   if 'from' in obj: obj={'schema_version':2,'sources':{'team':{'path':'.'}},'skills':[obj]}
   else: obj={'schema_version':2,'sources':{'team':obj},'skills':[{'name':'review','from':'team','directory':'.'}]}
 assert valid(schema,obj),obj
print(f'Author guide JSON examples: {len(examples)}/{len(examples)} schema-valid')
candidate='4087f02f5459a96d1a03d78ddb343d82608df0e6';baseoid='d019f0e7179520b5c8dcde321c4fe51e04552f58'
paths=subprocess.check_output(['git','diff','--name-only',baseoid,candidate],text=True).splitlines()
for path in paths: assert Path(path).read_bytes()==subprocess.check_output(['git','show',f'{candidate}:{path}']),path
assert subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip()==baseoid
print(f'Exact CR bytes: {len(paths)}/{len(paths)}; HEAD remains base; HEAD..main='+subprocess.check_output(['git','rev-list','--count','HEAD..main'],text=True).strip())
print('Manager semantic execution: 0/'+str(len(json.loads(Path('conformance/draft-sources-v1/semantic-cases.json').read_text())))+' (unverified)')
