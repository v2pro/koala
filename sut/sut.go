package sut

import (
	"net"
	"github.com/v2pro/koala/countlog"
	"bytes"
	"github.com/v2pro/koala/replaying"
	"time"
	"github.com/v2pro/koala/recording"
	"syscall"
)

var threadShutdownEvent = []byte("to-koala:thread-shutdown||")

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

type SendFlags int

func (thread *Thread) OnSend(socketFD SocketFD, span []byte, flags SendFlags) {
	sock := thread.lookupSocket(socketFD)
	if sock == nil {
		localAddr, err := syscall.Getsockname(int(socketFD))
		if err != nil {
			countlog.Error("event!sut.failed to find local address of new socket",
				"ctx", thread, "err", err)
			return
		}
		localAddr4, _ := localAddr.(*syscall.SockaddrInet4)
		remoteAddr, err := syscall.Getpeername(int(socketFD))
		if err != nil {
			countlog.Error("event!sut.failed to find remote address of new socket",
				"ctx", thread, "err", err)
			return
		}
		remoteAddr4, _ := remoteAddr.(*syscall.SockaddrInet4)
		if remoteAddr4 == nil {
			return
		}
		countlog.Debug("event!sut.found_new_socket_on_send",
			"threadID", thread.threadID,
			"socketFD", socketFD)
		sock = &socket{
			socketFD: socketFD,
			isServer: false,
			addr: net.TCPAddr{
				IP:   remoteAddr4.Addr[:],
				Port: remoteAddr4.Port,
			},
			localAddr: &net.TCPAddr{
				IP:   localAddr4.Addr[:],
				Port: localAddr4.Port,
			},
		}
		thread.socks[socketFD] = sock
	}
	event := "event!sut.inbound_send"
	if sock.isServer {
		thread.recordingSession.InboundSend(thread, span, sock.addr)
	} else {
		event = "event!sut.outbound_send"
		thread.recordingSession.OutboundSend(thread, span, sock.addr)
		if sock.localAddr != nil {
			replaying.StoreTmp(*sock.localAddr, thread.replayingSession)
		}
	}
	countlog.Trace(event,
		"threadID", thread.threadID,
		"socketFD", socketFD,
		"addr", sock.addr,
		"content", span)
}

type RecvFlags int

func (thread *Thread) OnRecv(socketFD SocketFD, span []byte, flags RecvFlags) {
	sock := thread.lookupSocket(socketFD)
	if sock == nil {
		countlog.Warn("event!sut.unknown-recv",
			"threadID", thread.threadID,
			"socketFD", socketFD)
		return
	}
	event := "event!sut.inbound_recv"
	if sock.isServer {
		if thread.recordingSession.HasResponded() {
			thread.recordingSession.Shutdown(thread)
			thread.recordingSession = &recording.Session{}
		}
		thread.recordingSession.InboundRecv(thread, span, sock.addr)
		replayingSession := replaying.RetrieveTmp(sock.addr)
		if replayingSession != nil {
			nanoOffset := replayingSession.InboundTalk.RequestTime - time.Now().UnixNano()
			SetTimeOffset(int(time.Duration(nanoOffset) / time.Second))
			thread.replayingSession = replayingSession
			countlog.Trace("event!sut.received_replaying_session",
				"threadID", thread.threadID,
				"replayingSession", thread.replayingSession,
				"addr", sock.addr)
		}
	} else {
		event = "event!sut.outbound_recv"
		thread.recordingSession.OutboundRecv(thread, span, sock.addr)
	}
	countlog.Trace(event,
		"threadID", thread.threadID,
		"socketFD", socketFD,
		"addr", sock.addr,
		"content", span)
}

func (thread *Thread) OnAccept(serverSocketFD SocketFD, clientSocketFD SocketFD, addr net.TCPAddr) {
	thread.socks[clientSocketFD] = &socket{
		socketFD: clientSocketFD,
		isServer: true,
		addr:     addr,
	}
	setGlobalSock(clientSocketFD, thread.socks[clientSocketFD])
	countlog.Debug("event!sut.accept",
		"threadID", thread.threadID,
		"serverSocketFD", serverSocketFD,
		"clientSocketFD", clientSocketFD,
		"addr", addr)
}

func (thread *Thread) OnBind(socketFD SocketFD, addr net.TCPAddr) {
	thread.socks[socketFD] = &socket{
		socketFD: socketFD,
		isServer: false,
		addr:     addr,
	}
	countlog.Debug("event!sut.bind",
		"threadID", thread.threadID,
		"socketFD", socketFD,
		"addr", &addr)
}

func (thread *Thread) OnConnect(socketFD SocketFD, remoteAddr net.TCPAddr) {
	thread.socks[socketFD] = &socket{
		socketFD: socketFD,
		isServer: false,
		addr:     remoteAddr,
	}
	localAddr, err := replaying.BindFDToLocalAddr(int(socketFD))
	if err != nil {
		countlog.Error("event!sut.failed to bind local addr", "err", err)
		return
	}
	thread.socks[socketFD].localAddr = localAddr
	if thread.replayingSession != nil {
		replaying.StoreTmp(*localAddr, thread.replayingSession)
	}
	countlog.Debug("event!sut.connect",
		"threadID", thread.threadID,
		"socketFD", socketFD,
		"addr", remoteAddr,
		"localAddr", thread.socks[socketFD].localAddr)
}

type SendToFlags int

func (thread *Thread) OnSendTo(socketFD SocketFD, span []byte, flags SendToFlags, addr net.TCPAddr) {
	if addr.String() != "127.127.127.127:127" {
		return
	}
	helperInfo := span
	countlog.Debug("event!sut.received_helper_info",
		"threadID", thread.threadID,
		"socketFD", socketFD,
		"addr", addr,
		"content", helperInfo)
	if bytes.HasPrefix(helperInfo, threadShutdownEvent) {
		thread.recordingSession.Shutdown(thread)
		countlog.Debug("event!sut.thread_shutdown",
			"threadID", thread.threadID)
	}
}
