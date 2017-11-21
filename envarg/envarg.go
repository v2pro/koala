package envarg

// #include <stdlib.h>
import "C"
import (
	"net"
	"strings"
	"github.com/v2pro/plz/countlog"
	"unsafe"
)

var inboundAddr *net.TCPAddr
var outboundAddr *net.TCPAddr
var sutAddr *net.TCPAddr
var logFile string
var logLevel = countlog.LevelDebug
var logFormat string

func init() {
	initInboundAddr()
	initOutboundAddr()
	initSutAddr()
	logFile = GetenvFromC("KOALA_LOG_FILE")
	if logFile == "" {
		logFile = "STDOUT"
	}
	initLogLevel()
	logFormat = GetenvFromC("KOALA_LOG_FORMAT")
	if logFormat == "" {
		logFormat = "HumanReadableFormat"
	}
}
func initLogLevel() {
	logLevelStr := strings.ToUpper(GetenvFromC("KOALA_LOG_LEVEL"))
	switch logLevelStr {
	case "TRACE":
		logLevel = countlog.LevelTrace
	case "DEBUG":
		logLevel = countlog.LevelDebug
	case "INFO":
		logLevel = countlog.LevelInfo
	case "WARN":
		logLevel = countlog.LevelWarn
	case "ERROR":
		logLevel = countlog.LevelError
	case "FATAL":
		logLevel = countlog.LevelFatal
	}
}
func initInboundAddr() {
	addrStr := GetenvFromC("KOALA_INBOUND_ADDR")
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
	addrStr := GetenvFromC("KOALA_OUTBOUND_ADDR")
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
	addrStr := GetenvFromC("KOALA_SUT_ADDR")
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

func IsTracing() bool {
	return isTracing
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

func LogFormat() string {
	return logFormat
}

// GetenvFromC to make getenv work in php-fpm child process
func GetenvFromC(key string) string {
	keyc := C.CString(key)
	defer C.free(unsafe.Pointer(keyc))
	v := C.getenv(keyc)
	if uintptr(unsafe.Pointer(v)) != 0 {
		return C.GoString(v)
	}
	return ""
}
