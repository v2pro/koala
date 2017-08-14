package sut

import (
	"net"
	"github.com/v2pro/koala/countlog"
	"fmt"
	"bytes"
	"encoding/json"
	"context"
	"github.com/v2pro/koala/replaying"
	"encoding/base64"
	"time"
	"github.com/v2pro/koala/recording"
)

var threadShutdownEvent = []byte("to-koala:thread-shutdown||")

func init() {
	logWriter := countlog.NewStdoutLogWriter(countlog.LEVEL_DEBUG)
	logWriter.FormatLog = func(event countlog.Event) string {
		msg := []byte{}
		threadId := getThreadId(event)
		if threadId == nil {
			msg = append(msg, fmt.Sprintf(
				"=== %s ===\n", event.Event)...)
		} else {
			msg = append(msg, fmt.Sprintf(
				"=== [%d] %s ===\n", threadId, event.Event)...)
		}
		for i := 0; i < len(event.Properties); i += 2 {
			k, _ := event.Properties[i].(string)
			if k == "" {
				continue
			}
			v := event.Properties[i+1]
			switch k {
			case "content", "request", "response", "matchedRequest", "matchedResponse":
				buf := v.([]byte)
				isBinary := false
				for _, b := range buf {
					if b == '\r' || b == '\n' {
						continue
					}
					if b < 32 || b > 127 {
						isBinary = true
						break
					}
				}
				if isBinary {
					v = base64.StdEncoding.EncodeToString(buf)
				} else {
					v = string(buf)
				}
			case "addr":
				addr := v.(net.TCPAddr)
				v = addr.String()
			case "threadID":
				continue
			case "lineNumber":
				continue
			case "ctx":
				continue
			case "replayingSession":
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

func getThreadId(event countlog.Event) interface{} {
	threadID := event.Get("threadID")
	if threadID != nil {
		return threadID
	}
	ctx, _ := event.Get("ctx").(context.Context)
	if ctx == nil {
		return nil
	}
	return ctx.Value("threadID")
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
		countlog.Warn("unknown-recv",
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
		isServer: true,
		addr:     addr,
	}
	countlog.Debug("event!sut.bind",
		"threadID", thread.threadID,
		"socketFD", socketFD,
		"addr", addr)
}

func (thread *Thread) OnConnect(socketFD SocketFD, remoteAddr net.TCPAddr) {
	thread.socks[socketFD] = &socket{
		socketFD: socketFD,
		isServer: false,
		addr:     remoteAddr,
	}
	if thread.replayingSession != nil {
		localAddr, err := replaying.BindFDToLocalAddr(int(socketFD))
		if err != nil {
			countlog.Error("event!sut.failed to bind local addr", "err", err)
			return
		}
		thread.socks[socketFD].localAddr = localAddr
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
