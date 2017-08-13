#include <dlfcn.h>
#include <stddef.h>
#include <stdio.h>
#include <time.h>
#include "time_hook.h"

//#define RTLD_NEXT	((void *) -1l)
//
//#define HOOK_SYS_FUNC(name) if( !orig_##name##_func ) { orig_##name##_func = (name##_pfn_t)dlsym(RTLD_NEXT,#name); }
//
//typedef int (*clock_gettime_pfn_t)(clockid_t clk_id, struct timespec *tp);
//static clock_gettime_pfn_t orig_clock_gettime_func;

void ftpl_init();

void time_hook_init() {
    ftpl_init();
//    HOOK_SYS_FUNC( clock_gettime );
}

//int clock_gettime(clockid_t clk_id, struct timespec *tp) {
//    printf("hello!\n");
//    fflush(stdout);
//	tp->tv_sec = tp->tv_sec + (time_t)(-200);
//    return orig_clock_gettime_func(clk_id, tp);
//}

void set_time_offset(int offset) {
}