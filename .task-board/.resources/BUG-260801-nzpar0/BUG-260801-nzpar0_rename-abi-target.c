/*
 * ABI-only reproduction for BUG-260801-nzpar0.
 *
 * These typedefs are the Windows SDK widths used by FILE_RENAME_INFO.  The
 * file intentionally has no platform headers, so Clang can lay it out for
 * MSVC x86 and x64 targets from a non-Windows diagnostic host.
 */
typedef unsigned char BOOLEAN;
typedef unsigned long DWORD;
typedef void *HANDLE;
typedef unsigned short WCHAR;

typedef struct _FILE_RENAME_INFO_DWORD {
    DWORD ReplaceOrFlags;
    HANDLE RootDirectory;
    DWORD FileNameLength;
    WCHAR FileName[1];
} FILE_RENAME_INFO_DWORD;

typedef struct _FILE_RENAME_INFO_BOOLEAN {
    BOOLEAN ReplaceIfExists;
    HANDLE RootDirectory;
    DWORD FileNameLength;
    WCHAR FileName[1];
} FILE_RENAME_INFO_BOOLEAN;

unsigned long abi_sizes(void) {
    return (unsigned long)(
        sizeof(FILE_RENAME_INFO_DWORD) +
        sizeof(FILE_RENAME_INFO_BOOLEAN)
    );
}
