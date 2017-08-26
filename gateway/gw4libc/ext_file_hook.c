#ifdef KOALA_REPLAYER
#define _GNU_SOURCE

#include <dlfcn.h>
#include <stddef.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/syscall.h>
#include <dirent.h>
#include <sys/stat.h>
#include "init.h"
#include "thread_id.h"
#include "_cgo_export.h"

extern long syscall(long number, ...);

#define HOOK_SYS_FUNC(name) if( !orig_##name##_func ) { orig_##name##_func = (name##_pfn_t)dlsym(RTLD_NEXT,#name); }

typedef int (*access_pfn_t)(const char *pathname, int mode);
static access_pfn_t orig_access_func;

typedef int (*__xstat_pfn_t) (int ver, const char *path, struct stat *buf);
static __xstat_pfn_t orig___xstat_func;

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

int __xstat (int ver, const char *pathname, struct stat *buf) {
    HOOK_SYS_FUNC( __xstat );
    if (is_go_initialized() != 1) {
        return orig___xstat_func(ver, pathname, buf);
    }
    pid_t thread_id = get_thread_id();
    struct ch_span pathname_span;
    pathname_span.Ptr = pathname;
    pathname_span.Len = strlen(pathname);
    struct ch_allocated_string redirect_to = on_xstat(thread_id, pathname_span);
    if (redirect_to.Ptr != NULL) {
        int result = orig___xstat_func(ver, redirect_to.Ptr, buf);
        free(redirect_to.Ptr);
        return result;
    }
    return orig___xstat_func(ver, pathname, buf);
}

#endif // KOALA_REPLAYER