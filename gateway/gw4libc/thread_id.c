#define _GNU_SOURCE

#include <unistd.h>
#include <sys/syscall.h>
#include "thread_id.h"

pid_t get_thread_id() {
    return syscall(__NR_gettid);
}