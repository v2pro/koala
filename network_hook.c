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
#include "span.h"
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

typedef ssize_t (*recv_pfn_t)(int socket, void *buffer, size_t size, int flags);
static recv_pfn_t orig_recv_func;

typedef int (*bind_pfn_t)(int socket, const struct sockaddr *addr, socklen_t length);
static bind_pfn_t orig_bind_func;

void libc_hook_init() {
    HOOK_SYS_FUNC( socket );
    HOOK_SYS_FUNC( send );
    HOOK_SYS_FUNC( sendto );
    HOOK_SYS_FUNC( connect );
    HOOK_SYS_FUNC( accept );
    HOOK_SYS_FUNC( recv );
    HOOK_SYS_FUNC( bind );
}

int socket(int domain, int type, int protocol) {
    return orig_socket_func(domain, type, protocol);
}

int bind (int socketFD, const struct sockaddr *addr, socklen_t length) {
    int errno = orig_bind_func(socketFD,addr, length);
    if (errno == 0 && addr->sa_family == AF_INET) {
        struct sockaddr_in *typed_addr = (struct sockaddr_in *)(addr);
        pid_t thread_id = syscall(__NR_gettid);
        on_bind(thread_id, socketFD, typed_addr);
    }
    return errno;
}

ssize_t send(int socketFD, const void *buffer, size_t size, int flags) {
    ssize_t sent_size = orig_send_func(socketFD, buffer, size, flags);
    if (sent_size >= 0) {
        struct ch_span span;
        span.Ptr = buffer;
        span.Len = sent_size;
        pid_t thread_id = syscall(__NR_gettid);
        on_send(thread_id, socketFD, span, flags);
    }
    return sent_size;
}

ssize_t recv (int socketFD, void *buffer, size_t size, int flags) {
    ssize_t received_size = orig_recv_func(socketFD, buffer, size, flags);
    if (received_size >= 0) {
        struct ch_span span;
        span.Ptr = buffer;
        span.Len = received_size;
        pid_t thread_id = syscall(__NR_gettid);
        on_recv(thread_id, socketFD, span, flags);
    }
    return received_size;
}

ssize_t sendto(int socketFD, const void *buf, size_t len, int flags,
               const struct sockaddr *dest_addr, socklen_t addrlen) {
    return orig_sendto_func(socketFD, buf, len, flags, dest_addr, addrlen);
}

int connect(int socketFD, const struct sockaddr *addr, socklen_t addrlen) {
    int errno = orig_connect_func(socketFD, addr, addrlen);
    if (errno == 0 && addr->sa_family == AF_INET) {
        struct sockaddr_in *typed_addr = (struct sockaddr_in *)(addr);
        pid_t thread_id = syscall(__NR_gettid);
        on_connect(thread_id, socketFD, typed_addr);
    }
    return errno;
}

int accept(int serverSocketFD, struct sockaddr *addr, socklen_t *addrlen) {
    int clientSocketFD = orig_accept_func(serverSocketFD, addr, addrlen);
    if (clientSocketFD > 0 && addr->sa_family == AF_INET) {
        struct sockaddr_in *typed_addr = (struct sockaddr_in *)(addr);
        pid_t thread_id = syscall(__NR_gettid);
        on_accept(thread_id, serverSocketFD, clientSocketFD, typed_addr);
    }
    return clientSocketFD;
}