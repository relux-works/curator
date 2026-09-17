import unittest,sys,pathlib
sys.path.insert(0,'tools')
def flatten(s):
 for x in s:
  if isinstance(x,unittest.TestSuite): yield from flatten(x)
  else: yield x
suite=list(flatten(unittest.defaultTestLoader.discover('tools',pattern='test_*.py')))
a,b=map(int,sys.argv[1:]); print(f'BOUND {a}:{b} of {len(suite)}',flush=True)
for i,t in enumerate(suite[a:b],a): print(i,t.id(),flush=True)
r=unittest.TextTestRunner(verbosity=2).run(unittest.TestSuite(suite[a:b])); sys.exit(not r.wasSuccessful())
