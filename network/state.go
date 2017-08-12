package network

import (
	"net"
	"sync"
)

type SocketFD int

type socket struct {
	socketFD SocketFD
	isServer bool
	addr     net.TCPAddr
}

type ThreadID int32

type Thread struct {
	threadID ThreadID
	socks    map[SocketFD]*socket
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
			threadID: threadID,
			socks:    map[SocketFD]*socket{},
		}
		globalThreads[threadID] = thread
	}
	return thread
}
