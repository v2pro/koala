package replaying

import (
	"testing"
	"github.com/stretchr/testify/require"
	"github.com/v2pro/koala/recording"
	"io/ioutil"
	"encoding/json"
	"fmt"
)

func Test_match_best_score(t *testing.T) {
	should := require.New(t)
	talk1 := recording.Talk{Request: []byte{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8}}
	talk2 := recording.Talk{Request: []byte{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 7}}
	replayingSession := ReplayingSession{
		Session: recording.Session{
			OutboundTalks: []*recording.Talk{&talk1, &talk2},
		},
	}
	_, _, matched := replayingSession.MatchOutboundTalk(nil, -1, []byte{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8, })
	should.Equal(&talk1, matched)
}

func Test_match_not_matched(t *testing.T) {
	should := require.New(t)
	talk1 := recording.Talk{Request: []byte{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8}}
	talk2 := recording.Talk{Request: []byte{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8}}
	talk3 := recording.Talk{Request: []byte{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8}}
	replayingSession := ReplayingSession{
		Session: recording.Session{
			OutboundTalks: []*recording.Talk{&talk1, &talk2, &talk3},
		},
	}
	index, _, _ := replayingSession.MatchOutboundTalk(nil, -1, []byte{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8, })
	should.Equal(0, index)
	index, _, _ = replayingSession.MatchOutboundTalk(nil, 0, []byte{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8, })
	should.Equal(1, index)
}

func Test_bad_case(t *testing.T) {
	should := require.New(t)
	bytes, err := ioutil.ReadFile("/tmp/koala-orig.json")
	should.Nil(err)
	origSession := ReplayingSession{
	}
	err = json.Unmarshal(bytes, &origSession.Session)
	bytes, err = ioutil.ReadFile("/tmp/koala-1.json")
	should.Nil(err)
	replayedSession := ReplayingSession{
	}
	err = json.Unmarshal(bytes, &replayedSession)
	should.Nil(err)

	req := replayedSession.ReplayedOutboundTalks[94].ReplayedRequest
	fmt.Println(string(req))
	fmt.Println(req)
	index, _, matched := origSession.MatchOutboundTalk(nil, -1,
		req)
	fmt.Println(string(matched.Request))
	should.Equal(1, index)
}
