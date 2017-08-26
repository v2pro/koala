#define _GNU_SOURCE

#include <dlfcn.h>
#include <stddef.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/syscall.h>
#include <dirent.h>
#include "init.h"
#include "thread_id.h"
#include "_cgo_export.h"

extern long syscall(long number, ...);

#define HOOK_SYS_FUNC(name) if( !orig_##name##_func ) { orig_##name##_func = (name##_pfn_t)dlsym(RTLD_NEXT,#name); }

typedef FILE * (*fopen_pfn_t)(const char *filename, const char *opentype);
static fopen_pfn_t orig_fopen_func;

typedef FILE * (*fopen64_pfn_t)(const char *filename, const char *opentype);
static fopen64_pfn_t orig_fopen64_func;

typedef int (*open_pfn_t)(const char *filename, int flags, mode_t mode);
static open_pfn_t orig_open_func;

typedef int (*open64_pfn_t)(const char *filename, int flags, mode_t mode);
static open64_pfn_t orig_open64_func;

typedef ssize_t (*write_pfn_t)(int, const void *, size_t);
static write_pfn_t orig_write_func;

typedef int (*access_pfn_t)(const char *pathname, int mode);
static access_pfn_t orig_access_func;

FILE * fopen(const char *filename, const char *opentype) {
    HOOK_SYS_FUNC( fopen );
    if (is_go_initialized() != 1) {
        return orig_fopen_func(filename, opentype);
    }
    pid_t thread_id = get_thread_id();
    struct ch_span filename_span;
    filename_span.Ptr = filename;
    filename_span.Len = strlen(filename);
    struct ch_span opentype_span;
    opentype_span.Ptr = opentype;
    opentype_span.Len = strlen(opentype);
    struct ch_allocated_string redirect_to = on_fopening_file(thread_id, filename_span, opentype_span);
    if (redirect_to.Ptr != NULL) {
        FILE *file = orig_fopen_func(redirect_to.Ptr, opentype);
        if (file != NULL) {
            filename_span.Ptr = redirect_to.Ptr;
            filename_span.Len = strlen(redirect_to.Ptr);
            on_fopened_file(thread_id, fileno(file), filename_span, opentype_span);
        }
        free(redirect_to.Ptr);
        return file;
    }
    FILE *file = orig_fopen_func(filename, opentype);
    if (file != NULL) {
        on_fopened_file(thread_id, fileno(file), filename_span, opentype_span);
    }
    return file;
}

FILE * fopen64(const char *filename, const char *opentype) {
    HOOK_SYS_FUNC( fopen64 );
    if (is_go_initialized() != 1) {
        return orig_fopen64_func(filename, opentype);
    }
    pid_t thread_id = get_thread_id();
    struct ch_span filename_span;
    filename_span.Ptr = filename;
    filename_span.Len = strlen(filename);
    struct ch_span opentype_span;
    opentype_span.Ptr = opentype;
    opentype_span.Len = strlen(opentype);
    struct ch_allocated_string redirect_to = on_fopening_file(thread_id, filename_span, opentype_span);
    if (redirect_to.Ptr != NULL) {
        FILE *file = orig_fopen64_func(redirect_to.Ptr, opentype);
        if (file != NULL) {
            filename_span.Ptr = redirect_to.Ptr;
            filename_span.Len = strlen(redirect_to.Ptr);
            on_fopened_file(thread_id, fileno(file), filename_span, opentype_span);
        }
        free(redirect_to.Ptr);
        return file;
    }
    FILE *file = orig_fopen64_func(filename, opentype);
    if (file != NULL) {
        on_fopened_file(thread_id, fileno(file), filename_span, opentype_span);
    }
    return file;
}

int open(const char *filename, int flags, mode_t mode) {
    HOOK_SYS_FUNC( open );
    if (is_go_initialized() != 1) {
        return orig_open_func(filename, flags, mode);
    }
    pid_t thread_id = get_thread_id();
    struct ch_span filename_span;
    filename_span.Ptr = filename;
    filename_span.Len = strlen(filename);
    struct ch_allocated_string redirect_to = on_opening_file(thread_id, filename_span, flags, mode);
    if (redirect_to.Ptr != NULL) {
        int file = orig_open_func(redirect_to.Ptr, flags, mode);
        if (file != -1) {
            filename_span.Ptr = redirect_to.Ptr;
            filename_span.Len = strlen(redirect_to.Ptr);
            on_opened_file(thread_id, file, filename_span, flags, mode);
        }
        free(redirect_to.Ptr);
        return file;
    }
    int file = orig_open_func(filename, flags, mode);
    if (file != -1) {
        on_opened_file(thread_id, file, filename_span, flags, mode);
    }
    return file;
}

int open64(const char *filename, int flags, mode_t mode) {
    HOOK_SYS_FUNC( open64 );
    if (is_go_initialized() != 1) {
        return orig_open64_func(filename, flags, mode);
    }
    pid_t thread_id = get_thread_id();
    struct ch_span filename_span;
    filename_span.Ptr = filename;
    filename_span.Len = strlen(filename);
    struct ch_allocated_string redirect_to = on_opening_file(thread_id, filename_span, flags, mode);
    if (redirect_to.Ptr != NULL) {
        int file = orig_open64_func(redirect_to.Ptr, flags, mode);
        if (file != -1) {
            filename_span.Ptr = redirect_to.Ptr;
            filename_span.Len = strlen(redirect_to.Ptr);
            on_opened_file(thread_id, file, filename_span, flags, mode);
        }
        free(redirect_to.Ptr);
        return file;
    }
    int file = orig_open64_func(filename, flags, mode);
    if (file != -1) {
        on_opened_file(thread_id, file, filename_span, flags, mode);
    }
    return file;
}

ssize_t write(int fileFD, const void *buffer, size_t size) {
    HOOK_SYS_FUNC( write );
    if (is_go_initialized() != 1) {
        return orig_write_func(fileFD, buffer, size);
    }
    ssize_t written_size = orig_write_func(fileFD, buffer, size);
    if (written_size >= 0) {
        pid_t thread_id = get_thread_id();
        struct ch_span span;
        span.Ptr = buffer;
        span.Len = written_size;
        on_write(thread_id, fileFD, span);
    }
    return written_size;
}

int access(const char *pathname, int mode) {
    HOOK_SYS_FUNC( access );
    if (is_go_initialized() != 1) {
        return orig_access_func(pathname, mode);
    }
    pid_t thread_id = get_thread_id();
    struct ch_span pathname_span;
    pathname_span.Ptr = pathname;
    pathname_span.Len = strlen(pathname);
    struct ch_allocated_string redirect_to = on_access(thread_id, pathname_span, mode);
    if (redirect_to.Ptr != NULL) {
        int result = orig_access_func(redirect_to.Ptr, mode);
        free(redirect_to.Ptr);
        return result;
    }
    return orig_access_func(pathname, mode);
}