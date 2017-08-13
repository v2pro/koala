package inbound

import (
	"net/http"
	"io/ioutil"
	"github.com/v2pro/koala/countlog"
	"encoding/json"
	"net"
	"time"
	"github.com/v2pro/koala/replaying"
	"github.com/v2pro/koala/st"
)

func Start() {
	go func() {
		defer func() {
			recovered := recover()
			if recovered != nil {
				countlog.Fatal("panic", "err", recovered)
			}
		}()
		http.HandleFunc("/", handleInbound)
		countlog.Info("inbound-started")
		err := http.ListenAndServe(":9001", nil)
		countlog.Info("inbound-exited", "err", err)
	}()
}

func handleInbound(respWriter http.ResponseWriter, req *http.Request) {
	defer func() {
		recovered := recover()
		if recovered != nil {
			countlog.Fatal("panic", "err", recovered)
		}
	}()
	countlog.Debug("inbound-received", "remoteAddr", req.RemoteAddr)
	reqBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		countlog.Error("failed to read request", "err", err)
		return
	}
	defer req.Body.Close()
	session := st.Session{}
	err = json.Unmarshal(reqBody, &session)
	if err != nil {
		countlog.Error("failed to unmarshal session", "err", err)
		return
	}
	localAddr, remoteAddr, err := replaying.ResolveAddresses(":9000")
	if err != nil {
		countlog.Error("failed to resolve addresses", "err", err)
		return
	}
	replayingSession := replaying.ReplayingSession{
		Session:                       session,
		ReplayedOutboundTalkCollector: make(chan replaying.ReplayedTalk, 4096),
		ReplayedRequestTime:           time.Now().UnixNano(),
	}
	replaying.StoreTmp(*localAddr, &replayingSession)
	conn, err := net.DialTCP("tcp", localAddr, remoteAddr)
	if err != nil {
		countlog.Error("failed to connect sut", "err", err)
		return
	}
	_, err = conn.Write(replayingSession.InboundTalk.Request)
	if err != nil {
		countlog.Error("failed to write sut", "err", err)
		return
	}
	response, err := readResponse(conn)
	if err != nil {
		return
	}
	replayingSession.Finish(response)
	marshaledReplayingSession, err := json.Marshal(replayingSession)
	if err != nil {
		countlog.Error("marshal replaying session failed", "err", err)
		return
	}
	_, err = respWriter.Write(marshaledReplayingSession)
	if err != nil {
		countlog.Error("failed to write response", "err", err)
		return
	}
}

func readResponse(conn *net.TCPConn) ([]byte, error) {
	buf := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(time.Second * 5))
	bytesRead, err := conn.Read(buf)
	if err != nil {
		countlog.Error("failed to read first packet from sut", "err", err)
		return nil, err
	}
	response := buf[:bytesRead]
	for {
		conn.SetReadDeadline(time.Now().Add(time.Millisecond * 100))
		bytesRead, err = conn.Read(buf)
		if err != nil {
			break
		}
		response = append(response, buf[:bytesRead]...)
	}
	return response, nil
}
