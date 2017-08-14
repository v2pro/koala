package main

import (
	"github.com/v2pro/koala/sut"
	"github.com/v2pro/koala/countlog"
)

// #cgo LDFLAGS: -ldl -lm -lrt
// #include <stddef.h>
// #include <netinet/in.h>
// #include <sys/types.h>
// #include <sys/socket.h>
// #include "span.h"
// #include "network_hook.h"
// #include "time_hook.h"
import "C"
import (
	"github.com/v2pro/koala/ch"
	"syscall"
	"net"
	"github.com/v2pro/koala/inbound"
	"github.com/v2pro/koala/outbound"
	"github.com/v2pro/koala/envarg"
)

func init() {
	defer func() {
		recovered := recover()
		if recovered != nil {
			countlog.Fatal("event!main.panic", "err", recovered,
				"stacktrace", countlog.ProvideStacktrace)
		}
	}()
	startLogging()
	C.network_hook_init()
	C.time_hook_init()
	sut.SetTimeOffset = func(offset int) {
		countlog.Debug("event!main.set_time_offset", "offset", offset)
		C.set_time_offset(C.int(offset))
	}
	if envarg.IsReplaying() {
		inbound.Start()
		outbound.Start()
		countlog.Info("event!main.koala_started",
			"mode", "replaying",
			"inboundAddr", envarg.InboundAddr(),
			"sutAddr", envarg.SutAddr(),
			"outboundAddr", envarg.OutboundAddr())
	} else {
		countlog.Info("event!main.koala_started",
			"mode", "recording")
	}
}

func startLogging() {
	logWriter := countlog.NewFileLogWriter(envarg.LogLevel(), envarg.LogFile())
	logWriter.LogFormatter = &countlog.HumanReadableFormat{
		ContextPropertyNames: []string{"threadID", "outboundSrc"},
		StringLengthCap:      512,
	}
	logWriter.EventWhitelist["event!replaying.talks_scored"] = true
	logWriter.Start()
}

//export on_connect
func on_connect(threadID C.pid_t, socketFD C.int, remoteAddr *C.struct_sockaddr_in) {
	defer func() {
		recovered := recover()
		if recovered != nil {
			countlog.Fatal("event!connect.panic", "err", recovered,
				"stacktrace", countlog.ProvideStacktrace)
		}
	}()
	origAddr := net.TCPAddr{
		IP:   ch.Int2ip(sockaddr_in_sin_addr_get(remoteAddr)),
		Port: int(ch.Ntohs(sockaddr_in_sin_port_get(remoteAddr))),
	}
	if origAddr.String() == "127.0.0.1:18500" {
		return
	}
	sut.GetThread(sut.ThreadID(threadID)).
		OnConnect(sut.SocketFD(socketFD), origAddr)
	if envarg.IsReplaying() {
		countlog.Debug("event!main.rewrite_connect_target",
			"origAddr", origAddr,
			"redirectTo", envarg.OutboundAddr())
		sockaddr_in_sin_addr_set(remoteAddr, ch.Ip2int(envarg.OutboundAddr().IP))
		sockaddr_in_sin_port_set(remoteAddr, ch.Htons(uint16(envarg.OutboundAddr().Port)))
	}
}

//export on_bind
func on_bind(threadID C.pid_t, socketFD C.int, addr *C.struct_sockaddr_in) {
	defer func() {
		recovered := recover()
		if recovered != nil {
			countlog.Fatal("event!bind.panic", "err", recovered,
				"stacktrace", countlog.ProvideStacktrace)
		}
	}()
	sut.GetThread(sut.ThreadID(threadID)).
		OnBind(sut.SocketFD(socketFD), net.TCPAddr{
		IP:   ch.Int2ip(sockaddr_in_sin_addr_get(addr)),
		Port: int(ch.Ntohs(sockaddr_in_sin_port_get(addr))),
	})
}

//export on_accept
func on_accept(threadID C.pid_t, serverSocketFD C.int, clientSocketFD C.int, addr *C.struct_sockaddr_in) {
	defer func() {
		recovered := recover()
		if recovered != nil {
			countlog.Fatal("event!accept.panic", "err", recovered,
				"stacktrace", countlog.ProvideStacktrace)
		}
	}()
	if sockaddr_in_sin_family_get(addr) != syscall.AF_INET {
		panic("expect ipv4 addr")
	}
	sut.GetThread(sut.ThreadID(threadID)).
		OnAccept(sut.SocketFD(serverSocketFD), sut.SocketFD(clientSocketFD), net.TCPAddr{
		IP:   ch.Int2ip(sockaddr_in_sin_addr_get(addr)),
		Port: int(ch.Ntohs(sockaddr_in_sin_port_get(addr))),
	})
}

//export on_send
func on_send(threadID C.pid_t, socketFD C.int, span C.struct_ch_span, flags C.int) {
	defer func() {
		recovered := recover()
		if recovered != nil {
			countlog.Fatal("event!send.panic", "err", recovered,
				"stacktrace", countlog.ProvideStacktrace)
		}
	}()
	sut.GetThread(sut.ThreadID(threadID)).
		OnSend(sut.SocketFD(socketFD), ch_span_to_bytes(span), sut.SendFlags(flags))
}

//export on_recv
func on_recv(threadID C.pid_t, socketFD C.int, span C.struct_ch_span, flags C.int) {
	defer func() {
		recovered := recover()
		if recovered != nil {
			countlog.Fatal("event!recv.panic", "err", recovered,
				"stacktrace", countlog.ProvideStacktrace)
		}
	}()
	sut.GetThread(sut.ThreadID(threadID)).
		OnRecv(sut.SocketFD(socketFD), ch_span_to_bytes(span), sut.RecvFlags(flags))
}

//export on_sendto
func on_sendto(threadID C.pid_t, socketFD C.int, span C.struct_ch_span, flags C.int, addr *C.struct_sockaddr_in) {
	defer func() {
		recovered := recover()
		if recovered != nil {
			countlog.Fatal("event!sendto.panic", "err", recovered,
				"stacktrace", countlog.ProvideStacktrace)
		}
	}()
	sut.GetThread(sut.ThreadID(threadID)).
		OnSendTo(sut.SocketFD(socketFD), ch_span_to_bytes(span), sut.SendToFlags(flags), net.TCPAddr{
		IP:   ch.Int2ip(sockaddr_in_sin_addr_get(addr)),
		Port: int(ch.Ntohs(sockaddr_in_sin_port_get(addr))),
	})
}

func main() {
}
