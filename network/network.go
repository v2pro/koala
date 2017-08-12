package network

import (
	"fmt"
	"net"
)

type SocketFD int

type SendFlags int

func OnSend(socketFD SocketFD, span []byte, flags SendFlags) {
	sock := socks[socketFD]
	if sock == nil {
		fmt.Println("=== send to unknown ===")
		return
	}
	if sock.isServer {
		fmt.Println("=== inbound send ===")
	} else {
		fmt.Println("=== outbound send ===")
	}
	fmt.Println(sock.addr)
	fmt.Println(string(span))
}

type RecvFlags int

func OnRecv(socketFD SocketFD, span []byte, flags RecvFlags) {
	sock := socks[socketFD]
	if sock == nil {
		fmt.Println("=== recv from unknown ===")
		return
	}
	if sock.isServer {
		fmt.Println("=== inbound recv ===")
	} else {
		fmt.Println("=== outbound recv ===")
	}
	fmt.Println(sock.addr)
	fmt.Println(string(span))
}

func OnAccept(serverSocketFD SocketFD, clientSocketFD SocketFD, addr net.TCPAddr) {
	socks[clientSocketFD] = &sock{
		socketFD: clientSocketFD,
		isServer: true,
		addr: addr,
	}
	fmt.Println("=== accept ===")
	fmt.Println(addr)
}

func OnBind(socketFD SocketFD, addr net.TCPAddr) {
	socks[socketFD] = &sock{
		socketFD: socketFD,
		isServer: true,
		addr: addr,
	}
	fmt.Println("=== bind ===")
	fmt.Println(addr)
}
