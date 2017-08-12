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
func on_connect(sockfd C.int, ip *C.char, port C.int) {
}

//export on_bind
func on_bind(socketFD C.int, addr *C.struct_sockaddr_in) {
	if sockaddr_in_sin_family_get(addr) != syscall.AF_INET {
		panic("expect ipv4 addr")
	}
	network.OnBind(network.SocketFD(socketFD), net.TCPAddr{
		IP:   ch.Int2ip(sockaddr_in_sin_addr_get(addr)),
		Port: int(ch.Ntohs(sockaddr_in_sin_port_get(addr))),
	})
}

//export on_accept
func on_accept(serverSocketFD C.int, clientSocketFD C.int, addr *C.struct_sockaddr_in) {
	if sockaddr_in_sin_family_get(addr) != syscall.AF_INET {
		panic("expect ipv4 addr")
	}
	network.OnAccept(network.SocketFD(serverSocketFD), network.SocketFD(clientSocketFD), net.TCPAddr{
		IP:   ch.Int2ip(sockaddr_in_sin_addr_get(addr)),
		Port: int(ch.Ntohs(sockaddr_in_sin_port_get(addr))),
	})
}

//export on_send
func on_send(socketFD C.int, span C.struct_ch_span, flags C.int) {
	network.OnSend(network.SocketFD(socketFD), ch_span_to_bytes(span), network.SendFlags(flags))
}

//export on_recv
func on_recv(socketFD C.int, span C.struct_ch_span, flags C.int) {
	network.OnRecv(network.SocketFD(socketFD), ch_span_to_bytes(span), network.RecvFlags(flags))
}

func main() {
}
