#ifdef KOALA_REPLAYER
#include <dlfcn.h>
#include <stddef.h>
#include <stdio.h>
#include <time.h>
#include <sys/timeb.h>
#include <sys/time.h>
#include "interpose.h"
#include "time_hook.h"

extern "C" {
    static int offset = 0;
    void set_time_offset(int val) {
        offset = val;
    }
}

#ifdef __APPLE__
INTERPOSE(gettimeofday)(struct timeval *tv, void *tz) {
#else
INTERPOSE(gettimeofday)(struct timeval *tv, struct timezone *tz) {
#endif
    auto result = real::gettimeofday(tv, tz);
    if (tv != NULL) {
        auto old_result = tv->tv_sec;
        tv->tv_sec = tv->tv_sec + offset;
    }
    return result;
}
#ifdef __APPLE__

#else
INTERPOSE(clock_gettime)(clockid_t clk_id, struct timespec *tp) {
    auto result = real::clock_gettime(clk_id, tp);
    if (tp != NULL) {
        tp->tv_sec = tp->tv_sec + offset;
    }
    return result;
}
INTERPOSE(time)(time_t *time_tptr) {
    time_t my_time;
    auto result = real::time(&my_time);
    auto old_result = result;
    result += offset;
    if (time_tptr != NULL) {
        *time_tptr = result;
    }
    return result;
}
INTERPOSE(ftime)(struct timeb *tb) {
    auto result = real::ftime(tb);
    if (tb != NULL) {
        tb->time = tb->time + offset;
    }
    return result;
}
#endif
#endif // KOALA_REPLAYER