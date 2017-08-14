package envarg

import (
	"os"
	"net"
)

var isReplaying = false
var inboundAddr *net.TCPAddr
var outboundAddr *net.TCPAddr
var sutAddr *net.TCPAddr

func init() {
	isReplaying = os.Getenv("KOALA_MODE") == "REPLAYING"
	initInboundAddr()
	initOutboundAddr()
	initSutAddr()
}
func initInboundAddr() {
	addrStr := os.Getenv("KOALA_INBOUND_ADDR")
	if addrStr == "" {
		addrStr = ":2514"
	}
	addr, err := net.ResolveTCPAddr("tcp", addrStr)
	if err != nil {
		panic("can not resolve inbound addr: " + err.Error())
	}
	inboundAddr = addr
}

func initOutboundAddr() {
	addrStr := os.Getenv("KOALA_OUTBOUND_ADDR")
	if addrStr == "" {
		addrStr = "127.0.0.1:2516"
	}
	addr, err := net.ResolveTCPAddr("tcp", addrStr)
	if err != nil {
		panic("can not resolve outbound addr: " + err.Error())
	}
	outboundAddr = addr
}

func initSutAddr() {
	addrStr := os.Getenv("KOALA_SUT_ADDR")
	if addrStr == "" {
		addrStr = "127.0.0.1:2515"
	}
	addr, err := net.ResolveTCPAddr("tcp", addrStr)
	if err != nil {
		panic("can not resolve sut addr: " + err.Error())
	}
	sutAddr = addr
}

func IsReplaying() bool {
	return isReplaying
}

func IsRecording() bool {
	return !isReplaying
}

func InboundAddr() *net.TCPAddr {
	return inboundAddr
}

func SutAddr() *net.TCPAddr {
	return sutAddr
}

func OutboundAddr() *net.TCPAddr {
	return outboundAddr
}
