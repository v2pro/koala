package replaying

import (
	"testing"
	"github.com/v2pro/koala/st"
	"github.com/stretchr/testify/require"
)

func Test_match_best_score(t *testing.T) {
	should := require.New(t)
	talk1 := st.Talk{Request: []byte{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8}}
	talk2 := st.Talk{Request: []byte{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 7}}
	replayingSession := ReplayingSession{
		Session: st.Session{
			OutboundTalks: []*st.Talk{&talk1, &talk2},
		},
	}
	_, matched := replayingSession.MatchOutboundTalk(-1, []byte{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8, })
	should.Equal(&talk1, matched)
}

func Test_match_not_matched(t *testing.T) {
	should := require.New(t)
	talk1 := st.Talk{Request: []byte{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8}}
	talk2 := st.Talk{Request: []byte{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8}}
	talk3 := st.Talk{Request: []byte{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8}}
	replayingSession := ReplayingSession{
		Session: st.Session{
			OutboundTalks: []*st.Talk{&talk1, &talk2, &talk3},
		},
	}
	index, _ := replayingSession.MatchOutboundTalk(-1, []byte{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8, })
	should.Equal(0, index)
	index, _ = replayingSession.MatchOutboundTalk(0, []byte{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8, })
	should.Equal(1, index)
}
