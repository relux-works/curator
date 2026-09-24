# TASK-260924-1kpw4w: implement-manifest-dependency-directory

## Description
Implement the core 4.4 amendment (TASK-260924-2am4qa, once accepted): dependencies.skills directory selects the skill in that subfolder of the pinned repository ref, same grammar/containment as the Skillfile individual selector; identity, closure unification, lock and audit record the directory; absent directory unchanged.

## Scope
(define task scope)

## Acceptance Criteria
manifest dependency with directory installs the subfolder skill through install/update with lock and audit; invalid grammar/containment/missing folder/no SKILL.md refused; diamond of two folders of one repo; absent directory byte-identical; mutants killed; CHANGELOG
