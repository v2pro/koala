package recording

import (
	"context"
	"fmt"
	"github.com/v2pro/plz/countlog"
	"net"
	"time"
)

type Session struct {
	SessionId           string
	CallFromInbound     *CallFromInbound
	ReturnInbound       *ReturnInbound
	Actions             []Action
	currentAppendFiles  map[string]*AppendFile `json:"-"`
	currentCallOutbound *CallOutbound          `json:"-"`
}

func NewSession(suffix int32) *Session {
	return &Session{
		SessionId: fmt.Sprintf("%d-%d", time.Now().UnixNano(), suffix),
	}
}

func (session *Session) newAction(actionType string) action {
	occurredAt := time.Now().UnixNano()
	return action{
		ActionIndex: len(session.Actions),
		OccurredAt:  occurredAt,
		ActionType:  actionType,
	}
}

func (session *Session) AppendFile(ctx context.Context, content []byte, fileName string) {
	if session == nil {
		return
	}
	if session.currentAppendFiles == nil {
		session.currentAppendFiles = map[string]*AppendFile{}
	}
	appendFile := session.currentAppendFiles[fileName]
	if appendFile == nil {
		appendFile = &AppendFile{
			action:   session.newAction("AppendFile"),
			FileName: fileName,
		}
		session.currentAppendFiles[fileName] = appendFile
		session.addAction(appendFile)
	}
	appendFile.Content = append(appendFile.Content, content...)
}

func (session *Session) ReadStorage(ctx context.Context, span []byte) {
	if session == nil {
		return
	}
	session.addAction(&ReadStorage{
		action:  session.newAction("ReadStorage"),
		Content: span,
	})
}

func (session *Session) RecvFromInbound(ctx context.Context, span []byte, peer net.TCPAddr) {
	if session == nil {
		return
	}
	if session.CallFromInbound == nil {
		session.CallFromInbound = &CallFromInbound{
			action: session.newAction("CallFromInbound"),
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
			action: session.newAction("ReturnInbound"),
		}
		session.addAction(session.ReturnInbound)
	}
	session.ReturnInbound.Response = append(session.ReturnInbound.Response, span...)
}

func (session *Session) RecvFromOutbound(ctx context.Context, span []byte, peer net.TCPAddr, local *net.TCPAddr, socketFD int) {
	if session == nil {
		return
	}
	if session.currentCallOutbound == nil {
		session.currentCallOutbound = &CallOutbound{
			action:   session.newAction("CallOutbound"),
			Peer:     peer,
			Local:    local,
			SocketFD: socketFD,
		}
		session.addAction(session.currentCallOutbound)
	}
	if (session.currentCallOutbound.Peer.String() != peer.String()) ||
		(session.currentCallOutbound.SocketFD != socketFD) {
		session.currentCallOutbound = &CallOutbound{
			action:   session.newAction("CallOutbound"),
			Peer:     peer,
			Local:    local,
			SocketFD: socketFD,
		}
		session.addAction(session.currentCallOutbound)
	}
	if session.currentCallOutbound.ResponseTime == 0 {
		session.currentCallOutbound.ResponseTime = time.Now().UnixNano()
	}
	session.currentCallOutbound.Response = append(session.currentCallOutbound.Response, span...)
}

func (session *Session) SendToOutbound(ctx context.Context, span []byte, peer net.TCPAddr, local *net.TCPAddr, socketFD int) {
	if session == nil {
		return
	}
	if (session.currentCallOutbound == nil) ||
		(session.currentCallOutbound.Peer.String() != peer.String()) ||
		(session.currentCallOutbound.SocketFD != socketFD) ||
		(len(session.currentCallOutbound.Response) > 0) {
		session.currentCallOutbound = &CallOutbound{
			action:   session.newAction("CallOutbound"),
			Peer:     peer,
			Local:    local,
			SocketFD: socketFD,
		}
		session.addAction(session.currentCallOutbound)
	} else if session.currentCallOutbound != nil && session.currentCallOutbound.ResponseTime > 0 {
		// last request get a bad response, e.g., timeout
		session.currentCallOutbound = &CallOutbound{
			action:   session.newAction("CallOutbound"),
			Peer:     peer,
			Local:    local,
			SocketFD: socketFD,
		}
		session.addAction(session.currentCallOutbound)
	}
	session.currentCallOutbound.Request = append(session.currentCallOutbound.Request, span...)
}

func (session *Session) SendUDP(ctx context.Context, span []byte, peer net.UDPAddr) {
	if session == nil {
		return
	}
	session.addAction(&SendUDP{
		action:  session.newAction("SendUDP"),
		Peer:    peer,
		Content: span,
	})
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
	if session.CallFromInbound == nil {
		return
	}
	if session.currentCallOutbound == nil || len(session.currentCallOutbound.Request) == 0 {
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

func (session *Session) addAction(action Action) {
	if !ShouldRecordAction(action) {
		return
	}
	session.Actions = append(session.Actions, action)
}
