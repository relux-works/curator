import sys,copy,contextlib,io
sys.path.insert(0,'tools')
import validate as v
original=v.load_json
service=original(v.SUITE/'vectors'/'registry-service.json')
for target in ['checkpoint-equal-inconsistent','live-below-checkpoint','live-above-prefix-mismatch','checkpoint-signature-invalid']:
    changed=copy.deepcopy(service)
    for c in changed['checkpoint_cases']:
        if c['name']==target:
            c.update(checkpoint_configured=True,signature_valid=True,live_version=8,checkpoint_version=8,same_boundary_body=True,prefix_reproduced=True,ready=True,diagnostic=None,posture=None)
    v.load_json=lambda path: changed if path==v.SUITE/'vectors'/'registry-service.json' else original(path)
    with contextlib.redirect_stdout(io.StringIO()) as output:
        code=v.main()
    print(target, 'replaced with equal-consistent while preserving name: main exit',code,output.getvalue().strip())
v.load_json=original
