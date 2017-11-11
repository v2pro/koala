package sut

import (
	"net"
	"time"
	"github.com/v2pro/koala/recording"
	"bytes"
	"github.com/v2pro/plz/countlog"
	"encoding/binary"
	"encoding/hex"
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
	isTraced bool
	nextAction string
	buffered []byte
	expectedBufferSize uint16
}

func (sock *socket) canGC(now time.Time) bool {
	if now.Sub(sock.lastAccessedAt) < time.Minute*5 {
		return false
	}
	return true
}

var magicInit = []byte{0xde, 0xad, 0xbe, 0xef, 0x01}
var magicSameTrace = []byte{0xde, 0xad}
var magicChangeTrace = []byte{0xbe, 0xef}

func (sock *socket) beforeSend(session *recording.Session, span []byte) []byte {
	traceHeader := session.GetTraceHeader()
	if sock.tracerState == nil {
		sock.tracerState = &tracerState{}
		extraHeader := append(magicInit, []byte{
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

func (sock *socket) afterSend(session *recording.Session, extraHeaderSentSize int, bodySentSize int) {
}

func (sock *socket) onRecv(session *recording.Session, span []byte) []byte {
	if sock.tracerState == nil {
		return sock.onRecv_initial(session, span)
	}
	switch sock.tracerState.nextAction {
	case "readTraceHeaderSize":
		return sock.onRecv_readTraceHeaderSize(session, span)
	case "readTraceHeader":
		return sock.onRecv_readTraceHeader(session, span)
	case "readBodySize":
		return sock.onRecv_readBodySize(session, span)
	case "readBody":
		return sock.onRecv_readBody(session, span)
	case "readMagic":
		return sock.onRecv_readMagic(session, span)
	default:
		countlog.Error("event!sock.onRecv_dispatch",
			"nextAction", sock.tracerState.nextAction,
			"socketFD", sock.socketFD)
		return nil
	}
}

func (sock *socket) onRecv_initial(session *recording.Session, span []byte) []byte {
	sock.tracerState = &tracerState{}
	if len(span) < 5 {
		sock.tracerState.isTraced = false
		countlog.Trace("event!sock.onRecv_initial.span too small", "socketFD", sock.socketFD)
		return span
	}
	if !bytes.Equal(magicInit[:5], span[:5]) {
		sock.tracerState.isTraced = false
		countlog.Trace("event!sock.onRecv_initial.not starts with magic", "socketFD", sock.socketFD)
		return span
	}
	sock.tracerState.nextAction = "readTraceHeaderSize"
	return sock.onRecv_readTraceHeaderSize(session, span[5:])
}

func (sock *socket) onRecv_readTraceHeaderSize(session *recording.Session, span []byte) []byte {
	alreadyRead := len(sock.tracerState.buffered)
	toRead := 2 - alreadyRead
	if len(span) < toRead {
		toRead = len(span)
		sock.tracerState.buffered = append(sock.tracerState.buffered, span[:toRead]...)
		return nil
	}
	sock.tracerState.buffered = append(sock.tracerState.buffered, span[:toRead]...)
	sock.tracerState.expectedBufferSize = binary.BigEndian.Uint16(sock.tracerState.buffered)
	sock.tracerState.buffered = nil
	sock.tracerState.nextAction = "readTraceHeader"
	return sock.onRecv_readTraceHeader(session, span[toRead:])
}

func (sock *socket) onRecv_readTraceHeader(session *recording.Session, span []byte) []byte {
	alreadyRead := len(sock.tracerState.buffered)
	toRead := int(sock.tracerState.expectedBufferSize) - alreadyRead
	if len(span) < toRead {
		toRead = len(span)
		sock.tracerState.buffered = append(sock.tracerState.buffered, span[:toRead]...)
		return nil
	}
	sock.tracerState.buffered = append(sock.tracerState.buffered, span[:toRead]...)
	session.TraceHeader = sock.tracerState.buffered
	sock.tracerState.expectedBufferSize = 0
	sock.tracerState.buffered = nil
	sock.tracerState.nextAction = "readBodySize"
	return sock.onRecv_readBodySize(session, span[toRead:])
}

func (sock *socket) onRecv_readBodySize(session *recording.Session, span []byte) []byte {
	alreadyRead := len(sock.tracerState.buffered)
	toRead := 2 - alreadyRead
	if len(span) < toRead {
		toRead = len(span)
		sock.tracerState.buffered = append(sock.tracerState.buffered, span[:toRead]...)
		return nil
	}
	sock.tracerState.buffered = append(sock.tracerState.buffered, span[:toRead]...)
	sock.tracerState.expectedBufferSize = binary.BigEndian.Uint16(sock.tracerState.buffered)
	sock.tracerState.buffered = nil
	sock.tracerState.nextAction = "readBody"
	return sock.onRecv_readBody(session, span[toRead:])
}

func (sock *socket) onRecv_readBody(session *recording.Session, span []byte) []byte {
	if len(span) < int(sock.tracerState.expectedBufferSize) {
		sock.tracerState.expectedBufferSize -= uint16(len(span))
		return span
	}
	bodySize := int(sock.tracerState.expectedBufferSize)
	sock.tracerState.expectedBufferSize = 0
	sock.tracerState.buffered = nil
	sock.tracerState.nextAction = "readMagic"
	moreBody := sock.onRecv_readMagic(session, span[bodySize:])
	if moreBody == nil {
		return span[:bodySize]
	}
	copy(span[bodySize:], moreBody)
	bodySize += len(moreBody)
	return span[:bodySize]
}

func (sock *socket) onRecv_readMagic(session *recording.Session, span []byte) []byte {
	alreadyRead := len(sock.tracerState.buffered)
	toRead := 2 - alreadyRead
	if len(span) < toRead {
		toRead = len(span)
		sock.tracerState.buffered = append(sock.tracerState.buffered, span[:toRead]...)
		return nil
	}
	sock.tracerState.buffered = append(sock.tracerState.buffered, span[:toRead]...)
	if bytes.Equal(magicSameTrace, sock.tracerState.buffered) {
		sock.tracerState.expectedBufferSize = 0
		sock.tracerState.buffered = nil
		sock.tracerState.nextAction = "readBodySize"
		return sock.onRecv_readBodySize(session, span[toRead:])

	} else if bytes.Equal(magicChangeTrace, sock.tracerState.buffered) {
		sock.tracerState.expectedBufferSize = 0
		sock.tracerState.buffered = nil
		sock.tracerState.nextAction = "readTraceHeaderSize"
		return sock.onRecv_readTraceHeaderSize(session, span[toRead:])
	} else {
		countlog.Error("event!sock.onRecv_readMagic.unexpected",
			"magic", hex.EncodeToString(sock.tracerState.buffered),
				"socketFD", sock.socketFD)
		return nil
	}
}