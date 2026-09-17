## Superseded archive-copy setup diagnostic: regenerate-check.log

```text
go run ./tools/generate-vectors -root .
git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json release/1.0.0-rc.6.json release/1.0.0-rc.7.json release/1.0.0-rc.8.json release/1.0.0-rc.9.json
diff --git a/conformance/v1/expected/byte-exact-snapshot_sha256.txt b/conformance/v1/expected/byte-exact-snapshot_sha256.txt
index ae77752..3270d63 100644
--- a/conformance/v1/expected/byte-exact-snapshot_sha256.txt
+++ b/conformance/v1/expected/byte-exact-snapshot_sha256.txt
@@ -1 +1 @@
-sha256:500ea934403d10a2a0b6b7e8874790e489ee002328d3dc0edbda2fe5be2bced0
+sha256:a34b7db9cfa7a022c49b9832414f70ef709635d6c1ee317f5ab41f726e66fbfe
diff --git a/conformance/v1/manifest.json b/conformance/v1/manifest.json
index 22f027f..fc69041 100644
--- a/conformance/v1/manifest.json
+++ b/conformance/v1/manifest.json
@@ -50,7 +50,7 @@
     },
     {
       "path": "expected/byte-exact-snapshot_sha256.txt",
-      "sha256": "sha256:1c54047e87b4315e8edf8d9cf8818130e76c316c966a18a5b6ecaf8b9a0362f7"
+      "sha256": "sha256:1f610707d84506ea9853d07c138c090367b687789548fda5fdada5b59efd6a36"
     },
     {
       "path": "expected/context_files.json",
@@ -250,7 +250,7 @@
     },
     {
       "path": "fixtures/byte-exact/subst.txt",
-      "sha256": "sha256:ec9a6c8c260b3977dad2f1cba09414599879ac5c3e7353c3205a40d5fbe122bc"
+      "sha256": "sha256:35728d9907b056b54acef663a7821d45e36965ff25771617aebc70c2bfa7369c"
     },
     {
       "path": "fixtures/external-repository/lfs-pointers.json",
@@ -4370,7 +4370,7 @@
     },
     {
       "path": "vectors/snapshot-acquisition.json",
-      "sha256": "sha256:98b2533f96e378b525c22c9c2f8be59da3c8a0de386670e99496c911d64e03ec"
+      "sha256": "sha256:0eed7fd343bdd1c450e4d0b1245b18f7d1d3e302cd172183120f4eb80c2db604"
     },
     {
       "path": "vectors/source-identities.json",
diff --git a/conformance/v1/vectors/snapshot-acquisition.json b/conformance/v1/vectors/snapshot-acquisition.json
index 470ded5..4ed32a4 100644
--- a/conformance/v1/vectors/snapshot-acquisition.json
+++ b/conformance/v1/vectors/snapshot-acquisition.json
@@ -10,7 +10,7 @@
         "The `subst.txt` entry MUST still contain the literal text `$Format:%H$` and `$Format:%h$`; the `crlf.txt` entry MUST contain CRLF line endings; the `mixed.txt` entry MUST contain both LF and CRLF line endings."
       ],
       "expected": "expected/byte-exact-snapshot_sha256.txt",
-      "expected_sha256": "sha256:500ea934403d10a2a0b6b7e8874790e489ee002328d3dc0edbda2fe5be2bced0",
+      "expected_sha256": "sha256:a34b7db9cfa7a022c49b9832414f70ef709635d6c1ee317f5ab41f726e66fbfe",
       "files": [
         {
           "bytes": 35,
@@ -33,9 +33,9 @@
           "sha256": "sha256:c76a5bc31dc90d785ff606b61890fb186b1b216635fbbc16ff9af27386dba279"
         },
         {
-          "bytes": 40,
+          "bytes": 65,
           "path": "subst.txt",
-          "sha256": "sha256:ec9a6c8c260b3977dad2f1cba09414599879ac5c3e7353c3205a40d5fbe122bc"
+          "sha256": "sha256:35728d9907b056b54acef663a7821d45e36965ff25771617aebc70c2bfa7369c"
         }
       ],
       "fixture": "fixtures/byte-exact",
diff --git a/release/1.0.0-rc.9.json b/release/1.0.0-rc.9.json
index 6d975f5..37324b8 100644
--- a/release/1.0.0-rc.9.json
+++ b/release/1.0.0-rc.9.json
@@ -12,7 +12,7 @@
     "verified_provider_contract": "host-execution-provider-v1"
   },
   "candidate_protocol_pin": {
-    "manifest_sha256": "sha256:7342f14dc74ccb56c4f1fac6ffdaa10694d791bd39f6627786c603d304ba3ad5",
+    "manifest_sha256": "sha256:44ddf4d215b41a2e3e2d30de579a880a8e8a5c74331b4d45ecfa76083850ec76",
     "suite_root": "conformance/v1"
   },
   "claim_v5": {
@@ -24,7 +24,7 @@
   "downstream_consumption": {
     "committed_release_pin_advanced": false,
     "environment": "CURATOR_CONFORMANCE_ROOT",
-    "required_manifest_sha256": "sha256:7342f14dc74ccb56c4f1fac6ffdaa10694d791bd39f6627786c603d304ba3ad5"
+    "required_manifest_sha256": "sha256:44ddf4d215b41a2e3e2d30de579a880a8e8a5c74331b4d45ecfa76083850ec76"
   },
   "historical_release": {
     "immutable": true,
make: *** [regenerate-check] Error 1

EXIT_CODE=2

```
## Superseded archive-copy setup diagnostic: validate.log

```text
python3 tools/validate.py
validation failed: byte-exact subst.txt lost a literal $Format: placeholder
make: *** [validate] Error 1

EXIT_CODE=2

```
