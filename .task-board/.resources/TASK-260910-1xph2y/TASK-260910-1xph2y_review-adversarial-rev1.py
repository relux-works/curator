import copy, json, subprocess
from pathlib import Path
from jsonschema import Draft202012Validator
from referencing import Registry, Resource
root=Path.cwd(); resources=[]; schemas={}
for d in ('schemas/v1','schemas/draft-sources-v1'):
 for p in Path(d).glob('*.json'):
  s=json.loads(p.read_text()); schemas[str(p)]=s
  resources.append((s['$id'],Resource.from_contents(s)))
r=Registry().with_resources(resources)
def valid(path, obj): return Draft202012Validator(schemas[path],registry=r).is_valid(obj)
p='schemas/draft-sources-v1/install-marker-v5.schema.json'
v5=json.loads(Path('conformance/draft-sources-v1/schema-cases/install-marker-v5/valid.json').read_text())
v4=json.loads(Path('conformance/v1/schema-cases/install-marker-v4/valid.json').read_text())
for field,value in [('attestation',{'registry':'trusted','status':'audited'}),('substituted','operator-development')]:
 old=copy.deepcopy(v4); old[field]=value
 new=copy.deepcopy(v5); new[field]=value
 print(field, 'v4=',valid('schemas/v1/install-marker-v4.schema.json',old),'v5=',valid(p,new))
 assert valid('schemas/v1/install-marker-v4.schema.json',old)
 assert not valid(p,new)
# Restrict refusal to >3 endpoints rather than >2: shipped 3-endpoint negative must catch this widening.
p='schemas/draft-sources-v1/source-policy-v1.schema.json'
m=copy.deepcopy(schemas[p]); m['properties']['repositories']['additionalProperties']['properties']['endpoints']['maxItems']=3
case=json.loads(Path('conformance/draft-sources-v1/schema-cases/source-policy-v1/invalid-too-many.json').read_text())
assert not valid(p,case)
assert Draft202012Validator(m,registry=r).is_valid(case)
print('Narrowed endpoint refusal mutant killed by invalid-too-many (original rejects; mutant admits)')
# Independently exercise the actual documented Draft202012Validator entry with prohibited forms.
p='schemas/draft-sources-v1/skillfile-v2.schema.json'
base={'schema_version':2,'sources':{'src':{'path':'.'}},'skills':[{'name':'review','from':'src','directory':'.'}]}
assert valid(p,base)
for extra in ({'git':'https://example.org/kit.git'},{'tag':'v1'},{'source':'review'},{'include':['*']}):
 obj=copy.deepcopy(base); obj['skills'][0].update(extra); assert not valid(p,obj),extra
for path in ('..','../review','x/../review','/review','x./review','x /review','x\\review'):
 obj=copy.deepcopy(base); obj['skills'][0]['directory']=path; assert not valid(p,obj),path
print('Independent selector attacks: 11/11 rejected; root selector positive admitted')
# Establish exact candidate read identity without creating a tree/index/commit.
candidate='96b33ebf97e4dfe01a705124219fbec1adf2f5c1'; baseoid='d019f0e7179520b5c8dcde321c4fe51e04552f58'
paths=subprocess.check_output(['git','diff','--name-only',baseoid,candidate],text=True).splitlines()
for path in paths:
 expected=subprocess.check_output(['git','show',f'{candidate}:{path}'])
 assert Path(path).read_bytes()==expected,path
print(f'Exact CR candidate bytes: {len(paths)}/{len(paths)} paths match; HEAD='+subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip())
