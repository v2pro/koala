package recording

import (
	"time"
	"net"
	"github.com/v2pro/koala/countlog"
	"context"
)

type Session struct {
	InboundTalk         *Talk
	OutboundTalks       []*Talk
	currentOutboundTalk *Talk
}

func (session *Session) InboundRecv(ctx context.Context, span []byte, peer net.TCPAddr) {
	if session == nil {
		return
	}
	if session.InboundTalk == nil {
		session.InboundTalk = &Talk{Peer: peer}
	}
	if session.InboundTalk.RequestTime == 0 {
		session.InboundTalk.RequestTime = time.Now().UnixNano()
	}
	session.InboundTalk.Request = append(session.InboundTalk.Request, span...)
}

func (session *Session) InboundSend(ctx context.Context, span []byte, peer net.TCPAddr) {
	if session == nil {
		return
	}
	if session.InboundTalk == nil {
		session.InboundTalk = &Talk{Peer: peer}
	}
	if session.InboundTalk.ResponseTime == 0 {
		session.InboundTalk.ResponseTime = time.Now().UnixNano()
	}
	session.InboundTalk.Response = append(session.InboundTalk.Response, span...)
}

func (session *Session) OutboundRecv(ctx context.Context, span []byte, peer net.TCPAddr) {
	if session == nil {
		return
	}
	if session.currentOutboundTalk == nil {
		session.currentOutboundTalk = &Talk{Peer: peer}
	}
	if session.currentOutboundTalk.ResponseTime == 0 {
		session.currentOutboundTalk.ResponseTime = time.Now().UnixNano()
	}
	session.currentOutboundTalk.Response = append(session.currentOutboundTalk.Response, span...)
}

func (session *Session) OutboundSend(ctx context.Context, span []byte, peer net.TCPAddr) {
	if session == nil {
		return
	}
	if session.currentOutboundTalk == nil {
		session.currentOutboundTalk = &Talk{Peer: peer}
	}
	if len(session.currentOutboundTalk.Response) > 0 {
		countlog.Trace("event!recording.outbound_talk_recorded",
			"addr", session.currentOutboundTalk.Peer,
			"request", session.currentOutboundTalk.Request,
			"response", session.currentOutboundTalk.Response,
			"ctx", ctx)
		session.OutboundTalks = append(session.OutboundTalks, session.currentOutboundTalk)
		session.currentOutboundTalk = &Talk{Peer: peer}
	}
	if session.currentOutboundTalk.RequestTime == 0 {
		session.currentOutboundTalk.RequestTime = time.Now().UnixNano()
	}
	session.currentOutboundTalk.Request = append(session.currentOutboundTalk.Request, span...)
}

func (session *Session) HasResponded() bool {
	if session == nil {
		return false
	}
	if session.InboundTalk == nil {
		return false
	}
	return len(session.InboundTalk.Response) > 0
}

func (session *Session) Shutdown(ctx context.Context) {
	if session == nil {
		return
	}
	session.OutboundTalks = append(session.OutboundTalks, session.currentOutboundTalk)
	for _, recorder := range Recorders {
		recorder.Record(session)
	}
	countlog.Debug("event!recording.session_recorded",
		"ctx", ctx,
		"session", session,
	)
}
