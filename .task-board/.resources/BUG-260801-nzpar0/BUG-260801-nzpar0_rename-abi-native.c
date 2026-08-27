#define WIN32_LEAN_AND_MEAN
#define _WIN32_WINNT 0x0A00
#include <windows.h>

#include <stddef.h>
#include <stdio.h>

int wmain(void) {
    printf("pointer_bits=%zu\n", sizeof(void *) * 8);
    printf("sizeof_BOOLEAN=%zu\n", sizeof(BOOLEAN));
    printf("sizeof_DWORD=%zu\n", sizeof(DWORD));
    printf("sizeof_HANDLE=%zu\n", sizeof(HANDLE));
    printf("sizeof_WCHAR=%zu\n", sizeof(WCHAR));
    printf("sizeof_FILE_RENAME_INFO=%zu\n", sizeof(FILE_RENAME_INFO));
    printf("offset_RootDirectory=%zu\n", offsetof(FILE_RENAME_INFO, RootDirectory));
    printf("offset_FileNameLength=%zu\n", offsetof(FILE_RENAME_INFO, FileNameLength));
    printf("offset_FileName=%zu\n", offsetof(FILE_RENAME_INFO, FileName));
    return 0;
}
