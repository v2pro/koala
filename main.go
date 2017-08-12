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
	"github.com/v2pro/koala/network"
	"net"
)

func init() {
	C.libc_hook_init()
}

//export on_connect
func on_connect(threadID C.pid_t, socketFD C.int, addr *C.struct_sockaddr_in) {
	if sockaddr_in_sin_family_get(addr) != syscall.AF_INET {
		panic("expect ipv4 addr")
	}
	network.GetThread(network.ThreadID(threadID)).
		OnConnect(network.SocketFD(socketFD), net.TCPAddr{
		IP:   ch.Int2ip(sockaddr_in_sin_addr_get(addr)),
		Port: int(ch.Ntohs(sockaddr_in_sin_port_get(addr))),
	})
}

//export on_bind
func on_bind(threadID C.pid_t, socketFD C.int, addr *C.struct_sockaddr_in) {
	if sockaddr_in_sin_family_get(addr) != syscall.AF_INET {
		panic("expect ipv4 addr")
	}
	network.GetThread(network.ThreadID(threadID)).
		OnBind(network.SocketFD(socketFD), net.TCPAddr{
		IP:   ch.Int2ip(sockaddr_in_sin_addr_get(addr)),
		Port: int(ch.Ntohs(sockaddr_in_sin_port_get(addr))),
	})
}

//export on_accept
func on_accept(threadID C.pid_t, serverSocketFD C.int, clientSocketFD C.int, addr *C.struct_sockaddr_in) {
	if sockaddr_in_sin_family_get(addr) != syscall.AF_INET {
		panic("expect ipv4 addr")
	}
	network.GetThread(network.ThreadID(threadID)).
		OnAccept(network.SocketFD(serverSocketFD), network.SocketFD(clientSocketFD), net.TCPAddr{
		IP:   ch.Int2ip(sockaddr_in_sin_addr_get(addr)),
		Port: int(ch.Ntohs(sockaddr_in_sin_port_get(addr))),
	})
}

//export on_send
func on_send(threadID C.pid_t, socketFD C.int, span C.struct_ch_span, flags C.int) {
	network.GetThread(network.ThreadID(threadID)).
		OnSend(network.SocketFD(socketFD), ch_span_to_bytes(span), network.SendFlags(flags))
}

//export on_recv
func on_recv(threadID C.pid_t, socketFD C.int, span C.struct_ch_span, flags C.int) {
	network.GetThread(network.ThreadID(threadID)).
		OnRecv(network.SocketFD(socketFD), ch_span_to_bytes(span), network.RecvFlags(flags))
}

//export on_sendto
func on_sendto(threadID C.pid_t, socketFD C.int, span C.struct_ch_span, flags C.int, addr *C.struct_sockaddr_in) {
	network.GetThread(network.ThreadID(threadID)).
		OnSendTo(network.SocketFD(socketFD), ch_span_to_bytes(span), network.SendToFlags(flags), net.TCPAddr{
		IP:   ch.Int2ip(sockaddr_in_sin_addr_get(addr)),
		Port: int(ch.Ntohs(sockaddr_in_sin_port_get(addr))),
	})
}

func main() {
}
