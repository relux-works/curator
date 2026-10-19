# BUG-261018-dbeghc local evidence index, revision 2

Archive retained outside Git at project-relative `.temp/STORY-261018-ogw2f1/worktree/.temp/BUG-261018-dbeghc/BUG-261018-dbeghc_local-evidence-02.zip`. SHA256 `ca255279e325cf0a53596e21776b1c1e96214337ab72129c1f5f90d83c07d95e`; 1083377 bytes. The previous archive remains retained but its early root identity is superseded. Board binary attachment was refused (`resource_binary_refused`, CLI exit 1); this plain-text index is the supported representation. Full sanitized logs and source/published hashes are in the archive. No truncated or missing log is accepted.

Accepted streams all follow restoration of the immutable pinned fixture root. The three original bulk streams and first 120 install tests were replayed; the PNPM replay without its pinned PATH was replaced by a whole-package pinned retry. Final inventory and platform gate exited 0. Commands ran standalone: `go test -json -count=1 -timeout=8m`, with package arguments or the retained -run masks; root replays also used `-p 2`. CURATOR_CONFORMANCE_ROOT pointed at the verified pinned root. CLI/pinned-PNPM commands prepend the task-local pnpm 10.33.0 path; signer-fixture batches remove inherited numbered Git overrides. Detailed actual exits follow.

## local-inventory-04.log

```text
{
  "candidate": "9e9d63976b5adfa61879183e913de8a1d7d61da8",
  "packages": 85,
  "served": 85,
  "heavy_package_inventory": {
    "cmd/curator": {
      "expected": 420,
      "observed": 420
    },
    "internal/install": {
      "expected": 363,
      "observed": 363
    },
    "internal/envprofile": {
      "expected": 375,
      "observed": 375
    }
  },
  "verdicts": 11746,
  "sha256": "38e39cf17fb2540cf235fdc89e777d05f6b562529fa4b288177ea2735c7ca55f"
}

```

## local-provenance.json

```text
{
  "candidate": "9e9d63976b5adfa61879183e913de8a1d7d61da8",
  "packages": 85,
  "served": 85,
  "heavy_package_inventory": {
    "cmd/curator": {
      "expected": 420,
      "observed": 420
    },
    "internal/install": {
      "expected": 363,
      "observed": 363
    },
    "internal/envprofile": {
      "expected": 375,
      "observed": 375
    }
  },
  "verdicts": 11746,
  "sha256": "38e39cf17fb2540cf235fdc89e777d05f6b562529fa4b288177ea2735c7ca55f",
  "sources": [
    {
      "log": "local-conformance-retry-01.json",
      "sha256": "4e0040641b80151d8bd8c8afe711ccb4cc937cb46045909fd4dc89da18b183f8",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 883,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/interop/environments": "pass"
      }
    },
    {
      "log": "local-artifactpolicy-retry-01.json",
      "sha256": "41bf07f897165bf40e983147b3c0d77444d1c266ded69e07624274ea0f78439e",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 1712,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/artifactpolicy": "pass"
      }
    },
    {
      "log": "local-cli-01-clean.json",
      "sha256": "7677d5ac6043dcf49098800c414bebf48c00cf23876390e03eb5e8cfc9e10a77",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 820,
      "package_verdicts": {
        "github.com/relux-works/curator/cmd/curator": "pass"
      }
    },
    {
      "log": "local-cli-02.json",
      "sha256": "a23f8f18a52c8b30c20594d8965d1e2dbee873afd87f8dfcb8732e6407aeb60f",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 721,
      "package_verdicts": {
        "github.com/relux-works/curator/cmd/curator": "pass"
      }
    },
    {
      "log": "local-root-final-01.json",
      "sha256": "787d6bf7e1bed86cfc09c1729d63b9f2d7fc53e644b5ee3c5db81f2d7aa92fa8",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 3,
      "package_verdicts": {
        "github.com/relux-works/curator/cmd/curator-spec-pin": "skip"
      }
    },
    {
      "log": "local-root-final-02.json",
      "sha256": "eb758f5e3225eada1f0e354a0a772e5874c846dabd477ef3b097fb53bf5bd175",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 88,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/adapters": "pass"
      }
    },
    {
      "log": "local-root-final-03.json",
      "sha256": "056e3c11fde4fbd765f828f4596e0c6e042d3ad4c55f2359c0fc988c3b54bd4b",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 3,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/artifactpolicy/conformance": "skip"
      }
    },
    {
      "log": "local-root-final-04.json",
      "sha256": "4aec9ec5176cd84d418a677ed2f1c8b1638fff7726b89142611d4c431fc1c616",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 2218,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/audit": "pass"
      }
    },
    {
      "log": "local-root-final-05.json",
      "sha256": "48b9d52af2d3e9e24bbccad1242330c5bb49056cf049c9a786452d1567ef9bfa",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 849,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/buildcache": "pass"
      }
    },
    {
      "log": "local-root-final-06.json",
      "sha256": "f085fa344182245ef16bca17a17e71291056c27bbba80f5c58463f3d52c79b61",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 452,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/buildmeta": "pass"
      }
    },
    {
      "log": "local-root-final-07.json",
      "sha256": "e387d6c39db846a45b1b79b6b477a1d9aeac0e83abb236d3bfb5dc8145615260",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 3054,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/buildrepo": "pass"
      }
    },
    {
      "log": "local-root-final-08.json",
      "sha256": "632c0548423f7fc0efb235e4e38032568476850692aa43a65499eb4f99ca5a28",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 135,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/buildsource": "pass"
      }
    },
    {
      "log": "local-root-final-09.json",
      "sha256": "cc8f3c7da97d4037ccc6c57da094d7ac75224933c1f772eac7e4566cd336bba7",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 88,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/capabilities": "pass"
      }
    },
    {
      "log": "local-root-final-10.json",
      "sha256": "1ae6fe9a1e4a47e8f948c13f157170075abc3ba018de78cb45ccde0af0722e13",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 410,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/closure": "pass"
      }
    },
    {
      "log": "local-root-final-11.json",
      "sha256": "609ac395fa7423c522da1fe6612ae3bd18259b59f0aa67e8480beafa61202d48",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 396,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/closureexec": "pass"
      }
    },
    {
      "log": "local-root-final-12.json",
      "sha256": "16b00af06856082acb77f56d264b883ee4790e841a7e436ec6edaa71fd2e25d1",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 1776,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/closuregraph": "pass"
      }
    },
    {
      "log": "local-root-final-13.json",
      "sha256": "580246c0ab2ab5290112a775751af7e0796a12264a565f3a34f9b36d58d9a9a4",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 2921,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/config": "pass"
      }
    },
    {
      "log": "local-root-final-14.json",
      "sha256": "2580f2916509c7eede2f6d42156f8d33107e996ba98b04866fcec278dbaa2c81",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 85,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/conformancecoverage": "pass"
      }
    },
    {
      "log": "local-root-final-15.json",
      "sha256": "ada6a157c1acd91b78abb74aa75e97c17c4eacce3afc5e95a0885ad4b75d2a4c",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 56,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/contextaudit": "pass"
      }
    },
    {
      "log": "local-root-final-16.json",
      "sha256": "7ba0310f323f1e35c19e5fc68079aaae4e0bbec61d79b96452fcd823a8967584",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 69,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/contextlock": "pass"
      }
    },
    {
      "log": "local-root-final-17.json",
      "sha256": "209e78e5103403746c450b91a7b3fa791a1b860cdcf55bd8b02f7f19f407a524",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 249,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/contextmaterialize": "pass"
      }
    },
    {
      "log": "local-root-final-18.json",
      "sha256": "1c9cb1fca88de01b36067bd01f01bdc2ee6d49ca2045665c266003e7fc606a99",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 40,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/contextpkg": "pass"
      }
    },
    {
      "log": "local-root-final-19.json",
      "sha256": "32fa8d4d245e7dcf2a1b10165e0a7b453e14e115d17ae77d3a68558928de8554",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 124,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/contextresolve": "pass"
      }
    },
    {
      "log": "local-root-final-20.json",
      "sha256": "0f075ff504fe101389bd8bb5828e2402bf18bdaf69153d37eaf369cea336b8a1",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 40,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/contextstore": "pass"
      }
    },
    {
      "log": "local-root-final-21.json",
      "sha256": "7822d6497d1f33c38015d67e609c69fdf11a79c09bcd6c47b8ecfea85005d8cd",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 2167,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/crossconformance": "pass"
      }
    },
    {
      "log": "local-root-final-22.json",
      "sha256": "7a7440903970005e7b12b5e1d5abd201e9c0602297e80ce468e2714b70bb216d",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 169,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/devsub": "pass"
      }
    },
    {
      "log": "local-root-final-23.json",
      "sha256": "14ed7089f5f716c950e064aac476540a58201073a1c5bd524178d7dce1755b02",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 12,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/envfiles": "pass"
      }
    },
    {
      "log": "local-root-final-24.json",
      "sha256": "63e87d0fa65c26462760162adf8bd2e4638f8f612e4482f4cbd7583e3ef13957",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 333,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/envfragment": "pass"
      }
    },
    {
      "log": "local-root-final-25.json",
      "sha256": "70464056a65384df3ff78d1883f0856ba35aac1923e4d315926be3d629e78603",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 527,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/envmarker": "pass"
      }
    },
    {
      "log": "local-root-final-26.json",
      "sha256": "d37a205c1864e9ff486ceab128900285a76e2f022cc9ff0c301ff7696ad6f2ae",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 64,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/envregistry": "pass"
      }
    },
    {
      "log": "local-root-final-27.json",
      "sha256": "827ed2cf8caadbae8a2d5bafe9524a6152279ace5df7d558c51c87ea881b3737",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 716,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/envsnapshot": "pass"
      }
    },
    {
      "log": "local-root-final-28.json",
      "sha256": "9998092b4b7143fb08eaeb1a4071136f39ddf1c1d8b1069873ecdf077e826879",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 142,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/gitcred": "pass"
      }
    },
    {
      "log": "local-root-final-29.json",
      "sha256": "83c98b3dd7f448843bc0c8009af4cc7bf938f481a4c2fc9d0d1e11ceeef91f9d",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 60,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/gitignore": "pass"
      }
    },
    {
      "log": "local-root-final-30.json",
      "sha256": "13f03b8b312d5ef0c381f2ac5c27bde3ed63e29b8139b2ed92e9af1c7dc78e8d",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 192,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/gitops": "pass"
      }
    },
    {
      "log": "local-root-final-31.json",
      "sha256": "0538090d52fa0d1ad861f6e89241667da296ef644876b5c4b0025027d8586683",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 112,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/globalbins": "pass"
      }
    },
    {
      "log": "local-root-final-32.json",
      "sha256": "4f3674049e8c5cda044895e6125519008063ccc18397f1b01ad42c4cb059d3db",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 2062,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/godriver": "pass"
      }
    },
    {
      "log": "local-root-final-33.json",
      "sha256": "d799e91396103a5306d64f71153cbeb9c6d2634ecead6e9576b8ff0fd3572db1",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 77,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/hashing": "pass"
      }
    },
    {
      "log": "local-root-final-34.json",
      "sha256": "03103329456ac075320b3861ec687fb94d450ecb2f5bd686843d3c6692b56e4a",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 133,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/hookapproval": "pass"
      }
    },
    {
      "log": "local-root-final-35.json",
      "sha256": "f642f39597c73dd60c9afaa7e0f04d55ee7d1dfb16928ebc172d4b908970868d",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 20,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/identifiers": "pass"
      }
    },
    {
      "log": "local-root-final-36.json",
      "sha256": "4d6cb345a6469094aabd0e7678c97eb64add30764bd3e89fc1413d2e92077979",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 49,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/identity": "pass"
      }
    },
    {
      "log": "local-root-final-37.json",
      "sha256": "cbf69229c5b1cecfac9d974dd3682c0db7ea751776b9779024046743a899f610",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 342,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/install/atomicity": "pass"
      }
    },
    {
      "log": "local-root-final-38.json",
      "sha256": "6a6882c99962b9fd9979537339970395dbadb1d28e6d3edcd83696aaab673610",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 274,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/interop": "pass"
      }
    },
    {
      "log": "local-root-final-39.json",
      "sha256": "13804f25467a0a67e3747d6648f5acda3cb09c685e37a45ec99c140d87e1a23a",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 24,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/locale": "pass"
      }
    },
    {
      "log": "local-root-final-40.json",
      "sha256": "5354be3d14339651e54ac48d3f900c5dee78e4cd3d6f5aa9fc368b6e68702285",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 76,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/managerlock": "pass"
      }
    },
    {
      "log": "local-root-final-41.json",
      "sha256": "4f02bbd796caec84237933b2b8db739b1063751e3fbfc024714e2cb72252edb8",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 456,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/manifest": "pass"
      }
    },
    {
      "log": "local-root-final-42.json",
      "sha256": "85724ac2bd97932c112911322422cb02260be2873d8b3a1ea4873a6953e3c5c2",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 2000,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/marker": "pass"
      }
    },
    {
      "log": "local-root-final-43.json",
      "sha256": "6207293187d4852c3eedf2feab87872e56ea05481acae941ddf5a0136212a7ad",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 32,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/mcp": "pass"
      }
    },
    {
      "log": "local-root-final-44.json",
      "sha256": "35bd3aab1d911da0df2627240835e7eee0ed5a960e91a6c4ee7830c633e424cb",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 201,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/moduleroots": "pass"
      }
    },
    {
      "log": "local-root-final-45.json",
      "sha256": "90c9616ebbf5cb85f95f3c87aba336b0d4d4fe0ac9ff1cd8db03e4e76736a62b",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 248,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/nodesource": "pass"
      }
    },
    {
      "log": "local-root-final-46.json",
      "sha256": "dbaaf855e177bfc830a2f9a5836e5fbb7af30d4c481efc3f945df670174cfb51",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 264,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/npmsource": "pass"
      }
    },
    {
      "log": "local-root-final-47.json",
      "sha256": "960c911f9259ca160ff6b7c05eb85e0051bc170e82a46faf0a3c7e25728a3124",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 12,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/opaquescan": "pass"
      }
    },
    {
      "log": "local-root-final-48.json",
      "sha256": "b4b57286c2d136cb5f847237fd3209559931d759efccc2998c0c5e71c6ced806",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 68,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/pathboundary": "pass"
      }
    },
    {
      "log": "local-root-final-49.json",
      "sha256": "2ad06589198821e02b45064d9c9ddb90d2e6bd995b3f90d740d3a13675b9093b",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 48,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/pkgversion": "pass"
      }
    },
    {
      "log": "local-root-final-51.json",
      "sha256": "d3237cdf93fc17f89e8f3a7d3f840b596ef038df686c43ac9db24294c7e6a6eb",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 40,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/privatedir": "pass"
      }
    },
    {
      "log": "local-root-final-52.json",
      "sha256": "7aa9a316e3c939122c9f79a4054055f17c9dba885c87ed4855ad2db5598f080b",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 104,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/protocoljson": "pass"
      }
    },
    {
      "log": "local-root-final-53.json",
      "sha256": "011894a9730d02a5f7de93260a267ad2c5df789079ce1eb64681a4b8a09b0e39",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 628,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/registry": "pass"
      }
    },
    {
      "log": "local-root-final-54.json",
      "sha256": "f05bbd8faa133decde81d68dac02c3679c9f50ca0f151e22ae4f2396ff998dcc",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 210,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/runtimestore": "pass"
      }
    },
    {
      "log": "local-root-final-55.json",
      "sha256": "15098fd3a2a02e81ede3e5446c700449dd82e50c9ff6b489b99f224b5f5edf61",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 308,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/rustsource": "pass"
      }
    },
    {
      "log": "local-root-final-56.json",
      "sha256": "f991ea8d0b39ff2f651001f68184c79e9d8da7afebe100b3ea9091ecaf36a4ac",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 273,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/scopes": "pass"
      }
    },
    {
      "log": "local-root-final-57.json",
      "sha256": "902afa481aafb712f98141242bb7ba70fd72058c33ec9ed0c4e746f57bb03eb4",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 233,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/scriptpolicy": "pass"
      }
    },
    {
      "log": "local-root-final-58.json",
      "sha256": "a5b52c4f64dce4e2440e2951ab5899cf364910cb0385cdd04ba6a5f9dd401cee",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 1101,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/scriptworker": "pass"
      }
    },
    {
      "log": "local-root-final-59.json",
      "sha256": "2ba61870d0c63c6c0c411afd89349f3618880e6a8fa36c421391f7acca52e646",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 931,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/shell": "pass"
      }
    },
    {
      "log": "local-root-final-60.json",
      "sha256": "438ef90134e5dd376d26dfb1619ceeca460828a17e6bbe1b87b16a5c92a56c76",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 92,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/skillcheck": "pass"
      }
    },
    {
      "log": "local-root-final-61.json",
      "sha256": "23515bd20f76adf580d6f1037c43629ab1479f90c904d57faf89c909512bfb5b",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 2069,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/skillspec": "pass"
      }
    },
    {
      "log": "local-root-final-62.json",
      "sha256": "07d20711d3e9b2124d171d9643d3351c8b1f7f9168e723de31ddf9dfa6c703c3",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 224,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/snapshot": "pass"
      }
    },
    {
      "log": "local-root-final-63.json",
      "sha256": "9e7b90026b14955c4930398121d23e2de10338136e58965a118f60255078fd42",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 320,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/sourcelock": "pass"
      }
    },
    {
      "log": "local-root-final-64.json",
      "sha256": "4263ce1133d487d4068d4ffc7a930ca326cfcbecfb0ee22dfa532a456747cbf7",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 132,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/staging": "pass"
      }
    },
    {
      "log": "local-root-final-65.json",
      "sha256": "f50b1e7a5c248f4abe0015983797b26181479797e254f947babada933705390a",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 36,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/stateread": "pass"
      }
    },
    {
      "log": "local-root-final-66.json",
      "sha256": "6bd3b506a42214b0f027736d6a4e27e1ff784bb1fbfcdfad77447e14318b744e",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 240,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/swiftpmbuild": "pass"
      }
    },
    {
      "log": "local-root-final-67.json",
      "sha256": "24a837de8b181031bccb5f6ff401165759ade7673a49c67afb4a0533ef768be8",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 1820,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/swiftpminterop": "pass"
      }
    },
    {
      "log": "local-root-final-68.json",
      "sha256": "a60c4066f6ff1d07076ac03a644dd33c75329bc7e6a522beb912aca2248bf3f7",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 172,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/swiftpmsource": "pass"
      }
    },
    {
      "log": "local-root-final-69.json",
      "sha256": "8b1aa28d7abb62e689510bdb27cae129e83de9ef55062215fa3c426e39c3641f",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 3,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/testcli": "skip"
      }
    },
    {
      "log": "local-root-final-70.json",
      "sha256": "f0aa503af6d3fa46b578cac3eb9bfec75b77ec9d4d7b7f83b7be58d3f601d039",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 3,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/testgitenv": "skip"
      }
    },
    {
      "log": "local-root-final-71.json",
      "sha256": "b389e96166a6ca85020cea694878fa8a56609fe57e690fb2f7e3ce0b1babd1eb",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 3,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/testtoolchain": "skip"
      }
    },
    {
      "log": "local-root-final-72.json",
      "sha256": "429f7af37eb824233fd8b0d2903c191a21ffa5fbf0c8b3a560e5b8fa036ede0d",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 1038,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/transaction": "pass"
      }
    },
    {
      "log": "local-root-final-73.json",
      "sha256": "7ad28561447c664f4b4651ecb09880df1252130fdd2425f983f89e710ad4766b",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 60,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/ui": "pass"
      }
    },
    {
      "log": "local-root-final-74.json",
      "sha256": "87ff56cf5b67ba3c2d0a3d230ffa825c6d6bb68adf9431f3b6463619920d70cd",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 12,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/verr": "pass"
      }
    },
    {
      "log": "local-root-final-75.json",
      "sha256": "875aea0cd4ab4a66f65c7240ea3731ea712fdabf0a47965675aab739d61f4b10",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 12,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/version": "pass"
      }
    },
    {
      "log": "local-root-final-76.json",
      "sha256": "d09fee0fe1cb95b53024c8ac96e3b50de301e0507f931f02b4675da2a1541778",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 40,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/whitelist": "pass"
      }
    },
    {
      "log": "local-root-final-77.json",
      "sha256": "c1e8019a0d38516c34a9dadb9c4a1b70ad37853158bc214860bf5a48ac01b598",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 216,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/yarnclassicsource": "pass"
      }
    },
    {
      "log": "local-root-final-78.json",
      "sha256": "7bca715bdd5f912b3f3839deca93092780454522717dda2bbdbd32246e396851",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 500,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/yarnmodernsource": "pass"
      }
    },
    {
      "log": "local-root-final-79.json",
      "sha256": "575607fba5b2bb021790d5961a8e8d169f283bd785412492fea360ec2c826840",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 168,
      "package_verdicts": {
        "github.com/relux-works/curator/tools/ciworkflow": "pass"
      }
    },
    {
      "log": "local-root-final-80.json",
      "sha256": "4f8d8090eac73d4c3543c4b621c6a422a2a5023d0a226c1bfbef37595e5c4d17",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 176,
      "package_verdicts": {
        "github.com/relux-works/curator/tools/goreleaserconfig": "pass"
      }
    },
    {
      "log": "local-pnpmsource-root-retry-01.json",
      "sha256": "4a81454f10b6d556e471239bd24816c1b34dfd8e894b1b5555e2d7095cf597a3",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 372,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/pnpmsource": "pass"
      }
    },
    {
      "log": "local-install-root-final-01.json",
      "sha256": "f3c0a2016b9aac85e123bddd960a42bf289242846a6d5b13267e357c1588a1d2",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 260,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/install": "pass"
      }
    },
    {
      "log": "local-install-root-final-02.json",
      "sha256": "3838edbbdc8816ec697a909b859b35852c75bcf0e6041a7a44893cf59cd64154",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 148,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/install": "pass"
      }
    },
    {
      "log": "local-install-root-final-03.json",
      "sha256": "86923558ac64b2110546df3b017cc5e81492a3dee7e38285ce5e763ca879d571",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 478,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/install": "pass"
      }
    },
    {
      "log": "local-install-root-final-04.json",
      "sha256": "d5f45dde1ec0d2c5340af88974dd9ddf7a27dd14db99f605e6ee4b6dde88d607",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 285,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/install": "pass"
      }
    },
    {
      "log": "local-install-root-final-05.json",
      "sha256": "560d5cafc426ba27660d2c8f1cd7a526524441ca37c8e271e9399c78bd43418e",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 632,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/install": "pass"
      }
    },
    {
      "log": "local-cli-03-part-01.json",
      "sha256": "ba3868ff874c0ecb347a9f0dbef31d8594433099e84c7392ee8da163ef1cf2f4",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 120,
      "package_verdicts": {
        "github.com/relux-works/curator/cmd/curator": "pass"
      }
    },
    {
      "log": "local-cli-03-part-02.json",
      "sha256": "96f4606e51a555bff4c4cf2018f3b053d70b08781e25787b3d20ddc54125e99f",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 199,
      "package_verdicts": {
        "github.com/relux-works/curator/cmd/curator": "pass"
      }
    },
    {
      "log": "local-cli-03-part-03.json",
      "sha256": "3195e371374c2ae2d2ae09d705dd9017db37cad7653c4aad5fb87a701613ea68",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 220,
      "package_verdicts": {
        "github.com/relux-works/curator/cmd/curator": "pass"
      }
    },
    {
      "log": "local-cli-03-part-04.json",
      "sha256": "d4272e05dfd8889b97ddf5656995d0b4dd79261b7d0582fa86a3b93643d24f86",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 44,
      "package_verdicts": {
        "github.com/relux-works/curator/cmd/curator": "pass"
      }
    },
    {
      "log": "local-cli-04-part-01.json",
      "sha256": "0aef6411cd0de3ec4bf7e6934eca95b8a028d7130093ccdb2a547dfa91e43097",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 505,
      "package_verdicts": {
        "github.com/relux-works/curator/cmd/curator": "pass"
      }
    },
    {
      "log": "local-cli-04-part-02.json",
      "sha256": "6a824255b43b231343a502618c6fd0d588c6bd926bbba7b5f1a61f4f7a331883",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 220,
      "package_verdicts": {
        "github.com/relux-works/curator/cmd/curator": "pass"
      }
    },
    {
      "log": "local-cli-04-part-03.json",
      "sha256": "6271b092e54455d27316d232417178dc221bdad6b8021c1bd8f9f22c4b46e9c8",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 283,
      "package_verdicts": {
        "github.com/relux-works/curator/cmd/curator": "pass"
      }
    },
    {
      "log": "local-cli-04-part-04.json",
      "sha256": "03927746a31ea862cb6d2fd49299f6f7ce3495c40c8857580180c94f68aecdca",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 75,
      "package_verdicts": {
        "github.com/relux-works/curator/cmd/curator": "pass"
      }
    },
    {
      "log": "local-cli-05-part-01.json",
      "sha256": "6c42de0b4fd803f1de0154e0720f6150ac823d31c5d15665aea692029766eab5",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 860,
      "package_verdicts": {
        "github.com/relux-works/curator/cmd/curator": "pass"
      }
    },
    {
      "log": "local-cli-05-part-02.json",
      "sha256": "9a21ece8e71418741279fba7fcdf76f16101ae5ea1a28dc733568c383a5c02d2",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 205,
      "package_verdicts": {
        "github.com/relux-works/curator/cmd/curator": "pass"
      }
    },
    {
      "log": "local-cli-05-part-03.json",
      "sha256": "5220d414bed32d6c0e077386f01982ac68f5f3a1c38bc4e3593c41b0a19aaaee",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 284,
      "package_verdicts": {
        "github.com/relux-works/curator/cmd/curator": "pass"
      }
    },
    {
      "log": "local-cli-05-part-04.json",
      "sha256": "698efa94c96ec48e782b496674f45c54182bda6862b174e2f9b813d57bcffaf2",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 40,
      "package_verdicts": {
        "github.com/relux-works/curator/cmd/curator": "pass"
      }
    },
    {
      "log": "local-cli-06-part-01.json",
      "sha256": "78ecbc0ad18120a13a8dcf96243ada27491e1be8955f5f6c06e9c4d110680bb7",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 149,
      "package_verdicts": {
        "github.com/relux-works/curator/cmd/curator": "pass"
      }
    },
    {
      "log": "local-install-03-part-01.json",
      "sha256": "665e04f4069ddcc9bfbf26d79d2875778ce4b9eab90a7507ea34b00b15a448e1",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 144,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/install": "pass"
      }
    },
    {
      "log": "local-install-03-part-02.json",
      "sha256": "b021d68140c0d51909c1fb741eab65f32ed4a82cc7a5109429d3c756203c4b42",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 136,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/install": "pass"
      }
    },
    {
      "log": "local-install-03-part-03.json",
      "sha256": "3cdba15e17c3ebc3123a2f60b722fb5b0ce63239e1f25d7ff0fbd6fe4b1c1f42",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 108,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/install": "pass"
      }
    },
    {
      "log": "local-install-04-part-01.json",
      "sha256": "99dc6ce05a4d103839e6158dcb0ed5c1c39a1139cb3d9d24f3d1fd1b2109754e",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 407,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/install": "pass"
      }
    },
    {
      "log": "local-install-04-part-02.json",
      "sha256": "787cd8b26175acfbffb08ba7971d347abd5fe20f12e44520bdf5ec6d7047ebd6",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 221,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/install": "pass"
      }
    },
    {
      "log": "local-install-04-part-03.json",
      "sha256": "43ee9d4766ece0a2e0b8433328ff27741c25203a1adedb07058a55503989f5d3",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 92,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/install": "pass"
      }
    },
    {
      "log": "local-install-05-part-01.json",
      "sha256": "0f2f05e9b8b40a1ec63f56c37e6f9e0d3ba1dff263b4764c1e64c9ba6b29e1a1",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 301,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/install": "pass"
      }
    },
    {
      "log": "local-install-05-part-02.json",
      "sha256": "0d023391b86741799fb4f8021de9a5887d695dcdf803aa77c708e667cdb3d4f3",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 272,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/install": "pass"
      }
    },
    {
      "log": "local-install-05-part-03.json",
      "sha256": "ae3dee3cc7c0e54bab299e23b0b62959539b872b30868544089166e36630e212",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 85,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/install": "pass"
      }
    },
    {
      "log": "local-install-06-part-01.json",
      "sha256": "e8e583204d221f85821419bcb42959affbfdc16f236479c7672d9ffaad725f66",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 229,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/install": "pass"
      }
    },
    {
      "log": "local-install-06-part-02.json",
      "sha256": "91b9d240fe463b73ee8c84fd08805022aec074832e688309feeb06b03d688c29",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 192,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/install": "pass"
      }
    },
    {
      "log": "local-install-06-part-03.json",
      "sha256": "fe1e568073381fe70795cf75b58afbe93db8e4173c995e6f19ab37c7085d8436",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 88,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/install": "pass"
      }
    },
    {
      "log": "local-install-07-part-01.json",
      "sha256": "183373811a8a48c2d5ca811fd80cb3d42b3ac2b181e6afb3ee88409303a50180",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 16,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/install": "pass"
      }
    },
    {
      "log": "local-envprofile-final-01.json",
      "sha256": "19e53c80cc2f22659d70b920b23a7a0b809c2e5a5d38efb477ea3efddabc024c",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 234,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/envprofile": "pass"
      }
    },
    {
      "log": "local-envprofile-final-02.json",
      "sha256": "da0c49fdde08c9c7fdec698c31991120b91a65469aad48b96619a19521c6ef1d",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 365,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/envprofile": "pass"
      }
    },
    {
      "log": "local-envprofile-final-03.json",
      "sha256": "fc290d32a77d209ac19b8c98ad6892544bfe34613436d94fd8a6ee3d16b662b2",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 141,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/envprofile": "pass"
      }
    },
    {
      "log": "local-envprofile-final-04.json",
      "sha256": "a9e5cb910c523225ed4c491ed007bae5a5b4e8e30465b53b62158d7e2c147763",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 104,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/envprofile": "pass"
      }
    },
    {
      "log": "local-envprofile-final-05.json",
      "sha256": "64a86de6d17462830d9ac8654c34f71d6f30dab6ebbeac8a23deaff57df5e548",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 405,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/envprofile": "pass"
      }
    },
    {
      "log": "local-envprofile-final-06.json",
      "sha256": "a10ae3c85d981cd70b973db7dbfd0f8ed1e755540acdd7a617e1efcdc2032089",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 318,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/envprofile": "pass"
      }
    },
    {
      "log": "local-envprofile-final-07.json",
      "sha256": "edbf24543d7a50825d94bfc991a69705cc8c3c2bb52f8b4ea7e4f32833234b54",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 196,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/envprofile": "pass"
      }
    },
    {
      "log": "local-envprofile-final-08.json",
      "sha256": "46b6190e9b7fda764d9b7cb404777322d1d5f0d1fc65557331a6735735050600",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 256,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/envprofile": "pass"
      }
    },
    {
      "log": "local-envprofile-final-09.json",
      "sha256": "8c2f268a8fcb73181e19c23ac0b514a84173513db41c711acd4f8487a17941cd",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 314,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/envprofile": "pass"
      }
    },
    {
      "log": "local-envprofile-final-10.json",
      "sha256": "045b3344e5c89fdb7a3e76787a22240665cddd4e750e9503c57d3fcb3f133aa2",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 224,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/envprofile": "pass"
      }
    },
    {
      "log": "local-envprofile-final-11.json",
      "sha256": "9ee6b85066d1cb926e81a3e53f1c31a18c0aa65be342ad0ff228a4f7fa7255b8",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 232,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/envprofile": "pass"
      }
    },
    {
      "log": "local-envprofile-final-12.json",
      "sha256": "b757996b87bbebf371c1e628e06b45ca1761240e4b9df430992b6df9a20e553b",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 293,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/envprofile": "pass"
      }
    },
    {
      "log": "local-envprofile-final-13.json",
      "sha256": "0b8afdf2088bb8f01e9a49eefd81fc163148efabc4e5f96825c11807bfe0e5f8",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 322,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/envprofile": "pass"
      }
    },
    {
      "log": "local-envprofile-final-14.json",
      "sha256": "e71e3fa22e20af74c2062f1ee5ff826fff42194d34ad78e8648c9441de214ed2",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 296,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/envprofile": "pass"
      }
    },
    {
      "log": "local-envprofile-final-15.json",
      "sha256": "2104a9e64a06324d22a3f1bdc3bba2161c4020d3da7236608b5300b4493cf8dc",
      "command_exit": 0,
      "excluded_package": null,
      "selected_events": 165,
      "package_verdicts": {
        "github.com/relux-works/curator/internal/envprofile": "pass"
      }
    }
  ]
}

```

## batch-results-01.json

```text
[
  {
    "name": "cli-status-retry-01",
    "exit_code": 0
  },
  {
    "name": "install-02",
    "exit_code": 0
  },
  {
    "name": "cli-01-clean",
    "exit_code": 0
  },
  {
    "name": "install-03-part-01",
    "exit_code": 0
  },
  {
    "name": "install-03-part-02",
    "exit_code": 0
  },
  {
    "name": "install-03-part-03",
    "exit_code": 0
  },
  {
    "name": "envprofile-01",
    "exit_code": 1
  },
  {
    "name": "envprofile-02",
    "exit_code": 0
  },
  {
    "name": "envprofile-01-part-01",
    "exit_code": 0
  },
  {
    "name": "envprofile-01-part-02",
    "exit_code": 0
  },
  {
    "name": "envprofile-01-part-03",
    "exit_code": 0
  },
  {
    "name": "envprofile-01-part-04",
    "exit_code": 0
  },
  {
    "name": "envprofile-01-part-05",
    "exit_code": 0
  },
  {
    "name": "envprofile-03-part-01",
    "exit_code": 0
  },
  {
    "name": "cli-02",
    "exit_code": 0
  },
  {
    "name": "envprofile-03-part-02",
    "exit_code": 0
  },
  {
    "name": "install-04-part-01",
    "exit_code": 0
  },
  {
    "name": "install-04-part-02",
    "exit_code": 0
  },
  {
    "name": "install-04-part-03",
    "exit_code": 0
  },
  {
    "name": "envprofile-03-part-03",
    "exit_code": 0
  },
  {
    "name": "install-05-part-01",
    "exit_code": 0
  },
  {
    "name": "envprofile-03-part-04",
    "exit_code": 0
  },
  {
    "name": "install-05-part-02",
    "exit_code": 0
  },
  {
    "name": "envprofile-03-part-05",
    "exit_code": 0
  },
  {
    "name": "install-05-part-03",
    "exit_code": 0
  },
  {
    "name": "cli-03",
    "exit_code": 1
  },
  {
    "name": "install-06-part-01",
    "exit_code": 0
  },
  {
    "name": "install-06-part-02",
    "exit_code": 0
  },
  {
    "name": "install-06-part-03",
    "exit_code": 0
  },
  {
    "name": "install-07-part-01",
    "exit_code": 0
  }
]

```

## batch-results-02.json

```text
[
  {
    "name": "cli-signer-retry-01",
    "exit_code": 0
  },
  {
    "name": "cli-03-part-01",
    "exit_code": 0
  },
  {
    "name": "cli-03-part-02",
    "exit_code": 0
  },
  {
    "name": "cli-03-part-03",
    "exit_code": 0
  },
  {
    "name": "cli-03-part-04",
    "exit_code": 0
  },
  {
    "name": "cli-04-part-01",
    "exit_code": 0
  },
  {
    "name": "cli-04-part-02",
    "exit_code": 0
  },
  {
    "name": "cli-04-part-03",
    "exit_code": 0
  },
  {
    "name": "cli-04-part-04",
    "exit_code": 0
  },
  {
    "name": "cli-05-part-01",
    "exit_code": 0
  },
  {
    "name": "cli-05-part-02",
    "exit_code": 0
  },
  {
    "name": "cli-05-part-03",
    "exit_code": 0
  },
  {
    "name": "cli-05-part-04",
    "exit_code": 0
  },
  {
    "name": "cli-06-part-01",
    "exit_code": 0
  }
]

```

## batch-results-03.json

```text
[
  {
    "name": "envprofile-final-03",
    "exit_code": 0
  },
  {
    "name": "envprofile-final-02",
    "exit_code": 0
  },
  {
    "name": "envprofile-final-01",
    "exit_code": 0
  },
  {
    "name": "envprofile-final-06",
    "exit_code": 0
  },
  {
    "name": "envprofile-final-04",
    "exit_code": 0
  },
  {
    "name": "envprofile-final-09",
    "exit_code": 0
  },
  {
    "name": "envprofile-final-05",
    "exit_code": 0
  },
  {
    "name": "envprofile-final-07",
    "exit_code": 0
  },
  {
    "name": "envprofile-final-10",
    "exit_code": 0
  },
  {
    "name": "envprofile-final-12",
    "exit_code": 0
  },
  {
    "name": "envprofile-final-08",
    "exit_code": 0
  },
  {
    "name": "envprofile-final-11",
    "exit_code": 0
  },
  {
    "name": "envprofile-final-15",
    "exit_code": 0
  },
  {
    "name": "envprofile-final-13",
    "exit_code": 0
  },
  {
    "name": "envprofile-final-14",
    "exit_code": 0
  }
]

```

## batch-results-04.json

```text
[
  {
    "name": "root-final-01",
    "target": "github.com/relux-works/curator/cmd/curator-spec-pin",
    "exit_code": 0
  },
  {
    "name": "root-final-02",
    "target": "github.com/relux-works/curator/internal/adapters",
    "exit_code": 0
  },
  {
    "name": "root-final-03",
    "target": "github.com/relux-works/curator/internal/artifactpolicy/conformance",
    "exit_code": 0
  },
  {
    "name": "root-final-05",
    "target": "github.com/relux-works/curator/internal/buildcache",
    "exit_code": 0
  },
  {
    "name": "root-final-04",
    "target": "github.com/relux-works/curator/internal/audit",
    "exit_code": 0
  },
  {
    "name": "root-final-06",
    "target": "github.com/relux-works/curator/internal/buildmeta",
    "exit_code": 0
  },
  {
    "name": "root-final-08",
    "target": "github.com/relux-works/curator/internal/buildsource",
    "exit_code": 0
  },
  {
    "name": "install-root-final-01",
    "target": "./internal/install",
    "exit_code": 0
  },
  {
    "name": "install-root-final-02",
    "target": "./internal/install",
    "exit_code": 0
  },
  {
    "name": "install-root-final-03",
    "target": "./internal/install",
    "exit_code": 0
  },
  {
    "name": "install-root-final-04",
    "target": "./internal/install",
    "exit_code": 0
  },
  {
    "name": "root-final-10",
    "target": "github.com/relux-works/curator/internal/closure",
    "exit_code": 0
  },
  {
    "name": "root-final-12",
    "target": "github.com/relux-works/curator/internal/closuregraph",
    "exit_code": 0
  },
  {
    "name": "root-final-14",
    "target": "github.com/relux-works/curator/internal/conformancecoverage",
    "exit_code": 0
  },
  {
    "name": "root-final-16",
    "target": "github.com/relux-works/curator/internal/contextlock",
    "exit_code": 0
  },
  {
    "name": "root-final-18",
    "target": "github.com/relux-works/curator/internal/contextpkg",
    "exit_code": 0
  },
  {
    "name": "root-final-20",
    "target": "github.com/relux-works/curator/internal/contextstore",
    "exit_code": 0
  },
  {
    "name": "root-final-22",
    "target": "github.com/relux-works/curator/internal/devsub",
    "exit_code": 0
  },
  {
    "name": "root-final-24",
    "target": "github.com/relux-works/curator/internal/envfragment",
    "exit_code": 0
  },
  {
    "name": "root-final-26",
    "target": "github.com/relux-works/curator/internal/envregistry",
    "exit_code": 0
  },
  {
    "name": "root-final-28",
    "target": "github.com/relux-works/curator/internal/gitcred",
    "exit_code": 0
  },
  {
    "name": "root-final-30",
    "target": "github.com/relux-works/curator/internal/gitops",
    "exit_code": 0
  },
  {
    "name": "root-final-07",
    "target": "github.com/relux-works/curator/internal/buildrepo",
    "exit_code": 0
  },
  {
    "name": "root-final-09",
    "target": "github.com/relux-works/curator/internal/capabilities",
    "exit_code": 0
  },
  {
    "name": "root-final-11",
    "target": "github.com/relux-works/curator/internal/closureexec",
    "exit_code": 0
  },
  {
    "name": "install-root-final-05",
    "target": "./internal/install",
    "exit_code": 0
  },
  {
    "name": "root-final-13",
    "target": "github.com/relux-works/curator/internal/config",
    "exit_code": 0
  },
  {
    "name": "root-final-15",
    "target": "github.com/relux-works/curator/internal/contextaudit",
    "exit_code": 0
  },
  {
    "name": "root-final-17",
    "target": "github.com/relux-works/curator/internal/contextmaterialize",
    "exit_code": 0
  },
  {
    "name": "root-final-19",
    "target": "github.com/relux-works/curator/internal/contextresolve",
    "exit_code": 0
  },
  {
    "name": "root-final-32",
    "target": "github.com/relux-works/curator/internal/godriver",
    "exit_code": 0
  },
  {
    "name": "root-final-34",
    "target": "github.com/relux-works/curator/internal/hookapproval",
    "exit_code": 0
  },
  {
    "name": "root-final-36",
    "target": "github.com/relux-works/curator/internal/identity",
    "exit_code": 0
  },
  {
    "name": "root-final-38",
    "target": "github.com/relux-works/curator/internal/interop",
    "exit_code": 0
  },
  {
    "name": "root-final-40",
    "target": "github.com/relux-works/curator/internal/managerlock",
    "exit_code": 0
  },
  {
    "name": "root-final-42",
    "target": "github.com/relux-works/curator/internal/marker",
    "exit_code": 0
  },
  {
    "name": "root-final-44",
    "target": "github.com/relux-works/curator/internal/moduleroots",
    "exit_code": 0
  },
  {
    "name": "root-final-46",
    "target": "github.com/relux-works/curator/internal/npmsource",
    "exit_code": 0
  },
  {
    "name": "root-final-48",
    "target": "github.com/relux-works/curator/internal/pathboundary",
    "exit_code": 0
  },
  {
    "name": "root-final-50",
    "target": "github.com/relux-works/curator/internal/pnpmsource",
    "exit_code": 0
  },
  {
    "name": "root-final-52",
    "target": "github.com/relux-works/curator/internal/protocoljson",
    "exit_code": 0
  },
  {
    "name": "root-final-54",
    "target": "github.com/relux-works/curator/internal/runtimestore",
    "exit_code": 0
  },
  {
    "name": "root-final-56",
    "target": "github.com/relux-works/curator/internal/scopes",
    "exit_code": 0
  },
  {
    "name": "root-final-58",
    "target": "github.com/relux-works/curator/internal/scriptworker",
    "exit_code": 0
  },
  {
    "name": "root-final-60",
    "target": "github.com/relux-works/curator/internal/skillcheck",
    "exit_code": 0
  },
  {
    "name": "root-final-62",
    "target": "github.com/relux-works/curator/internal/snapshot",
    "exit_code": 0
  },
  {
    "name": "root-final-64",
    "target": "github.com/relux-works/curator/internal/staging",
    "exit_code": 0
  },
  {
    "name": "root-final-66",
    "target": "github.com/relux-works/curator/internal/swiftpmbuild",
    "exit_code": 0
  },
  {
    "name": "root-final-21",
    "target": "github.com/relux-works/curator/internal/crossconformance",
    "exit_code": 0
  },
  {
    "name": "root-final-23",
    "target": "github.com/relux-works/curator/internal/envfiles",
    "exit_code": 0
  },
  {
    "name": "root-final-25",
    "target": "github.com/relux-works/curator/internal/envmarker",
    "exit_code": 0
  },
  {
    "name": "root-final-27",
    "target": "github.com/relux-works/curator/internal/envsnapshot",
    "exit_code": 0
  },
  {
    "name": "root-final-29",
    "target": "github.com/relux-works/curator/internal/gitignore",
    "exit_code": 0
  },
  {
    "name": "root-final-31",
    "target": "github.com/relux-works/curator/internal/globalbins",
    "exit_code": 0
  },
  {
    "name": "root-final-33",
    "target": "github.com/relux-works/curator/internal/hashing",
    "exit_code": 0
  },
  {
    "name": "root-final-35",
    "target": "github.com/relux-works/curator/internal/identifiers",
    "exit_code": 0
  },
  {
    "name": "root-final-68",
    "target": "github.com/relux-works/curator/internal/swiftpmsource",
    "exit_code": 0
  },
  {
    "name": "root-final-70",
    "target": "github.com/relux-works/curator/internal/testgitenv",
    "exit_code": 0
  },
  {
    "name": "root-final-37",
    "target": "github.com/relux-works/curator/internal/install/atomicity",
    "exit_code": 0
  },
  {
    "name": "root-final-39",
    "target": "github.com/relux-works/curator/internal/locale",
    "exit_code": 0
  },
  {
    "name": "root-final-41",
    "target": "github.com/relux-works/curator/internal/manifest",
    "exit_code": 0
  },
  {
    "name": "root-final-43",
    "target": "github.com/relux-works/curator/internal/mcp",
    "exit_code": 0
  },
  {
    "name": "root-final-45",
    "target": "github.com/relux-works/curator/internal/nodesource",
    "exit_code": 0
  },
  {
    "name": "root-final-47",
    "target": "github.com/relux-works/curator/internal/opaquescan",
    "exit_code": 0
  },
  {
    "name": "root-final-49",
    "target": "github.com/relux-works/curator/internal/pkgversion",
    "exit_code": 0
  },
  {
    "name": "root-final-72",
    "target": "github.com/relux-works/curator/internal/transaction",
    "exit_code": 0
  },
  {
    "name": "root-final-51",
    "target": "github.com/relux-works/curator/internal/privatedir",
    "exit_code": 0
  },
  {
    "name": "root-final-74",
    "target": "github.com/relux-works/curator/internal/verr",
    "exit_code": 0
  },
  {
    "name": "root-final-76",
    "target": "github.com/relux-works/curator/internal/whitelist",
    "exit_code": 0
  },
  {
    "name": "root-final-78",
    "target": "github.com/relux-works/curator/internal/yarnmodernsource",
    "exit_code": 0
  },
  {
    "name": "root-final-80",
    "target": "github.com/relux-works/curator/tools/goreleaserconfig",
    "exit_code": 0
  },
  {
    "name": "root-final-53",
    "target": "github.com/relux-works/curator/internal/registry",
    "exit_code": 0
  },
  {
    "name": "root-final-55",
    "target": "github.com/relux-works/curator/internal/rustsource",
    "exit_code": 0
  },
  {
    "name": "root-final-57",
    "target": "github.com/relux-works/curator/internal/scriptpolicy",
    "exit_code": 0
  },
  {
    "name": "root-final-59",
    "target": "github.com/relux-works/curator/internal/shell",
    "exit_code": 0
  },
  {
    "name": "root-final-61",
    "target": "github.com/relux-works/curator/internal/skillspec",
    "exit_code": 0
  },
  {
    "name": "root-final-63",
    "target": "github.com/relux-works/curator/internal/sourcelock",
    "exit_code": 0
  },
  {
    "name": "root-final-65",
    "target": "github.com/relux-works/curator/internal/stateread",
    "exit_code": 0
  },
  {
    "name": "root-final-67",
    "target": "github.com/relux-works/curator/internal/swiftpminterop",
    "exit_code": 0
  },
  {
    "name": "root-final-69",
    "target": "github.com/relux-works/curator/internal/testcli",
    "exit_code": 0
  },
  {
    "name": "root-final-71",
    "target": "github.com/relux-works/curator/internal/testtoolchain",
    "exit_code": 0
  },
  {
    "name": "root-final-73",
    "target": "github.com/relux-works/curator/internal/ui",
    "exit_code": 0
  },
  {
    "name": "root-final-75",
    "target": "github.com/relux-works/curator/internal/version",
    "exit_code": 0
  },
  {
    "name": "root-final-77",
    "target": "github.com/relux-works/curator/internal/yarnclassicsource",
    "exit_code": 0
  },
  {
    "name": "root-final-79",
    "target": "github.com/relux-works/curator/tools/ciworkflow",
    "exit_code": 0
  }
]

```

## fallback-mutants-01.log

```text
=== RUN   TestManagedModesNarrowingMutants
=== RUN   TestManagedModesNarrowingMutants/single-fallback-0644
    managed_modes_mutants_unix_test.go:97: expected-red go test exit 1:
        === RUN   TestManagedFallbackPermissions
        === RUN   TestManagedFallbackPermissions/tree-false
            managed_modes_unix_test.go:65: copy mode = 0644, want 0600 (permissions must not widen)
        --- FAIL: TestManagedFallbackPermissions (0.03s)
            --- FAIL: TestManagedFallbackPermissions/tree-false (0.03s)
        FAIL
        FAIL	github.com/relux-works/curator/internal/envprofile	0.365s
        FAIL
=== RUN   TestManagedModesNarrowingMutants/executable-tree-fallback-0755
    managed_modes_mutants_unix_test.go:97: expected-red go test exit 1:
        === RUN   TestManagedFallbackPermissions
        === RUN   TestManagedFallbackPermissions/tree-true
            managed_modes_unix_test.go:65: 0755 mode = 0755, want 0700 (permissions must not widen)
        --- FAIL: TestManagedFallbackPermissions (0.02s)
            --- FAIL: TestManagedFallbackPermissions/tree-true (0.02s)
        FAIL
        FAIL	github.com/relux-works/curator/internal/envprofile	0.354s
        FAIL
=== RUN   TestManagedModesNarrowingMutants/private-seed-0644
    managed_modes_mutants_unix_test.go:97: expected-red go test exit 1:
        === RUN   TestManagedSeedPreservesNativePermissions
        === RUN   TestManagedSeedPreservesNativePermissions/0600
            managed_modes_unix_test.go:95: config.toml mode = 0644, want 0600 (permissions must not widen)
        --- FAIL: TestManagedSeedPreservesNativePermissions (0.81s)
            --- FAIL: TestManagedSeedPreservesNativePermissions/0600 (0.81s)
        FAIL
        FAIL	github.com/relux-works/curator/internal/envprofile	1.146s
        FAIL
=== RUN   TestManagedModesNarrowingMutants/claude-trust-0644
    managed_modes_mutants_unix_test.go:97: expected-red go test exit 1:
        === RUN   TestManagedAuthoredFilesArePrivate
            managed_modes_unix_test.go:170: .claude.json mode = 0644, want 0600 (permissions must not widen)
        --- FAIL: TestManagedAuthoredFilesArePrivate (0.95s)
        FAIL
        FAIL	github.com/relux-works/curator/internal/envprofile	1.281s
        FAIL
=== RUN   TestManagedModesNarrowingMutants/codex-directory-drift-admitted
    managed_modes_mutants_unix_test.go:97: expected-red go test exit 1:
        === RUN   TestManagedHomeRepairTightensDirectories
        === RUN   TestManagedHomeRepairTightensDirectories/codex_cli
            managed_modes_unix_test.go:138: wide home must be stale: <nil>
        --- FAIL: TestManagedHomeRepairTightensDirectories (1.13s)
            --- FAIL: TestManagedHomeRepairTightensDirectories/codex_cli (1.13s)
        FAIL
        FAIL	github.com/relux-works/curator/internal/envprofile	1.466s
        FAIL
=== RUN   TestManagedModesNarrowingMutants/codex-directory-not-tightened
    managed_modes_mutants_unix_test.go:97: expected-red go test exit 1:
        === RUN   TestManagedHomeRepairTightensDirectories
        === RUN   TestManagedHomeRepairTightensDirectories/codex_cli
            managed_modes_unix_test.go:147: environment_repair_failed: managed directory permissions: <temporary-path> is not a private owner-only directory; managed directory permissions: <temporary-path> is not a private owner-only directory
        --- FAIL: TestManagedHomeRepairTightensDirectories (1.52s)
            --- FAIL: TestManagedHomeRepairTightensDirectories/codex_cli (1.52s)
        FAIL
        FAIL	github.com/relux-works/curator/internal/envprofile	1.858s
        FAIL
--- PASS: TestManagedModesNarrowingMutants (25.04s)
    --- PASS: TestManagedModesNarrowingMutants/single-fallback-0644 (2.91s)
    --- PASS: TestManagedModesNarrowingMutants/executable-tree-fallback-0755 (2.76s)
    --- PASS: TestManagedModesNarrowingMutants/private-seed-0644 (3.58s)
    --- PASS: TestManagedModesNarrowingMutants/claude-trust-0644 (3.81s)
    --- PASS: TestManagedModesNarrowingMutants/codex-directory-drift-admitted (3.74s)
    --- PASS: TestManagedModesNarrowingMutants/codex-directory-not-tightened (8.24s)
=== RUN   TestManagedFallbackPermissions
=== RUN   TestManagedFallbackPermissions/tree-false
=== RUN   TestManagedFallbackPermissions/tree-true
--- PASS: TestManagedFallbackPermissions (0.06s)
    --- PASS: TestManagedFallbackPermissions/tree-false (0.02s)
    --- PASS: TestManagedFallbackPermissions/tree-true (0.04s)
PASS
ok  	github.com/relux-works/curator/internal/envprofile	25.543s

```

## root-identity-control-01.log

```text
Traceback (most recent call last):
  File "<stdin>", line 4, in <module>
  File "root-identity-control", line 27, in <module>
AssertionError: ('local-packages-02.json', 'stream predates immutable fixture root')

```

## spec-root-identity-03.log

```text
{"pin": "43bf0a2506d5c354a73bbc3ea4623d4653db10c7", "conformance_blobs": 1423, "v1_blobs": 1295, "mismatches": []}

```

## local-platform/skips-observed.tsv

```text
package	test	class	verdict	reason
internal/buildrepo	TestResolvedTransportWindowsRefusesBeforeAnyProcess	platform-control	allowed-platform-control	windows refusal probe is exercised on Windows
internal/buildrepo	TestVerifyPublishedReleasePinWhenConformanceRootIsSupplied	root-content	allowed-root-content	the supplied conformance root publishes a release other than the pinned v1.0.0-rc.8
internal/buildsource	TestBuildSourceConformanceVectors/duplicate-build-source-path	host-capability	allowed-host-capability	this host folds the alias into one directory member, so no duplicate encoded path is reachable
internal/buildsource	TestBuildSourceConformanceVectors/invalid-unicode-build-source-path	host-capability	allowed-host-capability	this host cannot create a member whose name is not valid Unicode
internal/envprofile	TestFoldedPathsAreSourceInvalid	host-capability	tolerated-by-ledger	per-directory case sensitivity folds entries on this host; the collision gate runs on case-sensitive lanes
internal/gitops	TestExtractAdmitsDirectoryComponentFoldWhenCaseSensitive	host-capability	tolerated-by-ledger	test filesystem lacks case sensitivity: Dir and dir fold into one directory, no coexistence to assert
internal/godriver	TestBuildFailsClosedWhenTheGoChildCannotStart	platform-control	tolerated-by-ledger	active-process-count-limit is available only on Windows in rc5-native-control-inventory-v1
internal/godriver	TestFingerprintRejectsInvalidUnicodePath	host-capability	allowed-host-capability	filesystem rejects invalid UTF-8 names: open <temporary-path> illegal byte sequence
internal/godriver	TestFixedEnvironmentAndFiveDirectArgvFormsVector/linux-amd64	platform-control	allowed-platform-control	fixed environment case is exercised on its native runner
internal/godriver	TestFixedEnvironmentAndFiveDirectArgvFormsVector/windows-amd64	platform-control	allowed-platform-control	fixed environment case is exercised on its native runner
internal/godriver	TestRealGoV1ModuleRootsBuildIsBoundedAndNotLaunched	opt-in	allowed-opt-in	set CURATOR_REAL_GO_BUILD_TEST=1 for the bounded native go-v1 integration
internal/godriver	TestRealGoV1VendoredBuildIsBoundedAndNotLaunched	opt-in	allowed-opt-in	set CURATOR_REAL_GO_BUILD_TEST=1 for the bounded native go-v1 integration
internal/godriver	TestRealTrustedGoProbe	opt-in	allowed-opt-in	set CURATOR_REAL_GO_TEST=1 for the bounded native toolchain probe
internal/godriver	TestToolchainIdentityVectors/invalid-unicode-toolchain-path	host-capability	allowed-host-capability	this host cannot create a member whose name is not valid Unicode
internal/hashing	TestCrossCheckTree	opt-in	tolerated-by-ledger	cross-check env not set
internal/hookapproval	TestWindowsDriveIdentity	platform-control	allowed-platform-control	Windows drive identity is exercised on Windows
internal/install	TestAcquireDraftNetworkWindowsRefusesBeforeAnyProcess	platform-control	allowed-platform-control	windows refusal probe is exercised on Windows
internal/install	TestInstallHelperProcess	helper-process	allowed-helper-process	helper process
internal/scriptworker	TestLandlockConstructionConfirmsTypedRules	platform-control	tolerated-by-ledger	Landlock rights are a Linux-only control in script-worker-v1-native-control-inventory-v1
internal/scriptworker	TestLandlockProbeCancellationKillsProcessGroup	platform-control	allowed-platform-control	Landlock probe cancellation is a Linux-only control in script-worker-v1-native-control-inventory-v1 (requires process groups and a subreaper)
internal/scriptworker	TestLinuxHostConditionalUnavailableSucceeds	platform-control	tolerated-by-ledger	host-conditional Linux probing is a Linux-only control in script-worker-v1-native-control-inventory-v1
internal/scriptworker	TestLinuxHostProbeAndEvidenceAreConsistent	platform-control	tolerated-by-ledger	host-conditional Linux probing is a Linux-only control in script-worker-v1-native-control-inventory-v1
internal/scriptworker	TestLinuxLandlockConfinementMatchesProbe	platform-control	tolerated-by-ledger	Landlock rights are a Linux-only control in script-worker-v1-native-control-inventory-v1
internal/scriptworker	TestLinuxLandlockProbeChildAnswersAtTheBoundary	platform-control	tolerated-by-ledger	Landlock rights are a Linux-only control in script-worker-v1-native-control-inventory-v1
internal/scriptworker	TestLinuxLandlockProbeFindsPresentControls	platform-control	tolerated-by-ledger	Landlock rights are a Linux-only control in script-worker-v1-native-control-inventory-v1
internal/scriptworker	TestLinuxMissingDerivedPathRefusesWriteConfinement	platform-control	tolerated-by-ledger	Landlock rights are a Linux-only control in script-worker-v1-native-control-inventory-v1
internal/scriptworker	TestLinuxWriteConfinementDirectoryMutationMatchesProbe	platform-control	tolerated-by-ledger	Landlock rights are a Linux-only control in script-worker-v1-native-control-inventory-v1
internal/scriptworker	TestLinuxWriteConfinementTruncateMatchesProbe	platform-control	tolerated-by-ledger	Landlock rights are a Linux-only control in script-worker-v1-native-control-inventory-v1
internal/scriptworker	TestWindowsJobLimitsAppliedAndConfirmed	platform-control	tolerated-by-ledger	Job Object limits are a Windows-only control in script-worker-v1-native-control-inventory-v1
internal/scriptworker	TestWindowsRealInterpretersRunDeclaredExec	platform-control	tolerated-by-ledger	Windows-only control in script-worker-v1-native-control-inventory-v1 is exercised on the Windows runner
internal/shell	TestPowerShellHookRunsOnEveryPrompt	platform-control	tolerated-by-ledger	PowerShell prompt integration is exercised on Windows
internal/shell	TestShellHookTrustNativeRecordAuthorizesMSYSSpelling	platform-control	allowed-platform-control	native/MSYS spelling identity is exercised on Windows
internal/yarnmodernsource	TestModernRuntimeCyclesThroughRealPinnedYarn	opt-in	allowed-opt-in	set CURATOR_TEST_YARN_MODERN_JS to the @yarnpkg/cli-dist 4.9.2 bin/yarn.js integration tool
internal/yarnmodernsource	TestN01RealPinnedYarnPnPInvokeThroughVerifiedExecutor	opt-in	allowed-opt-in	set CURATOR_TEST_YARN_MODERN_JS to the @yarnpkg/cli-dist 4.9.2 bin/yarn.js integration tool
internal/yarnmodernsource	TestN10RealPinnedYarnRemotePeerContextsThroughVerifiedExecutor	opt-in	allowed-opt-in	set CURATOR_TEST_YARN_MODERN_JS to the @yarnpkg/cli-dist 4.9.2 bin/yarn.js integration tool
internal/yarnmodernsource	TestN10RealPinnedYarnWorkspacePeerPnPInvokeThroughVerifiedExecutor	opt-in	allowed-opt-in	set CURATOR_TEST_YARN_MODERN_JS to the @yarnpkg/cli-dist 4.9.2 bin/yarn.js integration tool

```

## candidate-integrity-03.log

```text
signature G; remote head equals signed CI head; current worktree equals candidate tree; Story HEAD unchanged

```

## candidate-signature-03.txt

```text
9e9d63976b5adfa61879183e913de8a1d7d61da8 G

```

## candidate-03-tree.txt

```text
df0863ad2abdd830f76d579ca9471498a63b6112

```

PNPM whole-package retry exited 0. platform-gate-02 exited 1 for mandatory PNPM skips; final platform-gate-03 exited 0 with 36 permitted skips. Root identity stale-input control exited 1 for its intended assertion. Both remain recorded as failing controls, never passing validation commands.
