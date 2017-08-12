package outbound

import (
	"net"
	"github.com/v2pro/koala/countlog"
	"time"
	"github.com/v2pro/koala/replaying"
)

func Start() {
	go server()
}

func server() {
	listener, err := net.Listen("tcp", ":9002")
	if err != nil {
		countlog.Error("failed to listen outbound", "err", err)
		return
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			countlog.Error("failed to accept outbound", "err", err)
			return
		}
		go handleOutbound(conn.(*net.TCPConn))
	}
}

func handleOutbound(conn *net.TCPConn) {
	tcpAddr := conn.RemoteAddr().(*net.TCPAddr)
	replayingSession := replaying.RetrieveTmp(*tcpAddr)
	if replayingSession == nil {
		countlog.Error("outbound can not find replaying session", "addr", *tcpAddr)
		return
	}
	request := []byte{}
	buf := make([]byte, 1024)
	for {
		conn.SetReadDeadline(time.Now().Add(time.Second))
		bytesRead, err := conn.Read(buf)
		if err != nil {
			break
		}
		request = append(request, buf[:bytesRead]...)
	}
	countlog.Debug("outbound received",
		"addr", *tcpAddr,
		"content", request,
		"replayingSession", replayingSession)
	matchedTalk := replayingSession.MatchOutboundTalk(request)
	if matchedTalk == nil {
		countlog.Error("failed to find matching talk", "addr", *tcpAddr)
		return
	}
	_, err := conn.Write(matchedTalk.Response)
	if err != nil {
		countlog.Error("failed to write back response from outbound", "addr", *tcpAddr)
		return
	}
}
