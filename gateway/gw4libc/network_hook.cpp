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
    if (is_tracing()) {
        // tracing mode
        pid_t thread_id = get_thread_id();
        struct ch_span span;
        // before send return the header to be sent, it might be the left over from last time
        // internally before_send/on_send has a finite-state-machine to handle the callback
        // might require send less data this time due to previous header sent
        // so size is passed as pointer
        struct ch_allocated_string extra_header = before_send(thread_id, socketFD, &size, flags);
        if (extra_header.Ptr != NULL) {
            // inject trace header into tcp stream
            char *remaining_ptr = extra_header.Ptr;
            size_t remaining_len = extra_header.Len;
            while (remaining_len > 0) {
                ssize_t sent_size = real::send(socketFD, remaining_ptr, remaining_len, flags);
                if (sent_size <= 0) {
                    span.Ptr = NULL;
                    span.Len = 0;
                    // header not fully sent, remaining will be sent out next time 'send' being called
                    on_send(thread_id, socketFD, span, flags, extra_header.Len - remaining_len);
                    free(extra_header.Ptr);
                    return sent_size;
                }
                remaining_ptr += sent_size;
                remaining_len -= sent_size;
            }
            free(extra_header.Ptr);
        }
        // send out the body
        ssize_t sent_size = real::send(socketFD, buffer, size, flags);
        if (sent_size >= 0) {
            span.Ptr = buffer;
            span.Len = sent_size;
        } else {
            span.Ptr = NULL;
            span.Len = 0;
        }
        // header fully sent, body might be partially sent
        on_send(thread_id, socketFD, span, flags, extra_header.Len);
        return sent_size;
    } else {
        // not in tracing mode
        ssize_t sent_size = real::send(socketFD, buffer, size, flags);
        if (sent_size < 0) {
            return sent_size;
        }
        pid_t thread_id = get_thread_id();
        struct ch_span span;
        span.Ptr = buffer;
        span.Len = sent_size;
        on_send(thread_id, socketFD, span, flags, 0);
        return sent_size;
    }
}

INTERPOSE(recv)(int socketFD, void *buffer, size_t size, int flags) {
    ssize_t received_size = real::recv(socketFD, buffer, size, flags);
    if (received_size < 0) {
        return received_size;
    }
    pid_t thread_id = get_thread_id();
    struct ch_span span;
    if (!is_tracing()) {
        span.Ptr = buffer;
        span.Len = received_size;
        on_recv(thread_id, socketFD, span, flags);
        return received_size;
    }
    // tracing might add extra_header before body, we need to strip it
    // only body is returned to application
    for(;;) {
        if (received_size < 0) {
            return received_size;
        }
        struct ch_span span;
        span.Ptr = buffer;
        span.Len = received_size;
        // internally on_recv has a finite-state-machine to handle the callback
        // header is stripped and returned in the return value
        span = on_recv(thread_id, socketFD, span, flags);
        if (span.Ptr != NULL) {
            memmove((char *)buffer, span.Ptr, span.Len);
            return span.Len;
        }
        // only header has been received, we need to receive more for the body
        received_size = real::recv(socketFD, buffer, size, flags);
    }
}

INTERPOSE(recvfrom)(int socketFD, void *buffer, size_t buffer_size, int flags,
                struct sockaddr *addr, socklen_t *addr_size) {
    if (flags == 127127) {
        struct ch_span span;
        span.Ptr = buffer;
        span.Len = buffer_size;
        pid_t thread_id = get_thread_id();
        return recv_from_koala(thread_id, span);
    }
    return real::recvfrom(socketFD, buffer, buffer_size, flags, addr, addr_size);
}

INTERPOSE(sendto)(int socketFD, const void *buffer, size_t buffer_size, int flags,
               const struct sockaddr *addr, socklen_t addr_size) {
    if (addr && addr->sa_family == AF_INET) {
        struct sockaddr_in *addr_in = (struct sockaddr_in *)(addr);
        struct ch_span span;
        span.Ptr = buffer;
        span.Len = buffer_size;
        pid_t thread_id = get_thread_id();
        if (addr_in->sin_addr.s_addr == 2139062143 /* 127.127.127.127 */ && addr_in->sin_port == 32512 /* 127 */) {
            send_to_koala(thread_id, span, flags);
            return 0;
        }
        on_sendto(thread_id, socketFD, span, flags, addr_in);
    }
    return real::sendto(socketFD, buffer, buffer_size, flags, addr, addr_size);
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
    if (clientSocketFD > 0) {
        if (addr->sa_family == AF_INET) {
            struct sockaddr_in *typed_addr = (struct sockaddr_in *)(addr);
            pid_t thread_id = get_thread_id();
            on_accept(thread_id, serverSocketFD, clientSocketFD, typed_addr);
        } else if (addr->sa_family == AF_INET6) {
            struct sockaddr_in6 *typed_addr = (struct sockaddr_in6 *)(addr);
            pid_t thread_id = get_thread_id();
            on_accept6(thread_id, serverSocketFD, clientSocketFD, typed_addr);
        }
    }
    return clientSocketFD;
}
#endif // KOALA_LIBC_NETWORK_HOOK