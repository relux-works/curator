from __future__ import annotations

from pathlib import Path

from csk.builds import go_v1


def _windows_startup_identity(
    tmp_path: Path,
    *,
    with_archive: bool = False,
) -> tuple[go_v1._StartupIdentity, go_v1._InterpreterIdentity, Path]:
    base_home = tmp_path / "base-python"
    base_home.mkdir()
    base_interpreter = base_home / "python.exe"
    base_interpreter.write_bytes(b"base interpreter")
    base_interpreter.chmod(0o755)
    runtime_image = base_home / "python314.dll"
    runtime_image.write_bytes(b"base Python runtime")
    runtime_image.chmod(0o755)
    if with_archive:
        (base_home / "python314.zip").write_bytes(
            b"fixed synthetic runtime archive"
        )
    (base_home / "Lib").mkdir()
    (base_home / "DLLs").mkdir()

    venv = tmp_path / "manager-venv"
    scripts = venv / "Scripts"
    scripts.mkdir(parents=True)
    venv_interpreter = scripts / "python.exe"
    venv_interpreter.write_bytes(b"venv launcher")
    venv_interpreter.chmod(0o755)
    (venv / "pyvenv.cfg").write_text(
        f"home = {base_home}\n"
        "include-system-site-packages = false\n"
        "version = 3.14.4\n",
        encoding="utf-8",
    )
    site_root = venv / "Lib" / "site-packages"
    site_root.mkdir(parents=True)

    executable = go_v1._resolve_executable_identity(venv_interpreter)
    runtime = go_v1._resolve_windows_interpreter_runtime(executable)
    interpreter = go_v1._InterpreterIdentity(
        invocation_path=venv_interpreter,
        links=(),
        executable=executable,
        runtime=runtime,
    )
    startup = go_v1._resolve_startup_identity(
        scripts / "csk.exe",
        interpreter,
        site_root,
    )
    return startup, interpreter, base_home


def test_windows_python_home_identity_survives_worker_protocol_round_trip(
    tmp_path: Path,
) -> None:
    startup, interpreter, _ = _windows_startup_identity(tmp_path)

    decoded = go_v1._startup_identity_from_mapping(
        startup.to_dict(),
        interpreter,
    )

    assert decoded == startup


def test_windows_fixed_runtime_archive_slot_is_identity_bound(
    tmp_path: Path,
) -> None:
    startup, _, base_home = _windows_startup_identity(
        tmp_path,
        with_archive=True,
    )

    assert [item.path for item in startup.archives] == [
        base_home / "python314.zip"
    ]
