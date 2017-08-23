package replaying

import (
	"github.com/v2pro/koala/recording"
	"github.com/v2pro/koala/countlog"
	"context"
	"bytes"
)

type ReplayingSession struct {
	Session         *recording.Session
	callOutbounds   []*recording.CallOutbound
	actionCollector chan ReplayedAction
}

func NewReplayingSession(session *recording.Session) ReplayingSession {
	callOutbounds := []*recording.CallOutbound{}
	for _, action := range session.Actions {
		if action.GetActionType() == "CallOutbound" {
			callOutbounds = append(callOutbounds, action.(*recording.CallOutbound))
		}
	}
	return ReplayingSession{
		Session:         session,
		callOutbounds:   callOutbounds,
		actionCollector: make(chan ReplayedAction, 4096),
	}
}

func (replayingSession *ReplayingSession) CallOutbound(ctx context.Context, callOutbound *CallOutbound) {
	select {
	case replayingSession.actionCollector <- callOutbound:
	default:
		countlog.Error("event!replaying.ActionCollector is full", "ctx", ctx)
	}
}

func (replayingSession *ReplayingSession) AppendFile(ctx context.Context, content []byte, fileName string) {
	if replayingSession == nil {
		return
	}
	appendFile := &AppendFile{
		replayedAction: newReplayedAction("AppendFile"),
		FileName:       fileName,
		Content:        content,
	}
	select {
	case replayingSession.actionCollector <- appendFile:
	default:
		countlog.Error("event!replaying.ActionCollector is full", "ctx", ctx)
	}
}

func findReadableChunk(key []byte) (int, int) {
	start := bytes.IndexFunc(key, func(r rune) bool {
		return r > 31 && r < 127
	})
	if start == -1 {
		return -1, -1
	}
	end := bytes.IndexFunc(key[start:], func(r rune) bool {
		return r <= 31 || r >= 127
	})
	if end == -1 {
		return start, len(key) - start
	}
	return start, end
}

func (replayingSession *ReplayingSession) Finish(response []byte) *ReplayedSession {
	replayedSession := &ReplayedSession{
		SessionId: replayingSession.Session.SessionId,
		CallFromInbound: &CallFromInbound{
			replayedAction: newReplayedAction("CallFromInbound"),
			Replayed:       replayingSession.Session.CallFromInbound,
		},
	}
	replayedSession.ReturnInbound = &ReturnInbound{
		replayedAction: newReplayedAction("ReturnInbound"),
		Response:       response,
	}
	done := false
	appendFiles := map[string]*AppendFile{}
	for !done {
		select {
		case action := <-replayingSession.actionCollector:
			switch typedAction := action.(type) {
			case *AppendFile:
				existingAppendFile := appendFiles[typedAction.FileName]
				if existingAppendFile == nil {
					appendFiles[typedAction.FileName] = typedAction
					replayedSession.Actions = append(replayedSession.Actions, action)
				} else {
					existingAppendFile.Content = append(existingAppendFile.Content, typedAction.Content...)
				}
			default:
				replayedSession.Actions = append(replayedSession.Actions, action)
			}
		default:
			done = true
		}
	}
	replayedSession.Actions = append(replayedSession.Actions, replayedSession.ReturnInbound)
	return replayedSession
}
