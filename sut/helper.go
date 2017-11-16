package sut

import (
	"github.com/v2pro/plz/countlog"
	"bytes"
)

var helperThreadShutdown = "to-koala!thread-shutdown"
var helperCallFunction = "to-koala!call-function"
var helperReturnFunction = "to-koala!return-function"
var helperReadStorage = "to-koala!read-storage"
var helperSetDelegatedFromThreadId = "to-koala!set-delegated-from-thread-id"

func SendToKoala(threadID ThreadID,socketFD SocketFD, span []byte, flags SendToFlags) {
	helperInfo := span
	countlog.Trace("event!sut.send_to_koala",
		"threadID", threadID,
		"socketFD", socketFD,
		"content", helperInfo)
	newlinePos := bytes.IndexByte(helperInfo, '\n')
	if newlinePos == -1 {
		return
	}
	body := helperInfo[newlinePos+1:]
	switch string(helperInfo[:newlinePos]) {
	case helperThreadShutdown:
		operateVirtualThread(ThreadID(socketFD), func(thread *Thread) {
			thread.OnShutdown()
		})
	case helperCallFunction:
		OperateThread(threadID, func(thread *Thread) {
			thread.replayingSession.CallFunction(thread, body)
		})
	case helperReturnFunction:
		OperateThread(threadID, func(thread *Thread) {
			thread.replayingSession.ReturnFunction(thread, body)
		})
	case helperReadStorage:
		OperateThread(threadID, func(thread *Thread) {
			thread.recordingSession.ReadStorage(thread, body)
		})
	case helperSetDelegatedFromThreadId:
		realThreadId := threadID
		virtualThreadId := ThreadID(socketFD)
		mapThreadRelation(realThreadId, virtualThreadId)
	default:
		countlog.Debug("event!sut.unknown_helper",
			"threadID", threadID,
				"helperType", string(helperInfo[:newlinePos]))
	}
}