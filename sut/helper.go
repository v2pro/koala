package sut

import (
	"github.com/v2pro/plz/countlog"
	"net"
	"bytes"
)

var helperThreadShutdown = "to-koala!thread-shutdown"
var helperCallFunction = "to-koala!call-function"
var helperReturnFunction = "to-koala!return-function"
var helperReadStorage = "to-koala!read-storage"

func (thread *Thread) onHelper(socketFD SocketFD, span []byte, flags SendToFlags, addr net.UDPAddr) {
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