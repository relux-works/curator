import sys,unittest,pathlib
root=pathlib.Path('/tmp/2mglq0-review/candidate');sys.path.insert(0,str(root/'tools'))
def flat(s):
 for t in s:
  if isinstance(t,unittest.TestSuite):yield from flat(t)
  else:yield t
alltests=list(flat(unittest.defaultTestLoader.discover(str(root/'tools'),pattern='test_*.py')))
heavy=[t for t in alltests if 'test_substituted_scenario_rejected_through_main' in t.id()]
rest=[t for t in alltests if t not in heavy]
which=sys.argv[1]; selected=heavy if which=='heavy' else rest[int(which)::4]
print('Discovery',len(alltests),'Selected',len(selected), 'shard',which,flush=True)
r=unittest.TextTestRunner(verbosity=1).run(unittest.TestSuite(selected));sys.exit(not r.wasSuccessful())
