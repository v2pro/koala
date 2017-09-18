package outbound

import (
	"net"
	"github.com/v2pro/koala/countlog"
	"time"
	"github.com/v2pro/koala/replaying"
	"github.com/v2pro/koala/recording"
	"github.com/v2pro/koala/envarg"
	"context"
	"github.com/v2pro/koala/internal"
	"io"
	"sync"
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
		request := readRequest(ctx, conn, buf, i == 0)
		if request == nil {
			return
		}
		replayingSession := replaying.RetrieveTmp(*tcpAddr)
		if replayingSession == nil {
			if len(request) == 0 {
				return
			}
			countlog.Error("event!outbound.outbound can not find replaying session",
				"ctx", ctx,
				"addr", *tcpAddr,
				"content", request)
			return
		}
		if len(request) == 0 {
			countlog.Error("event!outbound.received empty request", "ctx", ctx)
			return
		}
		callOutbound := replaying.NewCallOutbound(*tcpAddr, request)
		var matchedTalk *recording.CallOutbound
		var mark float64
		lastMatchedIndex, mark, matchedTalk = replayingSession.MatchOutboundTalk(ctx, lastMatchedIndex, request)
		if matchedTalk == nil && lastMatchedIndex != 0 {
			lastMatchedIndex, mark, matchedTalk = replayingSession.MatchOutboundTalk(ctx, -1, request)
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
		_, err := conn.Write(matchedTalk.Response)
		if err != nil {
			countlog.Error("event!outbound.failed to write back response from outbound",
				"ctx", ctx, "err", err)
			return
		}
		countlog.Debug("event!outbound.response",
			"ctx", ctx,
			"matchedMark", mark,
			"matchedActionIndex", matchedTalk.ActionIndex,
			"matchedIndex", lastMatchedIndex,
			"matchedRequest", matchedTalk.Request,
			"matchedResponse", matchedTalk.Response,
			"replayingSession", replayingSession)
	}
}

func readRequest(ctx context.Context, conn *net.TCPConn, buf []byte, isFirstPacket bool) []byte {
	request := []byte{}
	if isFirstPacket {
		conn.SetReadDeadline(time.Now().Add(time.Millisecond * 30))
	} else {
		conn.SetReadDeadline(time.Now().Add(time.Second * 5))
	}
	bytesRead, err := conn.Read(buf)
	if err == io.EOF {
		return nil
	}
	if err != nil {
		if isFirstPacket {
			countlog.Debug("event!outbound.write_mysql_greeting",
				"ctx", ctx)
			_, err := conn.Write(mysqlGreeting)
			if err != nil {
				countlog.Error("event!outbound.failed to write mysql greeting",
					"ctx", ctx,
					"err", err)
				return nil
			}
		}
	} else {
		request = append(request, buf[:bytesRead]...)
	}
	for {
		conn.SetReadDeadline(time.Now().Add(time.Millisecond * 2))
		bytesRead, err := conn.Read(buf)
		if err != nil {
			break
		}
		request = append(request, buf[:bytesRead]...)
	}
	countlog.Debug("event!outbound.request",
		"ctx", ctx,
		"content", request)
	return request
}
