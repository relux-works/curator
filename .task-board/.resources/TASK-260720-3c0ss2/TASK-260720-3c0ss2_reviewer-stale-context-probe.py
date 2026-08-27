from __future__ import annotations

import json
import sys
import tempfile
from pathlib import Path


WORKTREE = Path(
    "/Users/iv/Developer/intranet/cocoaskills/.temp/"
    "TASK-260720-3c0ss2/worktree"
)
sys.path.insert(0, str(WORKTREE / "tests"))
sys.path.insert(0, str(WORKTREE / "src"))

from conftest import make_config, make_project, make_skill_repo, write_skillfile
from csk import hashing, installer


with tempfile.TemporaryDirectory(prefix="csk-review-stale-context-") as raw_tmp:
    tmp_path = Path(raw_tmp)
    skills_root = tmp_path / "skills"
    skills_root.mkdir()
    csk_home = tmp_path / ".cocoaskills"
    csk_home.mkdir()
    project = make_project(tmp_path)
    make_skill_repo(
        skills_root,
        "skill-build",
        {
            "agent-skill.json": json.dumps(
                {
                    "schema_version": 6,
                    "build_roots": ["assets/build-tool"],
                    "commands": {
                        "build-tool": {
                            "type": "build",
                            "driver": "go-v1",
                            "source_dir": "assets/build-tool/cmd/tool",
                        }
                    },
                    "capabilities": {},
                }
            ),
            "assets/prompt.md": "visible prompt asset\n",
            "assets/build-tool/go.mod": "module example.com/tool\n\ngo 1.23\n",
            "assets/build-tool/cmd/tool/main.go": "package main\n",
        },
        tag="v1",
    )
    write_skillfile(
        project,
        {
            "schema_version": 1,
            "skills": [{"name": "skill-build", "tag": "v1"}],
        },
    )
    cfg = make_config(csk_home, skills_root, project)

    first = installer.install(cfg)[0]
    assert not first.errors, first.errors
    installed = project / ".agents" / "skills" / "skill-build"
    stale_root = installed / "assets" / "build-tool"
    (stale_root / "cmd" / "tool").mkdir(parents=True)
    (stale_root / "go.mod").write_text(
        "module example.com/tool\n\ngo 1.23\n",
        encoding="utf-8",
    )
    (stale_root / "cmd" / "tool" / "main.go").write_text(
        "package main\n",
        encoding="utf-8",
    )

    marker_path = installed / ".csk-install.json"
    marker = json.loads(marker_path.read_text(encoding="utf-8"))
    marker["files"] = sorted(
        [
            *marker["files"],
            "assets/build-tool/go.mod",
            "assets/build-tool/cmd/tool/main.go",
        ]
    )
    marker["content_sha256"] = hashing.content_sha256(installed)
    marker_path.write_text(
        json.dumps(marker, indent=2, sort_keys=True) + "\n",
        encoding="utf-8",
    )

    second = installer.install(cfg)[0]
    print(f"second_errors={second.errors!r}")
    print(f"second_messages={second.messages!r}")
    print(f"stale_build_root_exists={stale_root.exists()}")
    print(
        "stale_build_file_in_marker="
        f"{'assets/build-tool/go.mod' in marker['files']}"
    )
    assert not second.errors
    assert any("up-to-date" in message for message in second.messages)
    assert stale_root.exists()
