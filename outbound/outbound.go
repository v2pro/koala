package outbound

import (
	"net"
	"github.com/v2pro/koala/countlog"
	"time"
	"github.com/v2pro/koala/replaying"
	"io"
)

var mysqlGreeting = []byte{53, 0, 0, 0, 10, 53, 46, 48, 46, 53, 49, 98, 0, 1, 0, 0, 0, 47, 85, 62, 116, 80, 114, 109, 75, 0, 12, 162, 33, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 86, 76, 87, 84, 124, 52, 47, 46, 55, 107, 55, 110, 0}


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
	listener, err := net.Listen("tcp", "127.0.0.1:9002")
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
	defer conn.Close()
	tcpAddr := conn.RemoteAddr().(*net.TCPAddr)
	countlog.Debug("outbound-new-conn",
		"addr", *tcpAddr, )
	buf := make([]byte, 1024)
	for i := 0; i < 1024; i++ {
		request := []byte{}
		conn.SetReadDeadline(time.Now().Add(time.Millisecond * 1))
		bytesRead, err := conn.Read(buf)
		if err != nil {
			if i == 0 {
				_, err := conn.Write(mysqlGreeting)
				if err != nil {
					countlog.Error("failed to write mysql greeting", "addr", *tcpAddr, "err", err)
					return
				}
			} else {
				for {
					conn.SetReadDeadline(time.Now().Add(time.Second * 5))
					bytesRead, err := conn.Read(buf)
					if err == io.EOF {
						return
					}
					if err != nil {
						countlog.Error("outbound wait for follow up timed out", "err", err)
						continue
					}
					request = append(request, buf[:bytesRead]...)
					break
				}
			}
		} else {
			request = append(request, buf[:bytesRead]...)
		}
		for {
			conn.SetReadDeadline(time.Now().Add(time.Millisecond * 1))
			bytesRead, err := conn.Read(buf)
			if err != nil {
				break
			}
			request = append(request, buf[:bytesRead]...)
		}
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
			ReplayedRequest:     request,
			ReplayedRequestTime: time.Now().UnixNano(),
		}
		matchedTalk := replayingSession.MatchOutboundTalk(request)
		if matchedTalk == nil {
			countlog.Error("failed to find matching talk", "addr", *tcpAddr)
			return
		}
		replayedTalk.MatchedTalk = matchedTalk
		replayedTalk.ReplayedResponseTime = time.Now().UnixNano()
		size, err := conn.Write(matchedTalk.Response)
		if err != nil {
			countlog.Error("failed to write back response from outbound", "addr", *tcpAddr, "err", err)
			return
		}
		countlog.Debug("outbound-response",
			"addr", *tcpAddr,
			"matchedRequest", matchedTalk.Request,
			"matchedResponse", matchedTalk.Response,
			"replayingSession", replayingSession,
			"size", size)
		select {
		case replayingSession.ReplayedOutboundTalkCollector <- replayedTalk:
		default:
			countlog.Error("ReplayedOutboundTalkCollector is full")
		}
	}
}
