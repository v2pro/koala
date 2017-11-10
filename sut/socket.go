package sut

import (
	"net"
	"time"
)

type socket struct {
	socketFD       SocketFD
	isServer       bool
	addr           net.TCPAddr
	localAddr      *net.TCPAddr
	lastAccessedAt time.Time
	tracerState    *tracerState
}

type tracerState struct {
}

func (sock *socket) canGC(now time.Time) bool {
	if now.Sub(sock.lastAccessedAt) < time.Minute*5 {
		return false
	}
	return true
}

var tracedStreamMagic = []byte{0xde, 0xad, 0xbe, 0xef}

func (sock *socket) beforeSend(traceHeader []byte, span []byte) []byte {
	if sock.tracerState == nil {
		sock.tracerState = &tracerState{}
		extraHeader := append(tracedStreamMagic, []byte{
			byte(len(traceHeader) >> 8),
			byte(len(traceHeader)),
		}...)
		extraHeader = append(extraHeader, traceHeader...)
		extraHeader = append(extraHeader, []byte{
			byte(len(span) >> 8),
			byte(len(span)),
		}...)
		return extraHeader
	}
	return nil
}

func (sock *socket) afterSend(extraHeaderSentSize int, bodySentSize int) {
}
