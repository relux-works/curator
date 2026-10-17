from pathlib import Path
source=Path(__file__).with_name('reproduce.py').read_text().split("env,p=new('A')")[0]
exec(source.replace("'sandbox-05'", "'sandbox-06'"))
env,p=new('C-selected'); skill(p/'agents/skills/sample'); skillfile(p)
write(p/'.agents/skills/sample/SKILL.md','Only copy of manual skill instructions\n')
write(p/'.agents/skills/sample/private-note.md','Only copy of manual notes\n')
cli(env,p,'project','resolve','.')
snap('C selected BEFORE',p/'.agents')
cli(env,p,'install','.')
snap('C selected AFTER',p/'.agents')

env,p=new('C-native-adapter'); skill(p/'agents/skills/sample'); skillfile(p)
write(p/'.codex/skills/sample/SKILL.md','Only copy of manual native skill\n')
cli(env,p,'project','resolve','.')
snap('C adapter BEFORE',p/'.codex')
cli(env,p,'install','.')
snap('C adapter AFTER',p/'.codex'); snap('C adapter canonical AFTER',p/'.agents')

# Establish fixture observations as bounded executable controls, not a suite gate.
import hashlib
case=BASE/'sandbox-05'
a=case/'A/project'
assert not (a/'.agents/skills/unmanaged-skill').exists()
assert not (a/'.agents/skills/unmanaged-notes').exists()
assert not (a/'.agents/skills/loose.txt').exists()
assert (a/'.codex/skills/manual-link').is_symlink()
assert not (a/'.codex/skills/manual-link').exists()
assert (a/'.agents/skills/sample/SKILL.md').exists()
c=case/'C-project/project/.agents'
assert (c/'.instructions/hand.md').read_text()=='Only copy of hand-authored instructions\n'
assert (c/'README.md').read_text()=='Manual managed-root note\n'
native=case/'D-E/home/.claude'
backups=list((native/'.agent-environment-backup').glob('*/CLAUDE.md'))
assert any(x.read_text()=='Only copy of manual Claude instructions\n' for x in backups)
log('OBSERVATION ASSERTIONS: A deletion + dangling link + managed positive control; C unrelated files preserved; native takeover backup bytes preserved: PASS')
