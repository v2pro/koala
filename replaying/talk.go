package replaying

import "github.com/v2pro/koala/st"

type ReplayedTalk struct {
	MatchedTalk   *st.Talk
	ReplayedRequestTime int64
	ReplayedRequest []byte
	ReplayedResponseTime int64
}
