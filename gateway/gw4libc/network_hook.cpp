#ifdef KOALA_LIBC_NETWORK_HOOK
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
    pid_t thread_id = get_thread_id();
    struct ch_span span;
    ssize_t body_sent_size;
    if (!is_tracing()) {
        body_sent_size = real::send(socketFD, buffer, size, flags);
        if (body_sent_size >= 0) {
            span.Ptr = buffer;
            span.Len = body_sent_size;
            on_send(thread_id, socketFD, span, flags, 0);
        }
        return body_sent_size;
    }
    // tracing might add extra_header before body
    span.Ptr = buffer;
    span.Len = size;
    struct ch_allocated_string extra_header = before_send(thread_id, socketFD, &span, flags);
    size = span.Len; // might require send less data this time due to previous header sent
    if (extra_header.Ptr != NULL) {
        char *remaining_ptr = extra_header.Ptr;
        size_t remaining_len = extra_header.Len;
        while (remaining_len > 0) {
            ssize_t sent_size = real::send(socketFD, remaining_ptr, remaining_len, flags);
            if (sent_size <= 0) {
                span.Ptr = NULL;
                span.Len = 0;
                on_send(thread_id, socketFD, span, flags, extra_header.Len - remaining_len);
                return sent_size;
            }
            remaining_ptr += sent_size;
            remaining_len -= sent_size;
        }
        free(extra_header.Ptr);
    }
    body_sent_size = real::send(socketFD, buffer, size, flags);
    if (body_sent_size >= 0) {
        span.Ptr = buffer;
        span.Len = body_sent_size;
    } else {
        span.Ptr = NULL;
        span.Len = 0;
    }
    on_send(thread_id, socketFD, span, flags, extra_header.Len);
    return body_sent_size;
}

INTERPOSE(recv)(int socketFD, void *buffer, size_t size, int flags) {
    pid_t thread_id = get_thread_id();
    struct ch_span span;
    if (!is_tracing()) {
        ssize_t body_received_size = real::recv(socketFD, buffer, size, flags);
        if (body_received_size >= 0) {
            span.Ptr = buffer;
            span.Len = body_received_size;
            on_recv(thread_id, socketFD, span, flags);
        }
        return body_received_size;
    }
    // tracing might add extra_header before body
    for(;;) {
        ssize_t received_size = real::recv(socketFD, buffer, size, flags);
        if (received_size >= 0) {
            struct ch_span span;
            span.Ptr = buffer;
            span.Len = received_size;
            pid_t thread_id = get_thread_id();
            span = on_recv(thread_id, socketFD, span, flags);
            if (span.Ptr != NULL) {
                memmove((char *)buffer, span.Ptr, span.Len);
                return span.Len;
            }
            // continue receive more header
        }
        return received_size;
    }
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
#endif // KOALA_LIBC