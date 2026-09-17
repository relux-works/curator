import copy, io, contextlib, sys
from pathlib import Path
from unittest.mock import patch
sys.path.insert(0, str(Path(__file__).parent / "delivery" / "tools"))
import validate
original = validate.load_json
checks = [("environments-codex-seed.json", "provisioning_cases", "b-strips-servers-keeps-rest", "b-without-servers-no-warning"), ("environments-codex-seed.json", "posture_cases", "a-home-unstripped-under-b", "a-home-lists-ungoverned"), ("environments-store-boundary.json", "resolve_cases", "swapped-system-prompt-bytes-untrusted", None)]
for filename, family, target, donor in checks:
    path = validate.SUITE / "vectors" / filename
    changed = copy.deepcopy(original(path))
    row = next(x for x in changed[family] if x["name"] == target)
    if donor:
        replacement = copy.deepcopy(next(x for x in changed[family] if x["name"] == donor))
        row.clear(); row.update(replacement); row["name"] = target
    else:
        row["fragment_emitted"] = True
    def load(path_arg):
        return changed if Path(path_arg) == path else original(path_arg)
    err = io.StringIO()
    with patch.object(validate, "load_json", side_effect=load), contextlib.redirect_stderr(err):
        result = validate.main()
    message = err.getvalue().strip()
    assert result == 1 and ("codex-seed case" in message if donor else "store-boundary case" in message), (result, message)
    print(f"{filename}/{target}: validate.main exit={result}; {message}")
print("3/3 targeted invalid cases rejected through main; input-loader substitution only, no gate or registration mocked; no manager runtime claim")
