import sys,copy,contextlib,io
sys.path.insert(0,'tools')
import validate as v
original=v.load_json
service=original(v.SUITE/'vectors'/'registry-service.json')
with contextlib.redirect_stdout(io.StringIO()):
    code=v.main()
assert code==0
print('Published positive control: validate.main exit 0')
passing=next(c for c in service['checkpoint_cases'] if c['name']=='checkpoint-equal-consistent')
for target in ['checkpoint-equal-inconsistent','live-below-checkpoint','live-above-prefix-mismatch','checkpoint-signature-invalid']:
    changed=copy.deepcopy(service)
    changed['checkpoint_cases']=[dict(copy.deepcopy(passing),name=target) if c['name']==target else c for c in changed['checkpoint_cases']]
    v.load_json=lambda path: changed if path==v.SUITE/'vectors'/'registry-service.json' else original(path)
    with contextlib.redirect_stdout(io.StringIO()): code=v.main()
    print(target, 'replaced with published equal-consistent: main exit',code)
    assert code==1
v.load_json=original
print('4/4 scenario replacements rejected; 1/1 unmodified control accepted')
