import sys,unittest,builtins
sys.path.insert(0,'tools')
import test_validate as tv
which=sys.argv[1]
suite=unittest.defaultTestLoader.discover('tools',pattern='test_*.py')
def flatten(s):
 for t in s:
  if isinstance(t,unittest.TestSuite): yield from flatten(t)
  else: yield t
alltests=list(flatten(suite)); heavy='test_validate.WriteNofollowVectorTests.test_substituted_scenario_rejected_through_main'
if which.startswith('heavy'):
 i=int(which[-1]); names=builtins.sorted(tv.validate.WRITE_NOFOLLOW_CASES); chosen=set(names[i::3])
 tv.sorted=lambda x,*a,**kw: [n for n in builtins.sorted(x,*a,**kw) if n in chosen] if x is tv.validate.WRITE_NOFOLLOW_CASES else builtins.sorted(x,*a,**kw)
 selected=[t for t in alltests if t.id()==heavy]
 print('Heavy scenario shard',i,'count',len(chosen),'of',len(names),flush=True)
else:
 i=int(which); selected=[t for j,t in enumerate([t for t in alltests if t.id()!=heavy]) if j%3==i]
print('Tests',len(selected),flush=True)
r=unittest.TextTestRunner(verbosity=1).run(unittest.TestSuite(selected)); sys.exit(not r.wasSuccessful())
