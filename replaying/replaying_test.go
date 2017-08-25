package replaying

import (
	"testing"
	"github.com/stretchr/testify/require"
	"github.com/v2pro/koala/recording"
	"io/ioutil"
	"encoding/json"
	"fmt"
	"encoding/base64"
)

func Test_match_best_score(t *testing.T) {
	should := require.New(t)
	talk1 := &recording.CallOutbound{Request: []byte{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8}}
	talk2 := &recording.CallOutbound{Request: []byte{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 7}}
	replayingSession := ReplayingSession{
		CallOutbounds: []*recording.CallOutbound{talk1, talk2},
	}
	_, _, matched := replayingSession.MatchOutboundTalk(nil, -1, []byte{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8,})
	should.Equal(&talk1, matched)
}

func Test_match_not_matched(t *testing.T) {
	should := require.New(t)
	talk1 := &recording.CallOutbound{Request: []byte{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8}}
	talk2 := &recording.CallOutbound{Request: []byte{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8}}
	talk3 := &recording.CallOutbound{Request: []byte{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8}}
	replayingSession := ReplayingSession{
		CallOutbounds: []*recording.CallOutbound{talk1, talk2, talk3},
	}
	index, _, _ := replayingSession.MatchOutboundTalk(nil, -1, []byte{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8,})
	should.Equal(0, index)
	index, _, _ = replayingSession.MatchOutboundTalk(nil, 0, []byte{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8,})
	should.Equal(1, index)
}

func Test_bad_case(t *testing.T) {
	should := require.New(t)
	bytes, err := ioutil.ReadFile("/tmp/koala-original-session.json")
	should.Nil(err)
	origSession := NewReplayingSession()
	err = json.Unmarshal(bytes, origSession)
	bytes, err = ioutil.ReadFile("/tmp/koala-replayed-session.json")
	should.Nil(err)
	var replayedSession interface{}
	err = json.Unmarshal(bytes, &replayedSession)
	should.Nil(err)

	fmt.Println(string(origSession.CallOutbounds[1].Request))
	reqStr := get(replayedSession, "Actions", 22, "Request").(string)
	req, _ := base64.StdEncoding.DecodeString(reqStr)
	fmt.Println(string(req))
	index, mark, matched := origSession.MatchOutboundTalk(nil, -1, req)
	should.NotNil(matched)
	fmt.Println(string(matched.Request))
	fmt.Println(mark)
	should.Equal(1, index)
}

func get(obj interface{}, keys ...interface{}) interface{} {
	for _, key := range keys {
		switch typedKey := key.(type) {
		case int:
			obj = obj.([]interface{})[typedKey]
		case string:
			obj = obj.(map[string]interface{})[typedKey]
		default:
			panic("unsupported key type")
		}
	}
	return obj
}
