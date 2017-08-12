package inbound

import (
	"net/http"
	"io/ioutil"
	"github.com/v2pro/koala/countlog"
	"encoding/json"
	"net"
	"time"
	"github.com/v2pro/koala/st"
	"github.com/v2pro/koala/replaying"
)

func Start() {
	go func() {
		http.HandleFunc("/", handleInbound)
		http.ListenAndServe(":9001", nil)
	}()
}

func handleInbound(respWriter http.ResponseWriter, req *http.Request) {
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
	replaying.StoreTmp(*localAddr, &session)
	conn, err := net.DialTCP("tcp", localAddr, remoteAddr)
	if err != nil {
		countlog.Error("failed to connect sut", "err", err)
		return
	}
	_, err = conn.Write(session.InboundTalk.Request)
	if err != nil {
		countlog.Error("failed to write sut", "err", err)
		return
	}
	buf := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(time.Second * 5))
	bytesRead, err := conn.Read(buf)
	if err != nil {
		countlog.Error("failed to read first packet from sut", "err", err)
		return
	}
	response := buf[:bytesRead]
	for {
		conn.SetReadDeadline(time.Now().Add(time.Second))
		bytesRead, err = conn.Read(buf)
		if err != nil {
			break
		}
		response = append(response, buf[:bytesRead]...)
	}
	respWriter.Write(response)
}
