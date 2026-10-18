import subprocess
BASE='1724b4a52b4c2d163c0a3e6ab7c2ad2bbe54ac0c'
CANDIDATE='54bffdc4a8460c329fd5d68cd4dd412458720c4d'
AUTHORITY='5d8e6b07f7f485557d9bbb065c9692348ca065c2'
def paths(tree):
 return set(subprocess.check_output(['git','diff','--name-only','--no-renames',BASE,tree],text=True).splitlines())
overlap=sorted(paths(CANDIDATE)&paths(AUTHORITY))
print('base:',BASE,'candidate:',CANDIDATE,'authority:',AUTHORITY)
print('pre_review_path_overlap:', ', '.join(overlap) or 'none')
assert not overlap, 'pre_review_path_overlap: '+', '.join(overlap)
