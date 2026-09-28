from pathlib import Path
import subprocess
root=Path('.review-rev2/candidate').resolve()
cmd=['go','test','./pkg/agentic/systems/claude','-run','^TestTheBypassFlagIsSpelledAtExactlyTwoKnownSites$','-count=1']
rows=[]
def run(name, expected):
 p=subprocess.run(cmd,cwd=root,text=True,stdout=subprocess.PIPE,stderr=subprocess.STDOUT)
 rows.append(f'## {name}\nCommand: {" ".join(cmd)}\nExit: {p.returncode}\n{p.stdout}')
 assert p.returncode==expected, rows[-1]
for name,file,pkg in [('third package','pkg/agentic/review_third_spelling.go','agentic'),('second claude site','pkg/agentic/systems/claude/review_second.go','claude'),('second agy site','pkg/agentic/systems/agy/review_second.go','agy')]:
 path=root/file
 path.write_text('package '+pkg+'\nfunc reviewThirdSpelling() []string { return []string{"--dangerously-skip-permissions"} }\n')
 try: run(name,1)
 finally: path.unlink()
 run(name+' restored',0)
path=root/'pkg/agentic/systems/claude/args.go'
original=path.read_bytes()
try:
 path.write_bytes(original.replace(b'"--dangerously-skip-permissions"',b'"--review-silent"'))
 run('known claude const silent',1)
finally: path.write_bytes(original)
run('known site restored',0)
Path('.review-rev2/mutations.log').write_text('\n'.join(rows))
print('\n'.join(rows))
