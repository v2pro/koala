package gw4go

import (
	"github.com/v2pro/plz/countlog"
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
	setupGoRoutineExitHook()
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
			countlog.Trace("event!gw4go.ignore_connect",
				"threadID", internal.GetCurrentGoRoutineId(),
				"fd", fd)
			return
		}
		origAddr := net.TCPAddr{
			IP:   ipv4Addr.Addr[:],
			Port: ipv4Addr.Port,
		}
		sut.OperateThread(sut.ThreadID(internal.GetCurrentGoRoutineId()), func(thread *sut.Thread) {
			thread.OnConnect(
				sut.SocketFD(fd), origAddr,
			)
		})
		if envarg.IsReplaying() {
			countlog.Debug("event!gw4go.rewrite_connect_target",
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
		if internal.GetCurrentGoRoutineIsKoala() {
			countlog.Trace("event!gw4go.ignore_accept",
				"threadID", internal.GetCurrentGoRoutineId(),
				"serverSocketFD", serverSocketFD,
				"clientSocketFD", clientSocketFD)
			return
		}
		sut.OperateThread(sut.ThreadID(internal.GetCurrentGoRoutineId()), func(thread *sut.Thread) {
			thread.OnAccept(
				sut.SocketFD(serverSocketFD), sut.SocketFD(clientSocketFD), sockaddrToTCP(sa),
			)
		})
	})
}

func sockaddrToTCP(sa syscall.Sockaddr) net.TCPAddr {
	switch sa := sa.(type) {
	case *syscall.SockaddrInet4:
		return net.TCPAddr{IP: sa.Addr[0:], Port: sa.Port}
	case *syscall.SockaddrInet6:
		return net.TCPAddr{IP: sa.Addr[0:], Port: sa.Port, Zone: IP6ZoneToString(int(sa.ZoneId))}
	}
	return net.TCPAddr{}
}

func IP6ZoneToString(zone int) string {
	if zone == 0 {
		return ""
	}
	if ifi, err := net.InterfaceByIndex(zone); err == nil {
		return ifi.Name
	}
	return itod(uint(zone))
}

// Convert i to decimal string.
func itod(i uint) string {
	if i == 0 {
		return "0"
	}

	// Assemble decimal in reverse order.
	var b [32]byte
	bp := len(b)
	for ; i > 0; i /= 10 {
		bp--
		b[bp] = byte(i%10) + '0'
	}

	return string(b[bp:])
}

func setupBindHook() {
	internal.RegisterOnBind(func(fd int, sa syscall.Sockaddr) {
		ipv4Addr, _ := sa.(*syscall.SockaddrInet4)
		if ipv4Addr == nil {
			return
		}
		if internal.GetCurrentGoRoutineIsKoala() {
			countlog.Trace("event!gw4go.ignore_bind",
				"threadID", internal.GetCurrentGoRoutineId(),
				"fd", fd)
			return
		}
		sut.OperateThread(sut.ThreadID(internal.GetCurrentGoRoutineId()), func(thread *sut.Thread) {
			thread.OnBind(
				sut.SocketFD(fd), net.TCPAddr{
					IP:   ipv4Addr.Addr[:],
					Port: ipv4Addr.Port,
				},
			)
		})
	})
}

func setupRecvHook() {
	internal.RegisterOnRecv(func(fd int, network string, raddr net.Addr, span []byte) {
		if internal.GetCurrentGoRoutineIsKoala() {
			countlog.Trace("event!gw4go.ignore_recv",
				"threadID", internal.GetCurrentGoRoutineId(),
				"fd", fd)
			return
		}
		switch network {
		case "udp", "udp4", "udp6":
		default:
			sut.OperateThread(sut.ThreadID(internal.GetCurrentGoRoutineId()), func(thread *sut.Thread) {
				thread.OnRecv(sut.SocketFD(fd), span, 0)
			})
		}
	})
}

func setupSendHook() {
	internal.RegisterOnSend(func(fd int, network string, raddr net.Addr, span []byte) {
		if internal.GetCurrentGoRoutineIsKoala() {
			countlog.Trace("event!gw4go.ignore_send",
				"threadID", internal.GetCurrentGoRoutineId(),
				"fd", fd)
			return
		}
		switch network {
		case "udp", "udp4", "udp6":
			udpAddr := raddr.(*net.UDPAddr)
			sut.OperateThread(sut.ThreadID(internal.GetCurrentGoRoutineId()), func(thread *sut.Thread) {
				thread.OnSendTo(sut.SocketFD(fd), span, 0, *udpAddr)
			})
		default:
			sut.OperateThread(sut.ThreadID(internal.GetCurrentGoRoutineId()), func(thread *sut.Thread) {
				thread.OnSend(sut.SocketFD(fd), span, 0)
			})
		}
	})
}

func setupGoRoutineExitHook() {
	internal.RegisterOnGoRoutineExit(func(goid int64) {
		if internal.GetCurrentGoRoutineIsKoala() {
			countlog.Trace("event!gw4go.ignore_goroutine_exit",
				"threadID", goid)
			return
		}
		sut.OperateThread(sut.ThreadID(goid), func(thread *sut.Thread) {
			thread.OnShutdown()
		})
	})
}
