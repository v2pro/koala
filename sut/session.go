package sut

import (
	"time"
	"net"
	"github.com/v2pro/koala/countlog"
)

type Session struct {
	InboundTalk         *Talk
	OutboundTalks       []*Talk
	currentOutboundTalk *Talk
}

type Talk struct {
	Peer         net.TCPAddr
	RequestTime  int64
	Request      []byte
	ResponseTime int64
	Response     []byte
}

func (session *Session) InboundRecv(span []byte, peer net.TCPAddr) {
	if session.InboundTalk == nil {
		session.InboundTalk = &Talk{Peer: peer}
	}
	if session.InboundTalk.RequestTime == 0 {
		session.InboundTalk.RequestTime = time.Now().UnixNano()
	}
	session.InboundTalk.Request = append(session.InboundTalk.Request, span...)
}

func (session *Session) InboundSend(span []byte, peer net.TCPAddr) {
	if session.InboundTalk == nil {
		session.InboundTalk = &Talk{Peer: peer}
	}
	if session.InboundTalk.ResponseTime == 0 {
		session.InboundTalk.ResponseTime = time.Now().UnixNano()
	}
	session.InboundTalk.Response = append(session.InboundTalk.Response, span...)
}

func (session *Session) OutboundRecv(threadID ThreadID, span []byte, peer net.TCPAddr) {
	if session.currentOutboundTalk == nil {
		session.currentOutboundTalk = &Talk{Peer: peer}
	}
	if session.currentOutboundTalk.ResponseTime == 0 {
		session.currentOutboundTalk.ResponseTime = time.Now().UnixNano()
	}
	session.currentOutboundTalk.Response = append(session.currentOutboundTalk.Response, span...)
}

func (session *Session) OutboundSend(threadID ThreadID, span []byte, peer net.TCPAddr) {
	if session.currentOutboundTalk == nil {
		session.currentOutboundTalk = &Talk{Peer: peer}
	}
	if len(session.currentOutboundTalk.Response) > 0 {
		countlog.Debug("outbound-talk",
			"addr", session.currentOutboundTalk.Peer,
			"request", session.currentOutboundTalk.Request,
			"response", session.currentOutboundTalk.Response)
		session.OutboundTalks = append(session.OutboundTalks, session.currentOutboundTalk)
		session.currentOutboundTalk = &Talk{Peer: peer}
	}
	if session.currentOutboundTalk.RequestTime == 0 {
		session.currentOutboundTalk.RequestTime = time.Now().UnixNano()
	}
	session.currentOutboundTalk.Request = append(session.currentOutboundTalk.Request, span...)
}
