package network

import (
	"fmt"
	"net"
)

type SendFlags int

func (thread *Thread) OnSend(socketFD SocketFD, span []byte, flags SendFlags) {
	sock := thread.lookupSocket(socketFD)
	if sock == nil {
		fmt.Println(fmt.Sprintf("=== [%d] send to unknown ===", thread.threadID))
		return
	}
	if sock.isServer {
		fmt.Println(fmt.Sprintf("=== [%d] inbound send ===", thread.threadID))
	} else {
		fmt.Println(fmt.Sprintf("=== [%d] outbound send ===", thread.threadID))
	}
	fmt.Println(sock.addr)
	fmt.Println(string(span))
}

type RecvFlags int

func (thread *Thread) OnRecv(socketFD SocketFD, span []byte, flags RecvFlags) {
	sock := thread.lookupSocket(socketFD)
	if sock == nil {
		fmt.Println(fmt.Sprintf("=== [%d] recv from unknown ===", thread.threadID))
		return
	}
	if sock.isServer {
		fmt.Println(fmt.Sprintf("=== [%d] inbound recv ===", thread.threadID))
	} else {
		fmt.Println(fmt.Sprintf("=== [%d] outbound recv ===", thread.threadID))
	}
	fmt.Println(sock.addr)
	fmt.Println(string(span))
}

func (thread *Thread) lookupSocket(socketFD SocketFD) *socket {
	sock := thread.socks[socketFD]
	if sock == nil {
		sock = getGlobalSock(socketFD)
		if sock == nil {
			return nil
		}
		thread.socks[socketFD] = sock
	}
	return sock
}

func (thread *Thread) OnAccept(serverSocketFD SocketFD, clientSocketFD SocketFD, addr net.TCPAddr) {
	thread.socks[clientSocketFD] = &socket{
		socketFD: clientSocketFD,
		isServer: true,
		addr:     addr,
	}
	setGlobalSock(clientSocketFD, thread.socks[clientSocketFD])
	fmt.Println(fmt.Sprintf("=== [%d] accept ===", thread.threadID))
	fmt.Println(addr)
}

func (thread *Thread) OnBind(socketFD SocketFD, addr net.TCPAddr) {
	thread.socks[socketFD] = &socket{
		socketFD: socketFD,
		isServer: true,
		addr:     addr,
	}
	fmt.Println(fmt.Sprintf("=== [%d] bind ===", thread.threadID))
	fmt.Println(addr)
}

func (thread *Thread) OnConnect(socketFD SocketFD, addr net.TCPAddr) {
	thread.socks[socketFD] = &socket{
		socketFD: socketFD,
		isServer: false,
		addr:     addr,
	}
	fmt.Println(fmt.Sprintf("=== [%d] connect ===", thread.threadID))
	fmt.Println(addr)
}
