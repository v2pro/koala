#ifndef _GNU_SOURCE
#define _GNU_SOURCE
#endif

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
#include <sys/types.h>
#include "interpose.h"
#include "span.h"
#include "thread_id.h"
#include "_cgo_export.h"

INTERPOSE(bind)(int socketFD, const struct sockaddr *addr, socklen_t length) {
    auto result = real::bind(socketFD, addr, length);
    if (result == 0 && addr->sa_family == AF_INET) {
        struct sockaddr_in *typed_addr = (struct sockaddr_in *)(addr);
        pid_t thread_id = get_thread_id();
        on_bind(thread_id, socketFD, typed_addr);
    }
    return result;
}

INTERPOSE(send)(int socketFD, const void *buffer, size_t size, int flags) {
    ssize_t sent_size = real::send(socketFD, buffer, size, flags);
    if (sent_size >= 0) {
        struct ch_span span;
        span.Ptr = buffer;
        span.Len = sent_size;
        pid_t thread_id = get_thread_id();
        on_send(thread_id, socketFD, span, flags);
    }
    return sent_size;
}

INTERPOSE(recv)(int socketFD, void *buffer, size_t size, int flags) {
    ssize_t received_size = real::recv(socketFD, buffer, size, flags);
    if (received_size >= 0) {
        struct ch_span span;
        span.Ptr = buffer;
        span.Len = received_size;
        pid_t thread_id = get_thread_id();
        on_recv(thread_id, socketFD, span, flags);
    }
    return received_size;
}

INTERPOSE(sendto)(int socketFD, const void *buffer, size_t buffer_size, int flags,
               const struct sockaddr *addr, socklen_t addr_size) {
    auto result = real::sendto(socketFD, buffer, buffer_size, flags, addr, addr_size);
    if (addr && addr->sa_family == AF_INET) {
        struct sockaddr_in *typed_addr = (struct sockaddr_in *)(addr);
        struct ch_span span;
        span.Ptr = buffer;
        span.Len = buffer_size;
        pid_t thread_id = get_thread_id();
        on_sendto(thread_id, socketFD, span, flags, typed_addr);
    }
    return result;
}

INTERPOSE(connect)(int socketFD, const struct sockaddr *remote_addr, socklen_t remote_addr_len) {
    if (remote_addr->sa_family == AF_INET) {
        struct sockaddr_in *typed_remote_addr = (struct sockaddr_in *)(remote_addr);
        pid_t thread_id = get_thread_id();
        on_connect(thread_id, socketFD, typed_remote_addr);
    }
    return real::connect(socketFD, remote_addr, remote_addr_len);
}

INTERPOSE(accept)(int serverSocketFD, struct sockaddr *addr, socklen_t *addrlen) {
    int clientSocketFD = real::accept(serverSocketFD, addr, addrlen);
    if (clientSocketFD > 0 && addr->sa_family == AF_INET) {
        struct sockaddr_in *typed_addr = (struct sockaddr_in *)(addr);
        pid_t thread_id = get_thread_id();
        on_accept(thread_id, serverSocketFD, clientSocketFD, typed_addr);
    }
    return clientSocketFD;
}