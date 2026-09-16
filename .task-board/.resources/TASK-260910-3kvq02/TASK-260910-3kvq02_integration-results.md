STORY-260910-197y84  cleanup_pending
  story commit: 75565149de556434a55dd23e65337973a374112c
  board commit: 12f1287ee0fb538f9ca004dd53b870e824e5baf2
  reparented onto the advanced trunk through a three-way tree merge
  revalidated against the exact tree that landed
  note: safe cleanup is now eligible; `worktree cleanup` removes the workspace and branch only after exact commit ancestry, Story done, a committed board record, a clean workspace, and no active lease or RUN
  next 1: publish the landed commits as a non-default branch and open a pull request against the protected default branch
  next 2: review on the hosting platform, wait for the required checks, and merge the exact reviewed head
  next 3: in the control root, after the hosted merge, prove the landed commits delivered under their rewritten identities and move local trunk (a unique local commit refuses): task-board worktree reconcile-trunk
