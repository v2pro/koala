package network

import "net"

type sock struct {
	socketFD SocketFD
	isServer bool
	addr net.TCPAddr
}

var socks = map[SocketFD]*sock{}