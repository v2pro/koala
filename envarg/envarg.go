package envarg

// #include <stdlib.h>
import "C"
import (
	"net"
	"strings"
	"github.com/v2pro/koala/countlog"
	"unsafe"
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
	logFile = getenvFromC("KOALA_LOG_FILE")
	if logFile == "" {
		logFile = "STDOUT"
	}
	initLogLevel()
}
func initLogLevel() {
	logLevelStr := strings.ToUpper(getenvFromC("KOALA_LOG_LEVEL"))
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
	addrStr := getenvFromC("KOALA_INBOUND_ADDR")
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
	addrStr := getenvFromC("KOALA_OUTBOUND_ADDR")
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
	addrStr := getenvFromC("KOALA_SUT_ADDR")
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

// getenvFromC to make getenv work in php-fpm child process
func getenvFromC(key string) string {
	keyc := C.CString(key)
	defer C.free(unsafe.Pointer(keyc))
	v := C.getenv(keyc)
	if uintptr(unsafe.Pointer(v)) != 0 {
		return C.GoString(v)
	}
	return ""
}
