package gw4go

import (
	_ "github.com/v2pro/koala/extension"
	"github.com/v2pro/koala/countlog"
	"github.com/v2pro/koala/envarg"
	"github.com/v2pro/koala/inbound"
	"github.com/v2pro/koala/outbound"
	"syscall"
	"github.com/v2pro/koala/sut"
	"net"
	"github.com/v2pro/koala/internal"
)

func Start() {
	setupBindHook()
	setupAcceptHook()
	setupRecvHook()
	setupSendHook()
	setupConnectHook()
	startLogging()
	if envarg.IsReplaying() {
		inbound.Start()
		outbound.Start()
		mode := "replaying"
		if envarg.IsRecording() {
			mode += " & recording"
		}
		countlog.Info("event!main.koala_started",
			"mode", mode)
	} else {
		countlog.Info("event!main.koala_started",
			"mode", "recording")
	}
}
func setupConnectHook() {
	internal.RegisterOnConnect(func(fd int, sa syscall.Sockaddr) {
		ipv4Addr, _ := sa.(*syscall.SockaddrInet4)
		if ipv4Addr == nil {
			return
		}
		if internal.GetCurrentGoRoutineIsKoala() {
			countlog.Trace("event!internal.ignore_connect",
				"threadID", internal.GetCurrentGoRoutineId(),
				"fd", fd)
			return
		}
		origAddr := net.TCPAddr{
			IP:   ipv4Addr.Addr[:],
			Port: ipv4Addr.Port,
		}
		sut.GetThread(sut.ThreadID(internal.GetCurrentGoRoutineId())).OnConnect(
			sut.SocketFD(fd), origAddr,
		)
		if envarg.IsReplaying() {
			countlog.Debug("event!internal.rewrite_connect_target",
				"origAddr", origAddr,
				"redirectTo", envarg.OutboundAddr())
			for i := 0; i < 4; i++ {
				ipv4Addr.Addr[i] = envarg.OutboundAddr().IP[i]
			}
			ipv4Addr.Port = envarg.OutboundAddr().Port
		}
	})
}

func setupAcceptHook() {
	internal.RegisterOnAccept(func(serverSocketFD int, clientSocketFD int, sa syscall.Sockaddr) {
		ipv4Addr, _ := sa.(*syscall.SockaddrInet4)
		if ipv4Addr == nil {
			return
		}
		if internal.GetCurrentGoRoutineIsKoala() {
			countlog.Trace("event!internal.ignore_accept",
				"threadID", internal.GetCurrentGoRoutineId(),
				"serverSocketFD", serverSocketFD,
				"clientSocketFD", clientSocketFD)
			return
		}
		sut.GetThread(sut.ThreadID(internal.GetCurrentGoRoutineId())).OnAccept(
			sut.SocketFD(serverSocketFD), sut.SocketFD(clientSocketFD), net.TCPAddr{
				IP:   ipv4Addr.Addr[:],
				Port: ipv4Addr.Port,
			},
		)
	})
}

func setupBindHook() {
	internal.RegisterOnBind(func(fd int, sa syscall.Sockaddr) {
		ipv4Addr, _ := sa.(*syscall.SockaddrInet4)
		if ipv4Addr == nil {
			return
		}
		if internal.GetCurrentGoRoutineIsKoala() {
			countlog.Trace("event!internal.ignore_bind",
				"threadID", internal.GetCurrentGoRoutineId(),
				"fd", fd)
			return
		}
		sut.GetThread(sut.ThreadID(internal.GetCurrentGoRoutineId())).OnBind(
			sut.SocketFD(fd), net.TCPAddr{
				IP:   ipv4Addr.Addr[:],
				Port: ipv4Addr.Port,
			},
		)
	})
}

func setupRecvHook() {
	internal.RegisterOnRecv(func(fd int, span []byte) {
		if internal.GetCurrentGoRoutineIsKoala() {
			countlog.Trace("event!internal.ignore_recv",
				"threadID", internal.GetCurrentGoRoutineId(),
				"fd", fd)
			return
		}
		sut.GetThread(sut.ThreadID(internal.GetCurrentGoRoutineId())).OnRecv(
			sut.SocketFD(fd), span, 0)
	})
}

func setupSendHook() {
	internal.RegisterOnSend(func(fd int, span []byte) {
		if internal.GetCurrentGoRoutineIsKoala() {
			countlog.Trace("event!internal.ignore_send",
				"threadID", internal.GetCurrentGoRoutineId(),
				"fd", fd)
			return
		}
		sut.GetThread(sut.ThreadID(internal.GetCurrentGoRoutineId())).OnSend(
			sut.SocketFD(fd), span, 0)
	})
}

func startLogging() {
	if len(countlog.LogWriters) != 0 {
		// extension already setup alternative log writers
		return
	}
	logWriter := countlog.NewAsyncLogWriter(envarg.LogLevel(), countlog.NewFileLogOutput(envarg.LogFile()))
	logWriter.LogFormatter = &countlog.HumanReadableFormat{
		ContextPropertyNames: []string{"threadID", "outboundSrc"},
		StringLengthCap:      512,
	}
	logWriter.EventWhitelist["event!replaying.talks_scored"] = true
	logWriter.Start()
	countlog.LogWriters = append(countlog.LogWriters, logWriter)
}
