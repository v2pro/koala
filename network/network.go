package network

import (
	"net"
	"github.com/v2pro/koala/countlog"
	"fmt"
)

func init() {
	logWriter := countlog.NewStdoutLogWriter(countlog.DEBUG)
	logWriter.FormatLog = func(event countlog.Event) string {
		msg := []byte{}
		msg = append(msg, fmt.Sprintf(
			"=== [%d] %s ===\n", event.Get("threadID"), event.Event)...)
		msg = append(msg, fmt.Sprintf("addr: %v\n", event.Get("addr"))...)
		content, _ := event.Get("content").([]byte)
		if content != nil {
			msg = append(msg, fmt.Sprintf("%v\n", string(content))...)
		}
		return string(msg)
	}
	logWriter.Start()
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

type SendFlags int

func (thread *Thread) OnSend(socketFD SocketFD, span []byte, flags SendFlags) {
	sock := thread.lookupSocket(socketFD)
	if sock == nil {
		countlog.Warn("unknown-send",
			"threadID", thread.threadID,
			"socketFD", socketFD)
		return
	}
	event := "inbound-send"
	if !sock.isServer {
		event = "outbound-send"
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
		countlog.Warn("unknown-recv",
			"threadID", thread.threadID,
			"socketFD", socketFD)
		return
	}
	event := "inbound-recv"
	if !sock.isServer {
		event = "outbound-recv"
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
	countlog.Debug("accept",
		"threadID", thread.threadID,
		"serverSocketFD", serverSocketFD,
		"clientSocketFD", clientSocketFD,
		"addr", addr)
}

func (thread *Thread) OnBind(socketFD SocketFD, addr net.TCPAddr) {
	thread.socks[socketFD] = &socket{
		socketFD: socketFD,
		isServer: true,
		addr:     addr,
	}
	countlog.Debug("bind",
		"threadID", thread.threadID,
		"socketFD", socketFD,
		"addr", addr)
}

func (thread *Thread) OnConnect(socketFD SocketFD, addr net.TCPAddr) {
	thread.socks[socketFD] = &socket{
		socketFD: socketFD,
		isServer: false,
		addr:     addr,
	}
	countlog.Debug("connect",
		"threadID", thread.threadID,
		"socketFD", socketFD,
		"addr", addr)
}
