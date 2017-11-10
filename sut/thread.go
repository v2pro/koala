package sut

import (
	"bytes"
	"github.com/v2pro/koala/envarg"
	"github.com/v2pro/koala/recording"
	"github.com/v2pro/koala/replaying"
	"github.com/v2pro/koala/trace"
	"github.com/v2pro/plz/countlog"
	"net"
	"os"
	"strings"
	"time"
	"sync"
	"context"
)

// InboundRequestPrefix is used to recognize php-fpm FCGI_BEGIN_REQUEST packet.
// fastcgi_finish_request() will send STDOUT first, then recv STDIN (if POST body has not been read before)
// this behavior will trigger session shutdown as we are going to think the recv STDIN
// is the beginning of next request.
// Set InboundRequestPrefix to []byte{1, 1} to only begin new session for FCGI_BEGIN_REQUEST.
// First 0x01 is the version field of fastcgi protocol, second 0x01 is FCGI_BEGIN_REQUEST.
var InboundRequestPrefix = []byte{}

type Thread struct {
	context.Context
	mutex            *sync.Mutex
	threadID         ThreadID
	socks            map[SocketFD]*socket
	files            map[FileFD]*file
	recordingSession *recording.Session
	replayingSession *replaying.ReplayingSession
	lastAccessedAt   time.Time
}

type SendFlags int

func (thread *Thread) BeforeSend(socketFD SocketFD, span []byte, flags SendFlags) []byte {
	if !envarg.IsTracing() {
		return nil
	}
	if thread.recordingSession == nil {
		return nil
	}
	sock := thread.lookupSocket(socketFD)
	if sock == nil {
		countlog.Warn("event!sut.unknown-before-send",
			"threadID", thread.threadID,
			"socketFD", socketFD)
		return nil
	}
	if sock.isServer {
		return nil
	}
	traceHeader := thread.recordingSession.GetTraceHeader()
	extraHeader := sock.beforeSend(traceHeader, span)
	if extraHeader != nil {
		countlog.Trace("event!sut.before_send",
			"socketFD", socketFD,
			"threadID", thread.threadID,
			"content", span,
			"extraHeader", extraHeader)
	}
	return extraHeader
}

func (thread *Thread) OnSend(socketFD SocketFD, span []byte, flags SendFlags) {
	if len(span) == 0 {
		return
	}
	sock := thread.lookupSocket(socketFD)
	if sock == nil {
		countlog.Warn("event!sut.unknown-send",
			"threadID", thread.threadID,
			"socketFD", socketFD)
		return
	}
	event := "event!sut.inbound_send"
	if sock.isServer {
		thread.recordingSession.SendToInbound(thread, span, sock.addr)
	} else {
		event = "event!sut.outbound_send"
		thread.recordingSession.SendToOutbound(thread, span, sock.addr, sock.localAddr, int(sock.socketFD))
		if thread.replayingSession != nil {
			if sock.localAddr != nil {
				replaying.StoreTmp(*sock.localAddr, thread.replayingSession)
			} else {
				countlog.Error("event!sut.can not store replaying session due to no local addr",
					"threadID", thread.threadID)
			}
		}
	}
	countlog.Trace(event,
		"threadID", thread.threadID,
		"socketFD", socketFD,
		"addr", &sock.addr,
		"content", span)
}

func (thread *Thread) AfterSend(socketFD SocketFD, extraHeaderSentSize int, bodySentSize int) {
	if !envarg.IsTracing() {
		return
	}
	sock := getGlobalSock(socketFD)
	if sock == nil {
		countlog.Warn("event!sut.unknown-after-send",
			"threadID", thread.threadID,
			"socketFD", socketFD)
		return
	}
	sock.afterSend(extraHeaderSentSize, bodySentSize)
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
		if thread.recordingSession.HasResponded() && bytes.HasPrefix(span, InboundRequestPrefix) {
			countlog.Trace("event!sut.recv_from_inbound_found_responded",
				"threadID", thread.threadID,
				"socketFD", socketFD)
			thread.shutdownRecordingSession()
		}
		thread.recordingSession.RecvFromInbound(thread, span, sock.addr)
		replayingSession := replaying.RetrieveTmp(sock.addr)
		if replayingSession != nil {
			nanoOffset := replayingSession.CallFromInbound.GetOccurredAt() - time.Now().UnixNano()
			SetTimeOffset(int(time.Duration(nanoOffset) / time.Second))
			thread.replayingSession = replayingSession
			countlog.Trace("event!sut.received_replaying_session",
				"threadID", thread.threadID,
				"replayingSession", thread.replayingSession,
				"addr", sock.addr)
		}
	} else {
		event = "event!sut.outbound_recv"
		thread.recordingSession.RecvFromOutbound(thread, span, sock.addr, sock.localAddr, int(sock.socketFD))
	}
	countlog.Trace(event,
		"threadID", thread.threadID,
		"socketFD", socketFD,
		"addr", &sock.addr,
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
		"addr", &addr)
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
	setGlobalSock(socketFD, thread.socks[socketFD])
	if envarg.IsReplaying() {
		localAddr, err := bindFDToLocalAddr(int(socketFD))
		if err != nil {
			countlog.Error("event!sut.failed to bind local addr", "err", err)
			return
		}
		thread.socks[socketFD].localAddr = localAddr
		replaying.StoreTmp(*localAddr, thread.replayingSession)
	}
	countlog.Trace("event!sut.connect",
		"threadID", thread.threadID,
		"socketFD", socketFD,
		"addr", &remoteAddr,
		"localAddr", thread.socks[socketFD].localAddr)
}

type SendToFlags int

func (thread *Thread) OnSendTo(socketFD SocketFD, span []byte, flags SendToFlags, addr net.UDPAddr) {
	if addr.String() != "127.127.127.127:127" {
		countlog.Trace("event!sut.sendto",
			"threadID", thread.threadID,
			"socketFD", socketFD,
			"addr", &addr,
			"content", span)
		thread.recordingSession.SendUDP(thread, span, addr)
		thread.replayingSession.SendUDP(thread, span, addr)
		return
	}
	thread.onHelper(socketFD, span, flags, addr)
}

func (thread *Thread) OnOpeningFile(fileName string, flags int) string {
	countlog.Trace("event!sut.opening_file",
		"threadID", thread.threadID,
		"fileName", fileName,
		"flags", flags)
	if thread.replayingSession == nil {
		return ""
	}
	shouldTrace := thread.replayingSession.ShouldTraceFile(fileName)
	fileName = thread.tryMockFile(fileName)
	if shouldTrace {
		fileName = thread.instrumentFile(fileName)
	}
	fileName = thread.tryRedirectFile(fileName)
	shouldTrace = thread.replayingSession.ShouldTraceFile(fileName)
	fileName = thread.tryMockFile(fileName)
	if shouldTrace {
		fileName = thread.instrumentFile(fileName)
	}
	return fileName
}

func (thread *Thread) tryRedirectFile(fileName string) string {
	for redirectFrom, redirectTo := range thread.replayingSession.RedirectDirs {
		if strings.HasPrefix(fileName, redirectFrom) {
			redirectedFileName := strings.Replace(fileName, redirectFrom,
				redirectTo, 1)
			if redirectedFileName != "" {
				return redirectedFileName
			}
		}
	}
	return fileName
}

func (thread *Thread) instrumentFile(fileName string) string {
	instrumentedFileName := trace.InstrumentFile(fileName)
	if instrumentedFileName != "" {
		return instrumentedFileName
	}
	return fileName
}

func (thread *Thread) tryMockFile(fileName string) string {
	if thread.replayingSession.MockFiles != nil {
		mockContent := thread.replayingSession.MockFiles[fileName]
		if mockContent != nil {
			countlog.Trace("event!sut.mock_file",
				"fileName", fileName,
				"content", mockContent)
			mockedFileName := mockFile(mockContent)
			if mockedFileName != "" {
				return mockedFileName
			}
		}
	}
	return fileName
}

func (thread *Thread) OnOpenedFile(fileFD FileFD, fileName string, flags int) {
	countlog.Trace("event!sut.opened_file",
		"threadID", thread.threadID,
		"fileFD", fileFD,
		"fileName", fileName,
		"flags", flags)
	thread.files[fileFD] = &file{
		fileFD:   fileFD,
		fileName: fileName,
		flags:    flags,
	}
}

func (thread *Thread) OnWrite(fileFD FileFD, content []byte) {
	countlog.Trace("event!sut.write",
		"threadID", thread.threadID,
		"fileFD", fileFD)
	file := thread.files[fileFD]
	if file == nil {
		return
	}
	if file.flags&os.O_APPEND == 0 {
		return
	}
	countlog.Trace("event!sut.fileAppend",
		"threadID", thread.threadID,
		"fileFD", fileFD,
		"fileName", file.fileName,
		"content", content)
	thread.recordingSession.AppendFile(thread, content, file.fileName)
	thread.replayingSession.AppendFile(thread, content, file.fileName)
}

func (thread *Thread) OnShutdown() {
	countlog.Trace("event!sut.shutdown_thread",
		"threadID", thread.threadID)
	thread.shutdownRecordingSession()
}

func (thread *Thread) OnAccess() {
	if thread.recordingSession != nil && len(thread.recordingSession.Actions) > 500 {
		countlog.Warn("event!sut.recorded_too_many_actions",
			"threadID", thread.threadID,
			"sessionId", thread.recordingSession.SessionId)
		thread.shutdownRecordingSession()
	}
}

func (thread *Thread) shutdownRecordingSession() {
	if !envarg.IsRecording() {
		return
	}
	countlog.Trace("event!sut.shutdown_recording_session",
		"threadID", thread.threadID,
		"sessionId", thread.recordingSession.SessionId)
	thread.recordingSession.Shutdown(thread)
	thread.socks = map[SocketFD]*socket{} // socks on thread is a temp cache
	thread.recordingSession = recording.NewSession(int32(thread.threadID))
}
