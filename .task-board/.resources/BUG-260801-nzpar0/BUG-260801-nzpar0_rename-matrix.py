"""Task-scoped Windows rename ABI/handle probe for BUG-260801-nzpar0.

This is diagnostic code only. It creates and removes one process-scoped tree
below the current user's temporary directory and does not import CocoaSkills.
"""

from __future__ import annotations

import ctypes
import json
import os
import platform
import shutil
import struct
import sys
import tempfile
from pathlib import Path
from typing import Any


if os.name != "nt":
    raise SystemExit("rename_matrix.py must run on native Windows")


DELETE = 0x00010000
READ_CONTROL = 0x00020000
SYNCHRONIZE = 0x00100000
FILE_READ_DATA = 0x0001
FILE_WRITE_DATA = 0x0002
FILE_EXECUTE = 0x0020
FILE_READ_ATTRIBUTES = 0x0080
GENERIC_READ = 0x80000000
FILE_SHARE_READ = 0x00000001
FILE_SHARE_WRITE = 0x00000002
FILE_SHARE_DELETE = 0x00000004
SHARE_ALL = FILE_SHARE_READ | FILE_SHARE_WRITE | FILE_SHARE_DELETE
OPEN_EXISTING = 3
FILE_ATTRIBUTE_NORMAL = 0x00000080
FILE_FLAG_OPEN_REPARSE_POINT = 0x00200000
FILE_FLAG_BACKUP_SEMANTICS = 0x02000000
FILE_TYPE_DISK = 0x0001
FILE_STANDARD_INFO_CLASS = 1
FILE_RENAME_INFO_CLASS = 3
FILE_ID_INFO_CLASS = 18
FILE_RENAME_INFORMATION_CLASS = 10
OBJECT_BASIC_INFORMATION_CLASS = 0
ERROR_SHARING_VIOLATION = 32
INVALID_HANDLE_VALUE = ctypes.c_void_p(-1).value


class FileRenameDword(ctypes.Structure):
    _fields_ = [
        ("replace_or_flags", ctypes.c_uint32),
        ("root_directory", ctypes.c_void_p),
        ("file_name_length", ctypes.c_uint32),
        ("file_name", ctypes.c_wchar * 1),
    ]


class FileRenameBoolean(ctypes.Structure):
    _fields_ = [
        ("replace_if_exists", ctypes.c_ubyte),
        ("root_directory", ctypes.c_void_p),
        ("file_name_length", ctypes.c_uint32),
        ("file_name", ctypes.c_wchar * 1),
    ]


class FileRenamePacked(ctypes.Structure):
    _pack_ = 1
    _fields_ = [
        ("replace_or_flags", ctypes.c_uint32),
        ("root_directory", ctypes.c_void_p),
        ("file_name_length", ctypes.c_uint32),
        ("file_name", ctypes.c_wchar * 1),
    ]


class FileStandardInfo(ctypes.Structure):
    _fields_ = [
        ("allocation_size", ctypes.c_longlong),
        ("end_of_file", ctypes.c_longlong),
        ("number_of_links", ctypes.c_uint32),
        ("delete_pending", ctypes.c_ubyte),
        ("directory", ctypes.c_ubyte),
    ]


class FileId128(ctypes.Structure):
    _fields_ = [("identifier", ctypes.c_ubyte * 16)]


class FileIdInfo(ctypes.Structure):
    _fields_ = [
        ("volume_serial_number", ctypes.c_ulonglong),
        ("file_id", FileId128),
    ]


class IoStatusBlock(ctypes.Structure):
    _fields_ = [
        ("status_or_pointer", ctypes.c_ssize_t),
        ("information", ctypes.c_size_t),
    ]


class ObjectBasicInformation(ctypes.Structure):
    _fields_ = [
        ("attributes", ctypes.c_uint32),
        ("granted_access", ctypes.c_uint32),
        ("handle_count", ctypes.c_uint32),
        ("pointer_count", ctypes.c_uint32),
        ("paged_pool_charge", ctypes.c_uint32),
        ("non_paged_pool_charge", ctypes.c_uint32),
        ("reserved", ctypes.c_uint32 * 3),
        ("name_info_size", ctypes.c_uint32),
        ("type_info_size", ctypes.c_uint32),
        ("security_descriptor_size", ctypes.c_uint32),
        ("creation_time", ctypes.c_longlong),
    ]


kernel32 = ctypes.WinDLL("kernel32", use_last_error=True)
ntdll = ctypes.WinDLL("ntdll", use_last_error=True)

kernel32.CreateFileW.argtypes = [
    ctypes.c_wchar_p,
    ctypes.c_uint32,
    ctypes.c_uint32,
    ctypes.c_void_p,
    ctypes.c_uint32,
    ctypes.c_uint32,
    ctypes.c_void_p,
]
kernel32.CreateFileW.restype = ctypes.c_void_p
kernel32.CloseHandle.argtypes = [ctypes.c_void_p]
kernel32.CloseHandle.restype = ctypes.c_int
kernel32.GetFileType.argtypes = [ctypes.c_void_p]
kernel32.GetFileType.restype = ctypes.c_uint32
kernel32.GetFileInformationByHandleEx.argtypes = [
    ctypes.c_void_p,
    ctypes.c_int,
    ctypes.c_void_p,
    ctypes.c_uint32,
]
kernel32.GetFileInformationByHandleEx.restype = ctypes.c_int
kernel32.GetFinalPathNameByHandleW.argtypes = [
    ctypes.c_void_p,
    ctypes.c_wchar_p,
    ctypes.c_uint32,
    ctypes.c_uint32,
]
kernel32.GetFinalPathNameByHandleW.restype = ctypes.c_uint32
kernel32.SetFileInformationByHandle.argtypes = [
    ctypes.c_void_p,
    ctypes.c_int,
    ctypes.c_void_p,
    ctypes.c_uint32,
]
kernel32.SetFileInformationByHandle.restype = ctypes.c_int
ntdll.NtSetInformationFile.argtypes = [
    ctypes.c_void_p,
    ctypes.POINTER(IoStatusBlock),
    ctypes.c_void_p,
    ctypes.c_uint32,
    ctypes.c_int,
]
ntdll.NtSetInformationFile.restype = ctypes.c_long
ntdll.RtlNtStatusToDosError.argtypes = [ctypes.c_long]
ntdll.RtlNtStatusToDosError.restype = ctypes.c_uint32
ntdll.NtQueryObject.argtypes = [
    ctypes.c_void_p,
    ctypes.c_int,
    ctypes.c_void_p,
    ctypes.c_uint32,
    ctypes.POINTER(ctypes.c_uint32),
]
ntdll.NtQueryObject.restype = ctypes.c_long


def emit(event: str, **values: Any) -> None:
    print(json.dumps({"event": event, **values}, sort_keys=True), flush=True)


def extended(path: Path) -> str:
    value = os.fspath(path.resolve())
    if value.startswith("\\\\?\\"):
        return value
    if value.startswith("\\\\"):
        return "\\\\?\\UNC\\" + value[2:]
    return "\\\\?\\" + value


def format_error(error: int) -> str:
    try:
        return ctypes.FormatError(error).strip()
    except OSError:
        return ""


def open_handle(
    path: Path,
    desired_access: int,
    *,
    share_mode: int = SHARE_ALL,
) -> int:
    ctypes.set_last_error(0)
    raw = kernel32.CreateFileW(
        extended(path),
        desired_access,
        share_mode,
        None,
        OPEN_EXISTING,
        FILE_ATTRIBUTE_NORMAL
        | FILE_FLAG_OPEN_REPARSE_POINT
        | FILE_FLAG_BACKUP_SEMANTICS,
        None,
    )
    if raw in {None, INVALID_HANDLE_VALUE}:
        error = ctypes.get_last_error()
        raise OSError(error, format_error(error), os.fspath(path))
    return int(raw)


def close_handle(handle: int | None) -> None:
    if handle is not None:
        if not kernel32.CloseHandle(handle):
            error = ctypes.get_last_error()
            raise OSError(error, format_error(error))


def final_path(handle: int) -> str:
    required = int(kernel32.GetFinalPathNameByHandleW(handle, None, 0, 0))
    if not required:
        error = ctypes.get_last_error()
        raise OSError(error, format_error(error))
    buffer = ctypes.create_unicode_buffer(required + 1)
    copied = int(
        kernel32.GetFinalPathNameByHandleW(handle, buffer, len(buffer), 0)
    )
    if not copied or copied >= len(buffer):
        error = ctypes.get_last_error()
        raise OSError(error, format_error(error))
    return buffer.value


def handle_details(handle: int, requested_access: int) -> dict[str, Any]:
    standard = FileStandardInfo()
    if not kernel32.GetFileInformationByHandleEx(
        handle,
        FILE_STANDARD_INFO_CLASS,
        ctypes.byref(standard),
        ctypes.sizeof(standard),
    ):
        error = ctypes.get_last_error()
        raise OSError(error, format_error(error))
    identity = FileIdInfo()
    if not kernel32.GetFileInformationByHandleEx(
        handle,
        FILE_ID_INFO_CLASS,
        ctypes.byref(identity),
        ctypes.sizeof(identity),
    ):
        error = ctypes.get_last_error()
        raise OSError(error, format_error(error))
    basic = ObjectBasicInformation()
    returned = ctypes.c_uint32()
    status = int(
        ntdll.NtQueryObject(
            handle,
            OBJECT_BASIC_INFORMATION_CLASS,
            ctypes.byref(basic),
            ctypes.sizeof(basic),
            ctypes.byref(returned),
        )
    )
    status_u32 = status & 0xFFFFFFFF
    return {
        "requested_access": f"0x{requested_access:08x}",
        "granted_access": (
            f"0x{int(basic.granted_access):08x}" if status_u32 == 0 else None
        ),
        "nt_query_object_status": f"0x{status_u32:08x}",
        "file_type": int(kernel32.GetFileType(handle)),
        "is_disk": int(kernel32.GetFileType(handle)) == FILE_TYPE_DISK,
        "is_directory": bool(standard.directory),
        "links": int(standard.number_of_links),
        "volume_serial": f"0x{int(identity.volume_serial_number):016x}",
        "file_id": bytes(identity.file_id.identifier).hex(),
        "final_path": final_path(handle),
    }


def layout(structure: type[ctypes.Structure]) -> dict[str, Any]:
    return {
        "type": structure.__name__,
        "sizeof": ctypes.sizeof(structure),
        "alignment": ctypes.alignment(structure),
        "offset_root": structure.root_directory.offset,
        "offset_length": structure.file_name_length.offset,
        "offset_name": structure.file_name.offset,
    }


def make_buffer(
    structure: type[ctypes.Structure],
    name: str,
    root_handle: int | None,
    *,
    length_mode: str = "bytes",
    size_mode: str = "sizeof_plus",
) -> tuple[ctypes.Array[ctypes.c_char], dict[str, Any]]:
    raw_name = name.encode("utf-16-le")
    name_offset = structure.file_name.offset
    if size_mode == "sizeof_plus":
        buffer_size = ctypes.sizeof(structure) + len(raw_name)
    elif size_mode == "offset_plus":
        buffer_size = name_offset + len(raw_name)
    elif size_mode == "sizeof_plus_nul":
        buffer_size = ctypes.sizeof(structure) + len(raw_name) + 2
    else:
        raise ValueError(size_mode)
    buffer = ctypes.create_string_buffer(buffer_size)
    info = structure.from_buffer(buffer)
    first_field = structure._fields_[0][0]
    setattr(info, first_field, 0)
    info.root_directory = root_handle
    if length_mode == "bytes":
        info.file_name_length = len(raw_name)
    elif length_mode == "characters":
        info.file_name_length = len(name)
    elif length_mode == "bytes_plus_nul":
        info.file_name_length = len(raw_name) + 2
    else:
        raise ValueError(length_mode)
    ctypes.memmove(
        ctypes.addressof(buffer) + name_offset,
        raw_name,
        len(raw_name),
    )
    evidence = {
        **layout(structure),
        "name": name,
        "name_utf16le_hex": raw_name.hex(),
        "name_bytes": len(raw_name),
        "declared_name_length": int(info.file_name_length),
        "buffer_size": len(buffer),
        "size_mode": size_mode,
        "length_mode": length_mode,
        "header_hex": bytes(buffer[: min(len(buffer), 32)]).hex(),
    }
    return buffer, evidence


def create_source(path: Path, kind: str) -> None:
    if kind == "directory":
        path.mkdir()
        (path / "closed-child.txt").write_text("closed\n", encoding="utf-8")
    elif kind == "file":
        path.write_text("closed\n", encoding="utf-8")
    else:
        raise ValueError(kind)


def run_case(
    base: Path,
    label: str,
    *,
    api: str,
    relative_root: bool,
    root_access: int = FILE_READ_ATTRIBUTES | FILE_EXECUTE,
    structure: type[ctypes.Structure] = FileRenameDword,
    length_mode: str = "bytes",
    size_mode: str = "sizeof_plus",
    source_access: int = READ_CONTROL | FILE_READ_ATTRIBUTES | DELETE,
    source_share: int = SHARE_ALL,
    source_kind: str = "directory",
    collision: bool = False,
    root_is_file: bool = False,
    absolute_extended: bool = True,
) -> None:
    case_root = base / label
    source_parent = case_root / "source-parent"
    destination_parent = case_root / "destination-parent"
    source_parent.mkdir(parents=True)
    destination_parent.mkdir()
    source = source_parent / "source"
    destination_name = "destination"
    destination = destination_parent / destination_name
    create_source(source, source_kind)
    if collision:
        create_source(destination, source_kind)
    root_path = destination_parent
    if root_is_file:
        root_path = case_root / "not-a-directory.txt"
        root_path.write_text("not a directory\n", encoding="utf-8")

    source_handle: int | None = None
    root_handle: int | None = None
    try:
        source_handle = open_handle(
            source,
            source_access,
            share_mode=source_share,
        )
        if relative_root:
            root_handle = open_handle(root_path, root_access)
        before = handle_details(source_handle, source_access)
        root = (
            handle_details(root_handle, root_access)
            if root_handle is not None
            else None
        )
        if relative_root:
            name = destination_name
        elif absolute_extended:
            name = extended(destination)
        else:
            name = os.fspath(destination.resolve())
        buffer, buffer_evidence = make_buffer(
            structure,
            name,
            root_handle,
            length_mode=length_mode,
            size_mode=size_mode,
        )
        emit(
            "case_input",
            label=label,
            api=api,
            source_kind=source_kind,
            source_handle=before,
            root_handle=root,
            relative_root=relative_root,
            collision=collision,
            buffer=buffer_evidence,
        )
        if api == "SetFileInformationByHandle":
            ctypes.set_last_error(0)
            success = bool(
                kernel32.SetFileInformationByHandle(
                    source_handle,
                    FILE_RENAME_INFO_CLASS,
                    ctypes.byref(buffer),
                    len(buffer),
                )
            )
            error = 0 if success else int(ctypes.get_last_error())
            status_hex = None
            iosb_status_hex = None
            iosb_information = None
        elif api == "NtSetInformationFile":
            iosb = IoStatusBlock()
            status = int(
                ntdll.NtSetInformationFile(
                    source_handle,
                    ctypes.byref(iosb),
                    ctypes.byref(buffer),
                    len(buffer),
                    FILE_RENAME_INFORMATION_CLASS,
                )
            )
            status_u32 = status & 0xFFFFFFFF
            success = status_u32 == 0
            error = (
                0
                if success
                else int(ntdll.RtlNtStatusToDosError(ctypes.c_long(status)))
            )
            status_hex = f"0x{status_u32:08x}"
            iosb_status_hex = f"0x{int(iosb.status_or_pointer) & 0xFFFFFFFF:08x}"
            iosb_information = int(iosb.information)
        else:
            raise ValueError(api)
        after = handle_details(source_handle, source_access)
        emit(
            "case_result",
            label=label,
            success=success,
            winerror=error,
            winerror_text=format_error(error) if error else "",
            ntstatus=status_hex,
            iosb_status=iosb_status_hex,
            iosb_information=iosb_information,
            source_path_exists=source.exists(),
            destination_path_exists=destination.exists(),
            identity_preserved=(
                before["volume_serial"] == after["volume_serial"]
                and before["file_id"] == after["file_id"]
            ),
            source_handle_after=after,
        )
    except OSError as exc:
        emit(
            "case_open_error",
            label=label,
            winerror=getattr(exc, "winerror", None) or exc.errno,
            message=str(exc),
        )
    finally:
        close_handle(root_handle)
        close_handle(source_handle)


def share_conflict_probe(base: Path) -> None:
    case_root = base / "share-conflict"
    case_root.mkdir()
    source = case_root / "source"
    source.mkdir()
    blocker: int | None = None
    try:
        blocker = open_handle(
            source,
            FILE_READ_ATTRIBUTES,
            share_mode=FILE_SHARE_READ | FILE_SHARE_WRITE,
        )
        emit(
            "share_blocker",
            blocker=handle_details(blocker, FILE_READ_ATTRIBUTES),
            blocker_share=f"0x{FILE_SHARE_READ | FILE_SHARE_WRITE:08x}",
        )
        try:
            rename_handle = open_handle(
                source,
                DELETE | FILE_READ_ATTRIBUTES,
                share_mode=SHARE_ALL,
            )
        except OSError as exc:
            error = getattr(exc, "winerror", None) or exc.errno
            emit(
                "share_conflict_result",
                success=False,
                winerror=error,
                expected_winerror=ERROR_SHARING_VIOLATION,
                expected=(error == ERROR_SHARING_VIOLATION),
                message=str(exc),
            )
        else:
            close_handle(rename_handle)
            emit("share_conflict_result", success=True, expected=False)
    finally:
        close_handle(blocker)


def main() -> int:
    windows_version = sys.getwindowsversion()
    emit(
        "environment",
        python=sys.version,
        executable=sys.executable,
        platform=platform.platform(),
        windows_version={
            "major": windows_version.major,
            "minor": windows_version.minor,
            "build": windows_version.build,
            "platform": windows_version.platform,
            "service_pack": windows_version.service_pack,
        },
        machine=platform.machine(),
        pointer_bits=struct.calcsize("P") * 8,
        wchar_size=ctypes.sizeof(ctypes.c_wchar),
    )
    for structure in (FileRenameDword, FileRenameBoolean, FileRenamePacked):
        emit("ctypes_layout", **layout(structure))

    base = Path(tempfile.gettempdir()) / f"BUG-260801-nzpar0-{os.getpid()}"
    if base.exists():
        raise RuntimeError(f"refusing to reuse existing probe root: {base}")
    base.mkdir()
    emit("probe_root", path=os.fspath(base))
    try:
        run_case(
            base,
            "01-win32-relative-current",
            api="SetFileInformationByHandle",
            relative_root=True,
        )
        run_case(
            base,
            "02-win32-relative-boolean-layout",
            api="SetFileInformationByHandle",
            relative_root=True,
            structure=FileRenameBoolean,
        )
        run_case(
            base,
            "03-win32-relative-offset-buffer",
            api="SetFileInformationByHandle",
            relative_root=True,
            size_mode="offset_plus",
        )
        run_case(
            base,
            "04-win32-relative-character-length",
            api="SetFileInformationByHandle",
            relative_root=True,
            length_mode="characters",
        )
        run_case(
            base,
            "05-win32-relative-packed-wrong",
            api="SetFileInformationByHandle",
            relative_root=True,
            structure=FileRenamePacked,
        )
        run_case(
            base,
            "05a-win32-relative-length-includes-nul",
            api="SetFileInformationByHandle",
            relative_root=True,
            length_mode="bytes_plus_nul",
            size_mode="sizeof_plus_nul",
        )
        run_case(
            base,
            "06-win32-absolute-extended",
            api="SetFileInformationByHandle",
            relative_root=False,
        )
        run_case(
            base,
            "07-win32-absolute-dos",
            api="SetFileInformationByHandle",
            relative_root=False,
            absolute_extended=False,
        )
        run_case(
            base,
            "08-native-relative-current",
            api="NtSetInformationFile",
            relative_root=True,
        )
        run_case(
            base,
            "09-native-relative-file",
            api="NtSetInformationFile",
            relative_root=True,
            source_kind="file",
        )
        for index, (name, access) in enumerate(
            (
                ("read-attributes", FILE_READ_ATTRIBUTES),
                ("traverse", FILE_EXECUTE),
                ("traverse-read-attributes", FILE_EXECUTE | FILE_READ_ATTRIBUTES),
                ("generic-read", GENERIC_READ),
            ),
            start=10,
        ):
            run_case(
                base,
                f"{index:02d}-native-root-{name}",
                api="NtSetInformationFile",
                relative_root=True,
                root_access=access,
            )
        run_case(
            base,
            "14-native-source-without-delete",
            api="NtSetInformationFile",
            relative_root=True,
            source_access=READ_CONTROL | FILE_READ_ATTRIBUTES,
        )
        run_case(
            base,
            "15-native-no-replace-collision",
            api="NtSetInformationFile",
            relative_root=True,
            collision=True,
        )
        run_case(
            base,
            "16-native-root-is-file",
            api="NtSetInformationFile",
            relative_root=True,
            root_is_file=True,
        )
        share_conflict_probe(base)
    finally:
        shutil.rmtree(base)
        emit("cleanup", path=os.fspath(base), exists=base.exists())
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
