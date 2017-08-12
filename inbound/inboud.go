package inbound

import (
	"net/http"
	"io/ioutil"
	"github.com/v2pro/koala/countlog"
	"encoding/json"
	"github.com/v2pro/koala/sut"
	"fmt"
	"net"
	"time"
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
	session := sut.Session{}
	err = json.Unmarshal(reqBody, &session)
	if err != nil {
		countlog.Error("failed to unmarshal session", "err", err)
		return
	}
	conn, err := net.Dial("tcp", ":9000")
	if err != nil {
		countlog.Error("failed to connect sut", "err", err)
		return
	}
	tcpConn := conn.(*net.TCPConn)
	fmt.Println(tcpConn.LocalAddr())
	_, err = tcpConn.Write(session.InboundTalk.Request)
	if err != nil {
		countlog.Error("failed to write sut", "err", err)
		return
	}
	buf := make([]byte, 1024)
	tcpConn.SetReadDeadline(time.Now().Add(time.Second * 5))
	bytesRead, err := tcpConn.Read(buf)
	if err != nil {
		countlog.Error("failed to read first packet from sut", "err", err)
		return
	}
	response := buf[:bytesRead]
	for {
		tcpConn.SetReadDeadline(time.Now().Add(time.Second))
		bytesRead, err = tcpConn.Read(buf)
		if err != nil {
			break
		}
		response = append(response, buf[:bytesRead]...)
	}
	respWriter.Write(response)
}
