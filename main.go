package main

// #cgo LDFLAGS: -ldl
// #include <stddef.h>
// #include <netinet/in.h>
// #include <sys/types.h>
// #include <sys/socket.h>
// #include "network_hook.h"
import "C"
import (
	"unsafe"
	"fmt"
	"github.com/v2pro/koala/ch"
)

func init() {
	C.libc_hook_init()
}

//export on_connect
func on_connect(sockfd C.int, ip *C.char, port C.int) {
}

//export on_accept
func on_accept(serverSockFd C.int, clientSockFd C.int, sin *C.struct_sockaddr_in) {
	//sockaddr_in.Get_sin_family((unsafe.Pointer)(sin))
	fmt.Println(sockaddr_in_sin_family_get(sin))
	fmt.Println(ch.Ntohs(sockaddr_in_sin_port_get(sin)))
	fmt.Println(ch.Int2ip(sockaddr_in_sin_addr_get(sin)))
}

//export on_send
func on_send(sockfd C.int, buf unsafe.Pointer, len C.int) {
}

func main() {
}
