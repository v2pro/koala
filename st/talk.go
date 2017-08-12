package st

import "net"

type Talk struct {
	Peer         net.TCPAddr
	RequestTime  int64
	Request      []byte
	ResponseTime int64
	Response     []byte
}