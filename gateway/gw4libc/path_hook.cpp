#ifdef KOALA_REPLAYER
#ifndef _GNU_SOURCE
#define _GNU_SOURCE
#endif

#include <dlfcn.h>
#include <stddef.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/syscall.h>
#include <dirent.h>
#include <sys/stat.h>
#include <unistd.h>
#include "interpose.h"
#include "init.h"
#include "thread_id.h"
#include "_cgo_export.h"

#define PATH_HOOK_ENTER \
    char *redirect_to = try_redirect_path(path); \
    if (redirect_to != NULL) { \
        path = redirect_to; \
    }
#define PATH_HOOK_EXIT \
    if (redirect_to != NULL) { \
        free(redirect_to); \
    } \
    return result;

char *try_redirect_path(const char *path) {
    if (is_go_initialized() != 1) {
        return NULL;
    }
    pid_t thread_id = get_thread_id();
    struct ch_span path_span;
    path_span.Ptr = path;
    path_span.Len = strlen(path);
    struct ch_allocated_string redirect_to = redirect_path(thread_id, path_span);
    return redirect_to.Ptr;
}

INTERPOSE(access)(const char *path, int mode) {
    PATH_HOOK_ENTER
    auto result = real::access(path, mode);
    PATH_HOOK_EXIT
}

INTERPOSE(__xstat)(int ver, const char *path, struct stat *buf) {
    PATH_HOOK_ENTER
    auto result = real::__xstat(ver, path, buf);
    PATH_HOOK_EXIT
}

#endif // KOALA_REPLAYER