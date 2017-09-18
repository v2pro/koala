package sut

import (
	"net"
	"sync"
	"context"
	"github.com/v2pro/koala/replaying"
	"github.com/v2pro/koala/recording"
	"github.com/v2pro/koala/envarg"
	"time"
	"github.com/v2pro/koala/countlog"
)

type SocketFD int

type FileFD int

type ThreadID int32

type socket struct {
	socketFD       SocketFD
	isServer       bool
	addr           net.TCPAddr
	localAddr      *net.TCPAddr
	lastAccessedAt time.Time
}

type file struct {
	fileFD   FileFD
	fileName string
	flags    int
}

type Thread struct {
	context.Context
	threadID         ThreadID
	socks            map[SocketFD]*socket
	files            map[FileFD]*file
	recordingSession *recording.Session
	replayingSession *replaying.ReplayingSession
	lastAccessedAt   time.Time
}

var globalSocks = map[SocketFD]*socket{}
var globalSocksMutex = &sync.Mutex{}
var globalThreads = map[ThreadID]*Thread{}
var globalThreadsMutex = &sync.Mutex{}

func init() {
	go gcStatesInBackground()
}

func setGlobalSock(socketFD SocketFD, sock *socket) {
	globalSocksMutex.Lock()
	defer globalSocksMutex.Unlock()
	sock.lastAccessedAt = time.Now()
	globalSocks[socketFD] = sock
}

func getGlobalSock(socketFD SocketFD) *socket {
	globalSocksMutex.Lock()
	defer globalSocksMutex.Unlock()
	globalSocks[socketFD].lastAccessedAt = time.Now()
	return globalSocks[socketFD]
}

func GetThread(threadID ThreadID) *Thread {
	globalThreadsMutex.Lock()
	defer globalThreadsMutex.Unlock()
	thread := globalThreads[threadID]
	if thread == nil {
		thread = &Thread{
			Context:        context.WithValue(context.Background(), "threadID", threadID),
			threadID:       threadID,
			socks:          map[SocketFD]*socket{},
			files:          map[FileFD]*file{},
			lastAccessedAt: time.Now(),
		}
		if envarg.IsRecording() {
			thread.recordingSession = recording.NewSession(int32(threadID))
		}
		globalThreads[threadID] = thread
	}
	thread.OnAccess()
	thread.lastAccessedAt = time.Now()
	return thread
}

func gcStatesInBackground() {
	defer func() {
		recovered := recover()
		if recovered != nil {
			countlog.Fatal("event!sut.gc_states_in_background.panic", "err", recovered,
				"stacktrace", countlog.ProvideStacktrace)
		}
	}()
	for {
		time.Sleep(time.Second * 10)
		gcStatesOneRound()
	}
}

func gcStatesOneRound() {
	defer func() {
		recovered := recover()
		if recovered != nil {
			countlog.Fatal("event!sut.gc_states_one_round.panic", "err", recovered,
				"stacktrace", countlog.ProvideStacktrace)
		}
	}()
	expiredSocksCount := gcGlobalSocks()
	expiredThreadsCount := gcGlobalThreads()
	countlog.Trace("event!sut.gc_global_states",
		"expiredSocksCount", expiredSocksCount,
		"expiredThreadsCount", expiredThreadsCount)
}

func gcGlobalSocks() int {
	globalSocksMutex.Lock()
	defer globalSocksMutex.Unlock()
	now := time.Now()
	newMap := map[SocketFD]*socket{}
	expiredSocksCount := 0
	for fd, sock := range globalSocks {
		if now.Sub(sock.lastAccessedAt) < time.Second*5 {
			newMap[fd] = sock
		} else {
			expiredSocksCount++
		}
	}
	globalSocks = newMap
	return expiredSocksCount
}

func gcGlobalThreads() int {
	globalThreadsMutex.Lock()
	defer globalThreadsMutex.Unlock()
	now := time.Now()
	newMap := map[ThreadID]*Thread{}
	expiredThreadsCount := 0
	for threadId, thread := range globalThreads {
		if now.Sub(thread.lastAccessedAt) < time.Second*5 {
			newMap[threadId] = thread
		} else {
			expiredThreadsCount++
		}
	}
	globalThreads = newMap
	return expiredThreadsCount
}
