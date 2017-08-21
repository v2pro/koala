package replaying

import (
	"github.com/v2pro/koala/recording"
	"time"
	"strconv"
	"net"
)

type Action struct {
	ActionId   string
	OccurredAt int64
	ActionType string
}

func NewAction(actionType string) Action {
	occurredAt := time.Now().UnixNano()
	actionId := strconv.FormatInt(occurredAt, 10)
	return Action{
		ActionId:   actionId,
		OccurredAt: occurredAt,
		ActionType: actionType,
	}
}

type CallFromInbound struct {
	Action
	ReplayedTalk *recording.Talk
}

type CallOutbound struct {
	Action
	MatchedTalk      *recording.Talk
	MatchedTalkIndex int
	MatchedTalkMark  float64
	Request          []byte
	Peer             net.TCPAddr
}

type ReturnInbound struct {
	Action
	Response []byte
}

type CallFunction struct {
	Action
	CallFromFile string
	CallFromLine int
	CallIntoFile string
	CallIntoLine int
	FuncName     string
	Args         map[string]interface{}
}

type ReturnFunction struct {
	Action
	CallFunctionId string
	ReturnValue    interface{}
}

type AppendFile struct {
	Action
	FileName string
	Content  []byte
}
