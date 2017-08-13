#include <dlfcn.h>
#include <stddef.h>
#include <stdio.h>
#include <time.h>
#include "time_hook.h"

void ftpl_init();

void time_hook_init() {
    ftpl_init();
}

void parse_ft_string(const char *user_faked_time); // defined in libfaketime.c

void set_time_offset(int offset) {
    char str[80];
    if (offset > 0) {
        sprintf(str, "+%d", offset);
    } else if (offset < 0) {
        sprintf(str, "%d", offset);
    } else {
        return;
    }
    parse_ft_string(str);
}