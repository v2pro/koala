package network

import (
	"net"
	"github.com/v2pro/koala/countlog"
	"fmt"
	"bytes"
	"encoding/json"
)

var threadShutdownEvent = []byte("to-koala:thread-shutdown||")

func init() {
	logWriter := countlog.NewStdoutLogWriter(countlog.DEBUG)
	logWriter.FormatLog = func(event countlog.Event) string {
		msg := []byte{}
		msg = append(msg, fmt.Sprintf(
			"=== [%d] %s ===\n", event.Get("threadID"), event.Event)...)
		for i := 0; i < len(event.Properties); i += 2 {
			k, _ := event.Properties[i].(string)
			if k == "" {
				continue
			}
			v := event.Properties[i+1]
			switch k {
			case "content":
				v = string(v.([]byte))
			case "addr":
				addr := v.(net.TCPAddr)
				v = addr.String()
			case "threadID":
				continue
			case "lineNumber":
				continue
			case "session":
				b, err := json.MarshalIndent(v, "", "  ")
				if err != nil {
					panic(err)
				}
				v = string(b)
			}
			msg = append(msg, fmt.Sprintf("%s: %v\n", k, v)...)
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
	if sock.isServer {
		thread.session.InboundSend(span, sock.addr)
	} else {
		event = "outbound-send"
		thread.session.OutboundSend(thread.threadID, span, sock.addr)
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
	if sock.isServer {
		thread.session.InboundRecv(span, sock.addr)
	} else {
		event = "outbound-recv"
		thread.session.OutboundRecv(thread.threadID, span, sock.addr)
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

type SendToFlags int

func (thread *Thread) OnSendTo(socketFD SocketFD, span []byte, flags SendToFlags, addr net.TCPAddr) {
	countlog.Debug("sendto",
		"threadID", thread.threadID,
		"socketFD", socketFD,
		"addr", addr,
		"content", span)
	if bytes.HasPrefix(span, threadShutdownEvent) {
		thread.session.OutboundTalks = append(thread.session.OutboundTalks, thread.session.currentOutboundTalk)
		countlog.Fatal("session-produced",
			"threadID", thread.threadID,
			"session", thread.session,
		)
		countlog.Debug("thread-shutdown",
			"threadID", thread.threadID)
	}
}
