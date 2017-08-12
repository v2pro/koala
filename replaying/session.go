package replaying

import (
	"github.com/v2pro/koala/st"
	"time"
)

type ReplayingSession struct {
	st.Session `json:"-"`
	ReplayedOutboundTalkCollector chan ReplayedTalk `json:"-"`
	ReplayedRequestTime           int64
	ReplayedResponse              []byte
	ReplayedResponseTime          int64
	ReplayedOutboundTalks         []ReplayedTalk
}

func (replayingSession *ReplayingSession) MatchOutboundTalk(outboundRequest []byte) *st.Talk {
	return replayingSession.OutboundTalks[0]
}

func (replayingSession *ReplayingSession) Finish(response []byte) {
	replayingSession.ReplayedResponse = response
	replayingSession.ReplayedResponseTime = time.Now().UnixNano()
	done := false
	for !done {
		select {
		case replayedTalk := <- replayingSession.ReplayedOutboundTalkCollector:
			replayingSession.ReplayedOutboundTalks = append(replayingSession.ReplayedOutboundTalks, replayedTalk)
		default:
			done = true
		}
	}
}
