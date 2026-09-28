import sys,json,copy,itertools,pathlib,subprocess
root=pathlib.Path(__file__).resolve().parent/'candidate'
sys.path.insert(0,str(root/'tools'))
import validate as v
registry,paths=v.schema_registry()
sname='launch-env-fragment-v2.schema.json'
schema=v.load_json(paths[sname]); base=v.load_json(v.SUITE/'schema-cases/launch-env-fragment-v2/valid.json')
def valid(p):
 x=copy.deepcopy(base); x['permissions']=p
 return not list(v.Draft202012Validator(schema,registry=registry).iter_errors(x))
for mode,locked,source in itertools.product(['native','yolo'],[False,True],['profile','global','default']):
 expected=(locked==(source=='global') and (mode=='native' or source=='profile'))
 assert valid(dict(mode=mode,locked=locked,source=source))==expected
print('truth table 12/12; four valid, eight invalid')
# Drive the actual conformance validation entry with extra negative cases, in disposable copy.
indexpath=v.SUITE/'schema-cases/index.json'; original=indexpath.read_bytes(); index=json.loads(original)
shapes=[{},dict(mode='native',locked=False,source='profile',extra=True),dict(mode='yolo',locked=False,source='global'),dict(mode='native',locked='false',source='profile')]
try:
 for i,p in enumerate(shapes):
  instance=copy.deepcopy(base); instance['permissions']=p
  rel=f'launch-env-fragment-v2/reviewer-{i}.json'; (v.SUITE/'schema-cases'/rel).write_text(json.dumps(instance))
  index.append(dict(schema=sname,instance=rel,valid=False))
 indexpath.write_text(json.dumps(index)); v.validate_schemas(); print('validate_schemas additional attacks 4/4 rejected')
finally:
 indexpath.write_bytes(original)
 for i in range(len(shapes)): (v.SUITE/f'schema-cases/launch-env-fragment-v2/reviewer-{i}.json').unlink()
# Narrow each conditional rather than delete it. Existing cases must kill each mutant.
sp=paths[sname]; orig=sp.read_bytes()
try:
 for i in range(3):
  mutated=copy.deepcopy(schema)
  rule=mutated['properties']['permissions']['allOf'][i]
  if i==0: rule['if']['properties']['mode']={'const':'yolo'}
  if i==1: rule['if']['properties']['mode']={'const':'yolo'}
  if i==2: rule['if']['properties']['source']={'const':'global'}
  sp.write_text(json.dumps(mutated))
  try: v.validate_schemas()
  except v.ValidationFailure as e: print(f'mutant {i+1} killed: {e}')
  else: raise AssertionError(f'mutant {i+1} survived')
finally: sp.write_bytes(orig)
# v2 is exactly v1 plus revision identity and permission member.
a=v.load_json(paths['launch-env-fragment-v1.schema.json']); b=copy.deepcopy(schema)
for k in ['$id','title']: b[k]=a[k]
b['properties']['fragment']=a['properties']['fragment']; b['required'].remove('permissions'); del b['properties']['permissions']
assert a==b; print('v2 equals v1 plus token and required permissions: PASS')
# Unlocked system defaults are legitimate and become effective profile knobs under manager §1.3.
system={'schema_version':2,'environments':{'permissions':{'companyA':'native'}}}
sv=v.Draft202012Validator(v.load_json(paths['system-config-v2.schema.json']),registry=registry)
errors=list(sv.iter_errors(system)); assert not errors, [e.message for e in errors]
print('unlocked system permissions default valid; effective profile native unlocked encoding valid:',valid(dict(mode='native',locked=False,source='profile')))
