#ifndef __NETWORK_HOOK_H__
#define __NETWORK_HOOK_H__

void libc_hook_init();
int socket(int, int, int);
ssize_t send(int, const void *, size_t, int);
ssize_t sendto(int, const void *, size_t, int, const struct sockaddr *, socklen_t);
int connect(int, const struct sockaddr *, socklen_t);
int accept(int, struct sockaddr *, socklen_t *);

#endif