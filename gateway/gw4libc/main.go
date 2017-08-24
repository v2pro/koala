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
// #include "allocated_string.h"
// #include "time_hook.h"
// #include "init.h"
import "C"
import (
	"github.com/v2pro/koala/ch"
	"syscall"
	"net"
	"github.com/v2pro/koala/envarg"
	"github.com/v2pro/koala/gateway/gw4go"
)

func init() {
	defer func() {
		recovered := recover()
		if recovered != nil {
			countlog.Fatal("event!gw4libc.init.panic", "err", recovered,
				"stacktrace", countlog.ProvideStacktrace)
		}
	}()
	sut.SetTimeOffset = func(offset int) {
		countlog.Debug("event!main.set_time_offset", "offset", offset)
		C.set_time_offset(C.int(offset))
	}
	gw4go.Start()
	C.go_initialized()
}

//export on_connect
func on_connect(threadID C.pid_t, socketFD C.int, remoteAddr *C.struct_sockaddr_in) {
	defer func() {
		recovered := recover()
		if recovered != nil {
			countlog.Fatal("event!gw4libc.connect.panic", "err", recovered,
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
		countlog.Debug("event!gw4libc.redirect_connect_target",
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
			countlog.Fatal("event!gw4libc.bind.panic", "err", recovered,
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
			countlog.Fatal("event!gw4libc.accept.panic", "err", recovered,
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
			countlog.Fatal("event!gw4libc.send.panic", "err", recovered,
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
			countlog.Fatal("event!gw4libc.recv.panic", "err", recovered,
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
			countlog.Fatal("event!gw4libc.sendto.panic", "err", recovered,
				"stacktrace", countlog.ProvideStacktrace)
		}
	}()
	sut.GetThread(sut.ThreadID(threadID)).
		OnSendTo(sut.SocketFD(socketFD), ch_span_to_bytes(span), sut.SendToFlags(flags), net.UDPAddr{
		IP:   ch.Int2ip(sockaddr_in_sin_addr_get(addr)),
		Port: int(ch.Ntohs(sockaddr_in_sin_port_get(addr))),
	})
}

//export on_fopening_file
func on_fopening_file(threadID C.pid_t,
	filename C.struct_ch_span,
	opentype C.struct_ch_span) C.struct_ch_allocated_string {
	defer func() {
		recovered := recover()
		if recovered != nil {
			countlog.Fatal("event!gw4libc.fopening_file.panic", "err", recovered,
				"stacktrace", countlog.ProvideStacktrace)
		}
	}()
	redirectTo := sut.GetThread(sut.ThreadID(threadID)).
		OnOpeningFile(ch_span_to_string(filename), ch_span_to_open_flags(opentype))
	if redirectTo != "" {
		return C.struct_ch_allocated_string{C.CString(redirectTo)}
	}
	return C.struct_ch_allocated_string{nil}
}

//export on_fopened_file
func on_fopened_file(threadID C.pid_t,
	fileFD C.int,
	filename C.struct_ch_span,
	opentype C.struct_ch_span) {
	defer func() {
		recovered := recover()
		if recovered != nil {
			countlog.Fatal("event!gw4libc.fopened_file.panic", "err", recovered,
				"stacktrace", countlog.ProvideStacktrace)
		}
	}()
	sut.GetThread(sut.ThreadID(threadID)).
		OnOpenedFile(sut.FileFD(fileFD), ch_span_to_string(filename), ch_span_to_open_flags(opentype))
}

//export on_opening_file
func on_opening_file(threadID C.pid_t,
	filename C.struct_ch_span,
	flags C.int, mode C.mode_t) C.struct_ch_allocated_string {
	defer func() {
		recovered := recover()
		if recovered != nil {
			countlog.Fatal("event!gw4libc.opening_file.panic", "err", recovered,
				"stacktrace", countlog.ProvideStacktrace)
		}
	}()
	redirectTo := sut.GetThread(sut.ThreadID(threadID)).
		OnOpeningFile(ch_span_to_string(filename), int(flags))
	if redirectTo != "" {
		return C.struct_ch_allocated_string{C.CString(redirectTo)}
	}
	return C.struct_ch_allocated_string{nil}
}

//export on_opened_file
func on_opened_file(threadID C.pid_t,
	fileFD C.int,
	filename C.struct_ch_span,
	flags C.int, mode C.mode_t) {
	defer func() {
		recovered := recover()
		if recovered != nil {
			countlog.Fatal("event!gw4libc.opened_file.panic", "err", recovered,
				"stacktrace", countlog.ProvideStacktrace)
		}
	}()
	sut.GetThread(sut.ThreadID(threadID)).
		OnOpenedFile(sut.FileFD(fileFD), ch_span_to_string(filename), int(flags))
}

//export on_write
func on_write(threadID C.pid_t,
	fileFD C.int,
	span C.struct_ch_span) {
	defer func() {
		recovered := recover()
		if recovered != nil {
			countlog.Fatal("event!gw4libc.write.panic", "err", recovered,
				"stacktrace", countlog.ProvideStacktrace)
		}
	}()
	sut.GetThread(sut.ThreadID(threadID)).
		OnWrite(sut.FileFD(fileFD), ch_span_to_bytes(span))
}

func main() {
}
