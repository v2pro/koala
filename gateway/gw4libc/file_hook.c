#include <dlfcn.h>
#include <stddef.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/syscall.h>
#include "_cgo_export.h"

extern long syscall(long number, ...);

#define RTLD_NEXT	((void *) -1l)

#define HOOK_SYS_FUNC(name) if( !orig_##name##_func ) { orig_##name##_func = (name##_pfn_t)dlsym(RTLD_NEXT,#name); }

typedef FILE * (*fopen64_pfn_t)(const char *filename, const char *opentype);
static fopen64_pfn_t orig_fopen64_func;

void file_hook_init (void) __attribute__ ((constructor));
void file_hook_init() {
    HOOK_SYS_FUNC( fopen64 );
}

FILE * fopen64(const char *filename, const char *opentype) {
    pid_t thread_id = syscall(__NR_gettid);
    struct ch_span filename_span;
    filename_span.Ptr = filename;
    filename_span.Len = strlen(filename);
    struct ch_span opentype_span;
    opentype_span.Ptr = opentype;
    opentype_span.Len = strlen(opentype);
    struct ch_allocated_string redirectToFilename = on_opening_file(thread_id, filename_span, opentype_span);
    if (redirectToFilename.Ptr != NULL) {
        FILE *file = orig_fopen64_func(redirectToFilename.Ptr, opentype);
        if (file != NULL) {
            filename_span.Ptr = redirectToFilename.Ptr;
            filename_span.Len = strlen(redirectToFilename.Ptr);
            on_opened_file(thread_id, file, filename_span, opentype_span);
        }
        free(redirectToFilename.Ptr);
        return file;
    }
    FILE *file = orig_fopen64_func(filename, opentype);
    if (file != NULL) {
        on_opened_file(thread_id, file, filename_span, opentype_span);
    }
    return file;
}