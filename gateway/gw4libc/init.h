#ifndef __INIT_H__
#define __INIT_H__

#ifdef __cplusplus
extern "C" {
#endif

void go_initialized(int is_tracing);
int is_go_initialized();
int is_tracing();

#ifdef __cplusplus
}
#endif

#endif