package sut

import (
	"net"
	"sync"
	"context"
	"github.com/v2pro/koala/replaying"
	"github.com/v2pro/koala/recording"
)

type SocketFD int

type ThreadID int32

type socket struct {
	socketFD SocketFD
	isServer bool
	addr     net.TCPAddr
	localAddr *net.TCPAddr
}

type Thread struct {
	context.Context
	threadID         ThreadID
	socks            map[SocketFD]*socket
	recordingSession *recording.Session
	replayingSession *replaying.ReplayingSession
}

var globalSocks = map[SocketFD]*socket{}
var globalSocksMutex = &sync.Mutex{}
var globalThreads = map[ThreadID]*Thread{}
var globalThreadsMutex = &sync.Mutex{}

func setGlobalSock(socketFD SocketFD, sock *socket) {
	globalSocksMutex.Lock()
	defer globalSocksMutex.Unlock()
	globalSocks[socketFD] = sock
}

func getGlobalSock(socketFD SocketFD) *socket {
	globalSocksMutex.Lock()
	defer globalSocksMutex.Unlock()
	return globalSocks[socketFD]
}

func GetThread(threadID ThreadID) *Thread {
	globalThreadsMutex.Lock()
	defer globalThreadsMutex.Unlock()
	thread := globalThreads[threadID]
	if thread == nil {
		thread = &Thread{
			Context:          context.WithValue(context.Background(), "threadID", threadID),
			threadID:         threadID,
			socks:            map[SocketFD]*socket{},
		}
		if replaying.IsRecording() {
			thread.recordingSession = &recording.Session{}
		}
		globalThreads[threadID] = thread
	}
	return thread
}
