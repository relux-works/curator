from pathlib import Path
import subprocess,json
root=Path('.temp/TASK-260910-3kvq02')
cases=[
('legacy-metadata-link','internal/manifest/expand.go','if info.Mode()&os.ModeSymlink != 0 {','if info.Mode()&os.ModeSymlink != 0 && relative == skillspec.CanonicalManifestName {','./internal/manifest','TestExpandMetadataLinks/csk-skill.json'),
('missing-before-exclude','internal/manifest/expand.go','outputs); err != nil {\n\t\t\treturn nil, err','outputs); err != nil && len(s.Exclude) == 0 {\n\t\t\treturn nil, err','./internal/manifest','TestExpandRefusals/missing-excluded'),
('case-destination','internal/manifest/expand.go','key := strings.ToLower(member.Decl.Name)','key := member.Decl.Name','./internal/manifest','TestExpandRefusals/case-collision'),
('discovered-name','internal/manifest/expand.go','if !ok || !identifiers.Valid(name) {','if !ok || name == "" {','./internal/manifest','TestExpandRefusals/invalid-discovered-name'),
('description-type','internal/manifest/expand.go','if !ok || strings.TrimSpace(description) == "" {','if ok && strings.TrimSpace(description) == "" {','./internal/manifest','TestExpandRefusals/metadata-description'),
('physical-escape','internal/manifest/expand.go','if !physicalWithin(root, physical) {','if !physicalWithin(root, physical) && directory == ".." {','./internal/manifest','TestExpandPhysicalBoundariesAndPruning/escape'),
('manifest-identity','internal/manifest/expand.go','present && declared != name {','present && declared != name && spec.SchemaVersion > 1 {','./internal/manifest','TestExpandDeclaredManifestIdentity/other'),
('alias-commit','internal/closure/selections.go','ok && previous != commit {','ok && previous != commit && decl.Selector.Directory == "." {','./internal/closure','TestBuildExpandedGitIdentity/split-commit'),
('output-read-failure','internal/manifest/expand.go','err != nil && !os.IsNotExist(err) {','err != nil && os.IsPermission(err) {','./internal/manifest','TestExpandOutputReadFailure'),
]
results=[]
for name,file,old,new,pkg,test in cases:
 p=Path(file); original=p.read_bytes(); source=original.decode(); assert source.count(old)==1,(name,source.count(old))
 cmd=['go','test','-count=1',pkg,'-run','^'+test+'$']
 try:
  p.write_text(source.replace(old,new,1))
  with (root/(name+'.log')).open('w') as log:
   r=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT,timeout=90)
  result={'mutant':name,'command':cmd,'exit_code':r.returncode,'killed':r.returncode==1}
  results.append(result); print(json.dumps(result),flush=True)
 finally:
  p.write_bytes(original)
 assert p.read_bytes()==original
(root/'mutants.json').write_text(json.dumps(results,indent=2)+'\n')
