package outbound

import (
	"net"
	"github.com/v2pro/koala/countlog"
	"time"
	"github.com/v2pro/koala/replaying"
	"io"
	"github.com/v2pro/koala/recording"
	"github.com/v2pro/koala/envarg"
	"context"
	"github.com/v2pro/koala/internal"
)

var mysqlGreeting = []byte{53, 0, 0, 0, 10, 53, 46, 48, 46, 53, 49, 98, 0, 1, 0, 0, 0, 47, 85, 62, 116, 80, 114, 109, 75, 0, 12, 162, 33, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 86, 76, 87, 84, 124, 52, 47, 46, 55, 107, 55, 110, 0}

func Start() {
	go server()
}

func server() {
	internal.SetCurrentGoRoutineIsKoala()
	defer func() {
		recovered := recover()
		if recovered != nil {
			countlog.Fatal("event!outbound.panic", "err", recovered,
				"stacktrace", countlog.ProvideStacktrace)
		}
	}()
	listener, err := net.Listen("tcp", envarg.OutboundAddr().String())
	if err != nil {
		countlog.Error("event!outbound.failed to listen outbound", "err", err)
		return
	}
	countlog.Info("event!outbound.started", "outboundAddr", envarg.OutboundAddr())
	for {
		conn, err := listener.Accept()
		if err != nil {
			countlog.Error("event!outbound.failed to accept outbound", "err", err)
			return
		}
		go handleOutbound(conn.(*net.TCPConn))
	}
}

func handleOutbound(conn *net.TCPConn) {
	internal.SetCurrentGoRoutineIsKoala()
	defer func() {
		recovered := recover()
		if recovered != nil {
			countlog.Fatal("event!outbound.panic", "err", recovered,
				"stacktrace", countlog.ProvideStacktrace)
		}
	}()
	defer conn.Close()
	tcpAddr := conn.RemoteAddr().(*net.TCPAddr)
	countlog.Trace("event!outbound.new_conn",
		"addr", *tcpAddr, )
	buf := make([]byte, 1024)
	lastMatchedIndex := -1
	ctx := context.WithValue(context.Background(), "outboundSrc", tcpAddr.String())
	for i := 0; i < 1024; i++ {
		request := []byte{}
		conn.SetReadDeadline(time.Now().Add(time.Millisecond * 5))
		bytesRead, err := conn.Read(buf)
		if err != nil {
			if i == 0 {
				_, err := conn.Write(mysqlGreeting)
				if err != nil {
					countlog.Error("event!outbound.failed to write mysql greeting",
						"ctx", ctx,
						"err", err)
					return
				}
			} else {
				for {
					conn.SetReadDeadline(time.Now().Add(time.Second * 1))
					bytesRead, err := conn.Read(buf)
					if err == io.EOF {
						return
					}
					if err != nil {
						countlog.Error("event!outbound.outbound wait for follow up timed out",
							"ctx", ctx,
							"err", err)
						return
					}
					request = append(request, buf[:bytesRead]...)
					break
				}
			}
		} else {
			request = append(request, buf[:bytesRead]...)
		}
		for {
			conn.SetReadDeadline(time.Now().Add(time.Millisecond * 5))
			bytesRead, err := conn.Read(buf)
			if err != nil {
				break
			}
			request = append(request, buf[:bytesRead]...)
		}
		replayingSession := replaying.RetrieveTmp(*tcpAddr)
		if replayingSession == nil {
			countlog.Error("event!outbound.outbound can not find replaying session",
				"ctx", ctx,
				"addr", *tcpAddr)
			return
		}
		countlog.Debug("event!outbound.request",
			"ctx", ctx,
			"content", request,
			"replayingSession", replayingSession)
		callOutbound := replaying.NewCallOutbound(*tcpAddr, request)
		var matchedTalk *recording.CallOutbound
		var mark float64
		lastMatchedIndex, mark, matchedTalk = replayingSession.MatchOutboundTalk(ctx, lastMatchedIndex, request)
		if matchedTalk == nil && lastMatchedIndex != 0 {
			lastMatchedIndex, mark, matchedTalk = replayingSession.MatchOutboundTalk(ctx,-1, request)
		}
		if matchedTalk == nil {
			callOutbound.MatchedRequest = nil
			callOutbound.MatchedResponse = nil
			callOutbound.MatchedActionIndex = -1
		} else {
			callOutbound.MatchedRequest = matchedTalk.Request
			callOutbound.MatchedResponse = matchedTalk.Response
			callOutbound.MatchedActionIndex = matchedTalk.ActionIndex
		}
		callOutbound.MatchedMark = mark
		replayingSession.CallOutbound(ctx, callOutbound)
		if matchedTalk == nil {
			countlog.Error("event!outbound.failed to find matching talk", "ctx", ctx)
			return
		}
		_, err = conn.Write(matchedTalk.Response)
		if err != nil {
			countlog.Error("event!outbound.failed to write back response from outbound",
				"ctx", ctx, "err", err)
			return
		}
		countlog.Debug("event!outbound.response",
			"addr", *tcpAddr,
			"matchedMark", mark,
			"matchedActionIndex", matchedTalk.ActionIndex,
			"matchedIndex", lastMatchedIndex,
			"matchedRequest", matchedTalk.Request,
			"matchedResponse", matchedTalk.Response,
			"replayingSession", replayingSession)
	}
}
