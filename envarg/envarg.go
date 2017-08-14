package envarg

import (
	"os"
	"net"
	"strings"
	"github.com/v2pro/koala/countlog"
)

var inboundAddr *net.TCPAddr
var outboundAddr *net.TCPAddr
var sutAddr *net.TCPAddr
var logFile string
var logLevel = countlog.LEVEL_DEBUG

func init() {
	initInboundAddr()
	initOutboundAddr()
	initSutAddr()
	logFile = os.Getenv("KOALA_LOG_FILE")
	if logFile == "" {
		logFile = "STDOUT"
	}
	initLogLevel()
}
func initLogLevel() {
	logLevelStr := strings.ToUpper(os.Getenv("KOALA_LOG_LEVEL"))
	switch logLevelStr {
	case "TRACE":
		logLevel = countlog.LEVEL_TRACE
	case "DEBUG":
		logLevel = countlog.LEVEL_DEBUG
	case "INFO":
		logLevel = countlog.LEVEL_INFO
	case "WARN":
		logLevel = countlog.LEVEL_WARN
	case "ERROR":
		logLevel = countlog.LEVEL_ERROR
	case "FATAL":
		logLevel = countlog.LEVEL_FATAL
	}
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
	return isRecording
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

func LogFile() string {
	return logFile
}

func LogLevel() int {
	return logLevel
}
