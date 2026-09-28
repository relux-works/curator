import json,glob,subprocess,re,sys
refs=set()
def walk(x):
    if isinstance(x,dict):
        for k,v in x.items():
            if k=='$ref' and '../v1/' in v: refs.add(v)
            walk(v)
    elif isinstance(x,list):
        for i in x: walk(i)
for f in glob.glob('schemas/skillfile-sources-v1/*.json'): walk(json.load(open(f)))
def ptr(doc,p):
    for t in [t.replace('~1','/').replace('~0','~') for t in p.strip('/').split('/') if t]: doc=doc[t]
    return doc
bad=0
for r in sorted(refs):
    file,_,p=r.partition('#'); file=file.replace('../v1/','schemas/v1/')
    head=json.load(open(file))
    try: rc=json.loads(subprocess.check_output(['git','show','v1.0.0-rc.10:'+file],stderr=subprocess.DEVNULL))
    except Exception: print('MISSING-AT-rc.10',r); bad+=1; continue
    try: a=ptr(head,p); b=ptr(rc,p)
    except KeyError: print('PTR-MISSING-AT-rc.10',r); bad+=1; continue
    ok=json.dumps(a,sort_keys=True)==json.dumps(b,sort_keys=True)
    # also transitive local refs inside definition
    inner=set(re.findall(r'"\$ref": "(#[^"]+)"',json.dumps(a)))
    for i in inner:
        if json.dumps(ptr(head,i[1:]),sort_keys=True)!=json.dumps(ptr(rc,i[1:]),sort_keys=True): ok=False; print('  inner differs',i)
    print('OK ' if ok else 'DIFF',r, ('(+%d inner)'%len(inner)) if inner else ''); bad+= not ok
print('refs=%d bad=%d'%(len(refs),bad)); sys.exit(1 if bad else 0)
