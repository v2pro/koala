package recording

import (
	"time"
	"net"
	"github.com/v2pro/koala/countlog"
	"context"
	"fmt"
	"encoding/json"
)

type Session struct {
	SessionId           string
	CallFromInbound     *CallFromInbound
	ReturnInbound       *ReturnInbound
	Actions             []Action `json:"-"`
	TypelessActions     []json.RawMessage `json:"Actions"`
	currentAppendFiles  map[string]*AppendFile `json:"-"`
	currentCallOutbound *CallOutbound `json:"-"`
}

func NewSession(suffix int32) *Session {
	return &Session{
		SessionId: fmt.Sprintf("%d-%d", time.Now().UnixNano(), suffix),
	}
}

func (session *Session) FileAppend(ctx context.Context, content []byte, fileName string) {
	if session == nil {
		return
	}
	if session.currentAppendFiles == nil {
		session.currentAppendFiles = map[string]*AppendFile{}
	}
	appendFile := session.currentAppendFiles[fileName]
	if appendFile == nil {
		appendFile = &AppendFile{
			action:   newAction("AppendFile"),
			FileName: fileName,
		}
		session.currentAppendFiles[fileName] = appendFile
	}
	appendFile.Content = append(appendFile.Content, content...)
}

func (session *Session) RecvFromInbound(ctx context.Context, span []byte, peer net.TCPAddr) {
	if session == nil {
		return
	}
	if session.CallFromInbound == nil {
		session.CallFromInbound = &CallFromInbound{
			action: newAction("CallFromInbound"),
			Peer:   peer,
		}
	}
	session.CallFromInbound.Request = append(session.CallFromInbound.Request, span...)
}

func (session *Session) SendToInbound(ctx context.Context, span []byte, peer net.TCPAddr) {
	if session == nil {
		return
	}
	if session.ReturnInbound == nil {
		session.ReturnInbound = &ReturnInbound{
			action: newAction("ReturnInbound"),
		}
		session.Actions = append(session.Actions, session.ReturnInbound)
	}
	session.ReturnInbound.Response = append(session.ReturnInbound.Response, span...)
}

func (session *Session) RecvFromOutbound(ctx context.Context, span []byte, peer net.TCPAddr) {
	if session == nil {
		return
	}
	if session.currentCallOutbound == nil {
		session.currentCallOutbound = &CallOutbound{
			action: newAction("CallOutbound"),
			Peer:   peer,
		}
		session.Actions = append(session.Actions, session.currentCallOutbound)
	}
	if session.currentCallOutbound.ResponseTime == 0 {
		session.currentCallOutbound.ResponseTime = time.Now().UnixNano()
	}
	session.currentCallOutbound.Response = append(session.currentCallOutbound.Response, span...)
}

func (session *Session) SendToOutbound(ctx context.Context, span []byte, peer net.TCPAddr) {
	if session == nil {
		return
	}
	if session.currentCallOutbound == nil {
		session.currentCallOutbound = &CallOutbound{
			action: newAction("CallOutbound"),
			Peer:   peer,
		}
		session.Actions = append(session.Actions, session.currentCallOutbound)
	}
	if len(session.currentCallOutbound.Response) > 0 {
		countlog.Trace("event!recording.outbound_talk_recorded",
			"addr", session.currentCallOutbound.Peer,
			"request", session.currentCallOutbound.Request,
			"response", session.currentCallOutbound.Response,
			"ctx", ctx)
		session.Actions = append(session.Actions, session.currentCallOutbound)
		session.currentCallOutbound = &CallOutbound{
			action: newAction("CallOutbound"),
			Peer:   peer,
		}
		session.Actions = append(session.Actions, session.currentCallOutbound)
	}
	session.currentCallOutbound.Request = append(session.currentCallOutbound.Request, span...)
}

func (session *Session) HasResponded() bool {
	if session == nil {
		return false
	}
	if session.ReturnInbound == nil {
		return false
	}
	return true
}

func (session *Session) Shutdown(ctx context.Context) {
	if session == nil {
		return
	}
	for _, recorder := range Recorders {
		recorder.Record(session)
	}
	countlog.Debug("event!recording.session_recorded",
		"ctx", ctx,
		"session", session,
	)
}
