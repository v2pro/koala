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
	"syscall"
	"time"
)

// InboundRequestPrefix is used to recognize php-fpm FCGI_BEGIN_REQUEST packet.
// fastcgi_finish_request() will send STDOUT first, then recv STDIN (if POST body has not been read before)
// this behavior will trigger session shutdown as we are going to think the recv STDIN
// is the beginning of next request.
// Set InboundRequestPrefix to []byte{1, 1} to only begin new session for FCGI_BEGIN_REQUEST.
// First 0x01 is the version field of fastcgi protocol, second 0x01 is FCGI_BEGIN_REQUEST.
var InboundRequestPrefix = []byte{}
var helperThreadShutdown = "to-koala!thread-shutdown"
var helperCallFunction = "to-koala!call-function"
var helperReturnFunction = "to-koala!return-function"
var helperReadStorage = "to-koala!read-storage"

func (thread *Thread) lookupSocket(socketFD SocketFD) *socket {
	sock := thread.socks[socketFD]
	if sock != nil {
		return sock
	}
	sock = getGlobalSock(socketFD)
	if sock == nil {
		return nil
	}
	remoteAddr, err := syscall.Getpeername(int(socketFD))
	if err != nil {
		countlog.Error("event!failed to get peer name", "err", err, "socketFD", socketFD)
		return nil
	}
	remoteAddr4, _ := remoteAddr.(*syscall.SockaddrInet4)
	// if remote address changed, the fd must be closed and reused
	if remoteAddr4 != nil && (remoteAddr4.Port != sock.addr.Port ||
		remoteAddr4.Addr[0] != sock.addr.IP[0] ||
		remoteAddr4.Addr[1] != sock.addr.IP[1] ||
		remoteAddr4.Addr[2] != sock.addr.IP[2] ||
		remoteAddr4.Addr[3] != sock.addr.IP[3]) {
		sock = &socket{
			socketFD: socketFD,
			isServer: false,
			addr: net.TCPAddr{
				Port: remoteAddr4.Port,
				IP:   net.IP(remoteAddr4.Addr[:]),
			},
			lastAccessedAt: time.Now(),
		}
		setGlobalSock(socketFD, sock)
	}
	remoteAddr6, _ := remoteAddr.(*syscall.SockaddrInet6)
	if remoteAddr6 != nil && (remoteAddr6.Port != sock.addr.Port ||
		remoteAddr6.Addr[0] != sock.addr.IP[0] ||
		remoteAddr6.Addr[1] != sock.addr.IP[1] ||
		remoteAddr6.Addr[2] != sock.addr.IP[2] ||
		remoteAddr6.Addr[3] != sock.addr.IP[3] ||
		remoteAddr6.Addr[4] != sock.addr.IP[4] ||
		remoteAddr6.Addr[5] != sock.addr.IP[5]) {
		sock = &socket{
			socketFD: socketFD,
			isServer: false,
			addr: net.TCPAddr{
				Port: remoteAddr6.Port,
				IP:   net.IP(remoteAddr6.Addr[:]),
			},
			lastAccessedAt: time.Now(),
		}
		setGlobalSock(socketFD, sock)
	}
	thread.socks[socketFD] = sock
	return sock
}

type SendFlags int

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
		thread.recordingSession.RecvFromInbound(thread, span, sock.addr, sock.unixAddr)
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

func (thread *Thread) OnAcceptUnix(serverSocketFD SocketFD, clientSocketFD SocketFD, addr net.UnixAddr) {
	thread.socks[clientSocketFD] = &socket{
		socketFD: clientSocketFD,
		isServer: true,
		unixAddr: addr,
	}
	setGlobalSock(clientSocketFD, thread.socks[clientSocketFD])
	countlog.Debug("event!sut.accept_unix",
		"threadID", thread.threadID,
		"serverSocketFD", serverSocketFD,
		"clientSocketFD", clientSocketFD,
		"unixAddr", addr)
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

func (thread *Thread) OnBindUnix(socketFD SocketFD, addr net.UnixAddr) {
	thread.socks[socketFD] = &socket{
		socketFD: socketFD,
		isServer: false,
		unixAddr: addr,
	}
	countlog.Debug("event!sut.bind",
		"threadID", thread.threadID,
		"socketFD", socketFD,
		"unixAddr", addr)
}

func (thread *Thread) OnConnect(socketFD SocketFD, remoteAddr net.TCPAddr) {
	thread.socks[socketFD] = &socket{
		socketFD: socketFD,
		isServer: false,
		addr:     remoteAddr,
	}
	setGlobalSock(socketFD, thread.socks[socketFD])
	if envarg.IsReplaying() {
		localAddr, err := replaying.BindFDToLocalAddr(int(socketFD))
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

func (thread *Thread) OnConnectUnix(socketFD SocketFD, remoteAddr net.UnixAddr) {
	thread.socks[socketFD] = &socket{
		socketFD: socketFD,
		isServer: false,
		unixAddr: remoteAddr,
	}
	setGlobalSock(socketFD, thread.socks[socketFD])
	//TODO: replaying
	if envarg.IsReplaying() {
		localAddr, err := replaying.BindFDToLocalAddr(int(socketFD))
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
		"unixAddr", thread.socks[socketFD].unixAddr)
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
	helperInfo := span
	countlog.Trace("event!sut.received_helper_info",
		"threadID", thread.threadID,
		"socketFD", socketFD,
		"addr", &addr,
		"content", helperInfo)
	newlinePos := bytes.IndexByte(helperInfo, '\n')
	if newlinePos == -1 {
		return
	}
	helperType := string(helperInfo[:newlinePos])
	body := helperInfo[newlinePos+1:]
	switch helperType {
	case helperThreadShutdown:
		thread.OnShutdown()
	case helperCallFunction:
		thread.replayingSession.CallFunction(thread, body)
	case helperReturnFunction:
		thread.replayingSession.ReturnFunction(thread, body)
	case helperReadStorage:
		thread.recordingSession.ReadStorage(thread, body)
	default:
		countlog.Debug("event!sut.unknown_helper",
			"threadID", thread.threadID, "helperType", helperType)
	}
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
	newSession := recording.NewSession(int32(thread.threadID))
	countlog.Trace("event!sut.shutdown_recording_session",
		"threadID", thread.threadID,
		"sessionId", thread.recordingSession.SessionId,
		"nextSessionId", newSession.SessionId,
		"sessionSummary", thread.recordingSession.Summary())
	thread.recordingSession.NextSessionId = newSession.SessionId
	thread.recordingSession.Shutdown(thread)
	thread.socks = map[SocketFD]*socket{} // socks on thread is a temp cache
	thread.recordingSession = newSession
}
