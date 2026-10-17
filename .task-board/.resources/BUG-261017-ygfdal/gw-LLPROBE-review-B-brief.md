# Review B brief: BUG-261017-ygfdal, Landlock probe recursion in test binaries

You decide accept or reject. Read the producer brief (gw-LLPROBE-brief.md) and
the bug description.

Check, with evidence, sweeping the whole diff:

1. Enumerate yourself every test package that can reach `probeScriptLandlock`.
   Each must terminate the hidden mode without running tests. A missed package
   is a reject.
2. The product-side guard makes probe-in-probe impossible, and keeps the hidden
   mode off every user-visible surface and the sanitized environment intact.
3. Timeout or cancellation leaves no descendants (process group kill on Linux),
   with a test that would fail without it.
4. The red-on-main test is real: it fails on main for the stated reason (not a
   setup failure) and passes with the fix.
5. Probe results are unchanged: present with Landlock, absent without, no error
   on expected absence.
6. No host names, IPs, personal names or local paths in committed text.

Record a surface table in your verdict.
