STORY-260916-prdjid  cleanup_pending
  story commit: 6c820908188d8cbf18613ed63018b893b6014d4a
  board commit: b6fdd60eb40643622902c2071a288b63cf0d05ad
  reparented onto the advanced trunk through a three-way tree merge
  revalidated against the exact tree that landed
  note: safe cleanup is now eligible; `worktree cleanup` removes the workspace and branch only after exact commit ancestry, Story done, a committed board record, a clean workspace, and no active lease or RUN
  next 1: publish the landed commits as a non-default branch and open a pull request against the protected default branch
  next 2: review on the hosting platform, wait for the required checks, and merge the exact reviewed head
  next 3: in the control root, after the hosted merge, prove the landed commits delivered under their rewritten identities and move local trunk (a unique local commit refuses): task-board worktree reconcile-trunk
