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
	defer func() {
		recovered := recover()
		if recovered != nil {
			countlog.Fatal("panic", "err", recovered)
		}
	}()
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
	defer func() {
		recovered := recover()
		if recovered != nil {
			countlog.Fatal("panic", "err", recovered)
		}
	}()
	for {
		request := []byte{}
		buf := make([]byte, 1024)
		for {
			conn.SetReadDeadline(time.Now().Add(time.Second * 5))
			bytesRead, err := conn.Read(buf)
			if err != nil {
				continue
			}
			request = append(request, buf[:bytesRead]...)
			break
		}
		for {
			conn.SetReadDeadline(time.Now().Add(time.Second))
			bytesRead, err := conn.Read(buf)
			if err != nil {
				break
			}
			request = append(request, buf[:bytesRead]...)
		}
		tcpAddr := conn.RemoteAddr().(*net.TCPAddr)
		replayingSession := replaying.RetrieveTmp(*tcpAddr)
		if replayingSession == nil {
			countlog.Error("outbound can not find replaying session", "addr", *tcpAddr)
			return
		}
		countlog.Debug("outbound-request",
			"addr", *tcpAddr,
			"content", request,
			"replayingSession", replayingSession)
		replayedTalk := replaying.ReplayedTalk{
			ReplayedRequest: request,
			ReplayedRequestTime: time.Now().UnixNano(),
		}
		matchedTalk := replayingSession.MatchOutboundTalk(request)
		if matchedTalk == nil {
			countlog.Error("failed to find matching talk", "addr", *tcpAddr)
			return
		}
		replayedTalk.MatchedTalk = matchedTalk
		replayedTalk.ReplayedResponseTime = time.Now().UnixNano()
		countlog.Debug("outbound-response",
			"addr", *tcpAddr,
			"content", matchedTalk.Response,
			"replayingSession", replayingSession)
		_, err := conn.Write(matchedTalk.Response)
		if err != nil {
			countlog.Error("failed to write back response from outbound", "addr", *tcpAddr)
			return
		}
		select {
		case replayingSession.ReplayedOutboundTalkCollector <- replayedTalk:
		default:
			countlog.Error("ReplayedOutboundTalkCollector is full")
		}
	}
}
