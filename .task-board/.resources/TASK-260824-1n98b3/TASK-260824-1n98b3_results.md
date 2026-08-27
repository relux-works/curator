# Consumer pins advanced to rc.9

- curator: SPEC_PIN -> 0ed5c691e9208eea52f21db2fc05e226ce3516fd (the rc.9 release commit), PR 38 merged as 272b203; post-merge main run 32770177743 SUCCESS (all three OSes, default lanes against the released rc.9 suite).
- cocoaskills: released-suite pin advanced to the same rc.9 release commit, PR 46 rebase-merged; post-merge main run SUCCESS (including the post-merge-only Windows Merge-protocol shards).
- Both landed as separate consumer PRs per the impact-analysis landing order step 9; candidate inputs never became defaults implicitly.
- Producer runs RUN-260824-0647e3 and RUN-260824-82714a authored the halves and timed out post-delivery; orchestrator finished merges and verification. REVIEWER SCOPE: verify the two pin values against the rc.9 release commit and the two post-merge run conclusions; nothing else.