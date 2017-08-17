package replaying

import (
	"github.com/v2pro/koala/recording"
)

type ReplayedTalk struct {
	MatchedTalk   *recording.Talk
	MatchedTalkIndex int
	MatchedTalkMark float64
	ReplayedRequestTime int64
	ReplayedRequest []byte
	ReplayedResponseTime int64
}
