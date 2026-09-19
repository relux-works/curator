STORY-260910-1s75e1  cleanup_pending
  story commit: f99db0b69a952e74567a6475455cdcf3c0c516a7
  board commit: 34ce1394a3707eed5f327f802a53d73111d068a9
  reparented onto the advanced trunk through a three-way tree merge
  revalidated against the exact tree that landed
  note: safe cleanup is now eligible; `worktree cleanup` removes the workspace and branch only after exact commit ancestry, Story done, a committed board record, a clean workspace, and no active lease or RUN
  next 1: publish the landed commits as a non-default branch and open a pull request against the protected default branch
  next 2: review on the hosting platform, wait for the required checks, and merge the exact reviewed head
  next 3: in the control root, after the hosted merge, prove the landed commits delivered under their rewritten identities and move local trunk (a unique local commit refuses): task-board worktree reconcile-trunk
