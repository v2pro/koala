package outbound

import (
	"net"
	"github.com/v2pro/plz/countlog"
	"time"
	"github.com/v2pro/koala/replaying"
	"github.com/v2pro/koala/recording"
	"github.com/v2pro/koala/envarg"
	"context"
	"github.com/v2pro/koala/internal"
	"io"
)


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
	protocol := ""
	for i := 0; i < 1024; i++ {
		guessedProtocol, request := readRequest(ctx, conn, buf, i == 0)
		if guessedProtocol != "" {
			protocol = guessedProtocol
		}
		if len(request) == 0 {
			if protocol == "mysql" && request != nil {
				continue
			}
			countlog.Warn("event!outbound.received empty request", "ctx", ctx)
			return
		}
		replayingSession := replaying.RetrieveTmp(*tcpAddr)
		if replayingSession == nil {
			if len(request) == 0 {
				countlog.Warn("event!outbound.read request empty", "ctx", ctx)
				return
			}
			if protocol == "mysql" {
				resp := simulateMysql(ctx, request)
				if resp != nil {
					_, err := conn.Write(resp)
					if err != nil {
						countlog.Error("event!outbound.failed to write back response from outbound",
							"ctx", ctx, "err", err)
						return
					}
					continue
				}
			}
			countlog.Error("event!outbound.outbound can not find replaying session",
				"ctx", ctx,
				"addr", *tcpAddr,
				"content", request)
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
			if protocol == "mysql" {
				resp := simulateMysql(ctx, request)
				if resp != nil {
					_, err := conn.Write(resp)
					if err != nil {
						countlog.Error("event!outbound.failed to write back response from outbound",
							"ctx", ctx, "err", err)
						return
					}
					continue
				}
			}
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

func readRequest(ctx context.Context, conn *net.TCPConn, buf []byte, isFirstPacket bool) (string, []byte) {
	request := []byte{}
	if isFirstPacket {
		conn.SetReadDeadline(time.Now().Add(time.Millisecond * 5))
	} else {
		conn.SetReadDeadline(time.Now().Add(time.Second * 5))
	}
	bytesRead, err := conn.Read(buf)
	if err == io.EOF {
		return "", nil
	}
	protocol := ""
	if err != nil {
		if isFirstPacket {
			countlog.Debug("event!outbound.write_mysql_greeting",
				"ctx", ctx)
			_, err := conn.Write(mysqlGreeting)
			if err != nil {
				countlog.Error("event!outbound.failed to write mysql greeting",
					"ctx", ctx,
					"err", err)
				return "",nil
			}
			protocol = "mysql"
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
	return protocol, request
}
