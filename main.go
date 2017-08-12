package main

// #cgo LDFLAGS: -ldl
// #include <stddef.h>
// #include <netinet/in.h>
// #include <sys/types.h>
// #include <sys/socket.h>
// #include "span.h"
// #include "network_hook.h"
import "C"
import (
	"github.com/v2pro/koala/ch"
	"syscall"
	"net"
	"github.com/v2pro/koala/inbound"
	"github.com/v2pro/koala/sut"
	"github.com/v2pro/koala/outbound"
	"github.com/v2pro/koala/countlog"
)

func init() {
	C.libc_hook_init()
	inbound.Start()
	outbound.Start()
}

//export on_connect
func on_connect(threadID C.pid_t, socketFD C.int, remoteAddr *C.struct_sockaddr_in) {
	if sockaddr_in_sin_family_get(remoteAddr) != syscall.AF_INET {
		panic("expect ipv4 remoteAddr")
	}
	sut.GetThread(sut.ThreadID(threadID)).
		OnConnect(sut.SocketFD(socketFD), net.TCPAddr{
		IP:   ch.Int2ip(sockaddr_in_sin_addr_get(remoteAddr)),
		Port: int(ch.Ntohs(sockaddr_in_sin_port_get(remoteAddr))),
	})
	redirectTo, err := net.ResolveTCPAddr("tcp", "127.0.0.1:9002")
	if err != nil {
		countlog.Error("failed to resolve redirect to remoteAddr", "err", err)
		return
	}
	sockaddr_in_sin_addr_set(remoteAddr, ch.Ip2int(redirectTo.IP))
	sockaddr_in_sin_port_set(remoteAddr, ch.Htons(uint16(redirectTo.Port)))
}

//export on_bind
func on_bind(threadID C.pid_t, socketFD C.int, addr *C.struct_sockaddr_in) {
	if sockaddr_in_sin_family_get(addr) != syscall.AF_INET {
		panic("expect ipv4 addr")
	}
	sut.GetThread(sut.ThreadID(threadID)).
		OnBind(sut.SocketFD(socketFD), net.TCPAddr{
		IP:   ch.Int2ip(sockaddr_in_sin_addr_get(addr)),
		Port: int(ch.Ntohs(sockaddr_in_sin_port_get(addr))),
	})
}

//export on_accept
func on_accept(threadID C.pid_t, serverSocketFD C.int, clientSocketFD C.int, addr *C.struct_sockaddr_in) {
	if sockaddr_in_sin_family_get(addr) != syscall.AF_INET {
		panic("expect ipv4 addr")
	}
	sut.GetThread(sut.ThreadID(threadID)).
		OnAccept(sut.SocketFD(serverSocketFD), sut.SocketFD(clientSocketFD), net.TCPAddr{
		IP:   ch.Int2ip(sockaddr_in_sin_addr_get(addr)),
		Port: int(ch.Ntohs(sockaddr_in_sin_port_get(addr))),
	})
}

//export on_send
func on_send(threadID C.pid_t, socketFD C.int, span C.struct_ch_span, flags C.int) {
	sut.GetThread(sut.ThreadID(threadID)).
		OnSend(sut.SocketFD(socketFD), ch_span_to_bytes(span), sut.SendFlags(flags))
}

//export on_recv
func on_recv(threadID C.pid_t, socketFD C.int, span C.struct_ch_span, flags C.int) {
	sut.GetThread(sut.ThreadID(threadID)).
		OnRecv(sut.SocketFD(socketFD), ch_span_to_bytes(span), sut.RecvFlags(flags))
}

//export on_sendto
func on_sendto(threadID C.pid_t, socketFD C.int, span C.struct_ch_span, flags C.int, addr *C.struct_sockaddr_in) {
	sut.GetThread(sut.ThreadID(threadID)).
		OnSendTo(sut.SocketFD(socketFD), ch_span_to_bytes(span), sut.SendToFlags(flags), net.TCPAddr{
		IP:   ch.Int2ip(sockaddr_in_sin_addr_get(addr)),
		Port: int(ch.Ntohs(sockaddr_in_sin_port_get(addr))),
	})
}

func main() {
}
