import sys,copy
sys.path.insert(0,'tools')
import validate as v
x=v.load_json(v.SUITE/'vectors/security-posture.json')
c=next(c for c in x['cases'] if c['name']=='refusal-mcp-allowlist-empty-with-declarations')
c['machine']['security_posture']='permissive'
p,ps,e,s,n=v._posture_resolve(c['name'],c)
a=c['expected']; a.update(profile=p,profile_source=ps,effective=e,sources=s,passable_env_null_explicit=n,diagnostics=v._posture_diagnostics(c['name'],c,p,e,n),outcome='proceeds')
for row in a['manager_rows']+a['environments_rows']:
 g=row['gate']
 if g=='security_posture': row.update(value=p,source=ps)
 mapping={'audit-mode':'audit_mode','registry-policy':'audit_registry_policy','transitive-system-modules':'transitive_system_modules','source-signers':'require_source_signers'}
 if g in mapping:
  k=mapping[g]; row.update(value=('true' if e[k] else 'false') if g=='source-signers' else e[k],source=s[k])
v.validate_security_posture_vectors(x)
print('SURVIVED: same-name MCP refusal replaced with permissive warning/proceeds; validator accepted')
