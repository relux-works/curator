STORY-260919-37szes  cleanup_pending
  story commit: 854f1f8ab2df0097491af7f9b1c77129694080c0
  board commit: f0a92b8b1a076a03738835c4b9d45f3f3e319b8c
  note: safe cleanup is now eligible; `worktree cleanup` removes the workspace and branch only after exact commit ancestry, Story done, a committed board record, a clean workspace, and no active lease or RUN
  next 1: publish the landed commits as a non-default branch and open a pull request against the protected default branch
  next 2: review on the hosting platform, wait for the required checks, and merge the exact reviewed head
  next 3: in the control root, after the hosted merge, prove the landed commits delivered under their rewritten identities and move local trunk (a unique local commit refuses): task-board worktree reconcile-trunk
