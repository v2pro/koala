#include <dlfcn.h>
#include <stddef.h>
#include <stdio.h>
#include <string.h>
#include <netdb.h>
#include <math.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <netinet/ip.h>
#include <arpa/inet.h>
#define _GNU_SOURCE
#include <unistd.h>
#include <sys/types.h>
#include <sys/syscall.h>
#include "network_hook.h"
#include "_cgo_export.h"

#define RTLD_NEXT	((void *) -1l)

#define HOOK_SYS_FUNC(name) if( !orig_##name##_func ) { orig_##name##_func = (name##_pfn_t)dlsym(RTLD_NEXT,#name); }

typedef int (*socket_pfn_t)(int, int, int);
static socket_pfn_t orig_socket_func;

typedef ssize_t (*send_pfn_t)(int, const void *, size_t, int);
static send_pfn_t orig_send_func;

typedef ssize_t (*sendto_pfn_t)(int, const void *, size_t, int, const struct sockaddr *, socklen_t);
static sendto_pfn_t orig_sendto_func;

typedef int (*connect_pfn_t)(int, const struct sockaddr *, socklen_t);
static connect_pfn_t orig_connect_func;

typedef int (*accept_pfn_t)(int, struct sockaddr *, socklen_t *);
static accept_pfn_t orig_accept_func;

void libc_hook_init() {
    HOOK_SYS_FUNC( socket );
    HOOK_SYS_FUNC( send );
    HOOK_SYS_FUNC( sendto );
    HOOK_SYS_FUNC( connect );
    HOOK_SYS_FUNC( accept );
}

int socket(int domain, int type, int protocol) {
//    pid_t thread_id = syscall(__NR_gettid);
//    printf("create socket from %d\n", thread_id);
    return orig_socket_func(domain, type, protocol);
}

ssize_t send(int sockfd, const void *buf, size_t len, int flags) {
//    fwrite(buf, len, 1, stdout);
//    printf("\n");
    on_send(sockfd, (void *)(buf), len);
    return orig_send_func(sockfd, buf, len, flags);
}

ssize_t sendto(int sockfd, const void *buf, size_t len, int flags,
               const struct sockaddr *dest_addr, socklen_t addrlen) {
    on_send(sockfd, (void *)(buf), len);
    return orig_sendto_func(sockfd, buf, len, flags, dest_addr, addrlen);
}

int connect(int sockfd, const struct sockaddr *addr, socklen_t addrlen) {
    if (addr->sa_family == AF_INET) {
        struct sockaddr_in *sin = (struct sockaddr_in *)(addr);
        on_connect(sockfd, inet_ntoa(sin->sin_addr), ntohs(sin->sin_port));
    }
    return orig_connect_func(sockfd, addr, addrlen);
}

int accept(int serverSockFd, struct sockaddr *addr, socklen_t *addrlen) {
    int clientSockFd = orig_accept_func(serverSockFd, addr, addrlen);
    if (addr->sa_family == AF_INET) {
        struct sockaddr_in *sin = (struct sockaddr_in *)(addr);
        on_accept(serverSockFd, clientSockFd, sin);
    }
    return clientSockFd;
}
