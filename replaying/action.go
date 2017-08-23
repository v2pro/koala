package replaying

import (
	"time"
	"strconv"
	"net"
	"github.com/v2pro/koala/recording"
)

type replayedAction struct {
	actionId   string
	occurredAt int64
	actionType string
}

type ReplayedAction interface {
	ActionId() string
	ActionType() string
	OccurredAt() int64
}

func (action *replayedAction) ActionType() string {
	return action.actionType
}

func (action *replayedAction) ActionId() string {
	return action.actionId
}

func (action *replayedAction) OccurredAt() int64 {
	return action.occurredAt
}

func newReplayedAction(actionType string) replayedAction {
	occurredAt := time.Now().UnixNano()
	actionId := strconv.FormatInt(occurredAt, 10)
	return replayedAction{
		actionId:   actionId,
		occurredAt: occurredAt,
		actionType: actionType,
	}
}

type CallFromInbound struct {
	replayedAction
	Replayed *recording.CallFromInbound
}

type ReturnInbound struct {
	replayedAction
	Response []byte
}

type CallOutbound struct {
	replayedAction
	MatchedTalk      *recording.CallOutbound
	MatchedTalkIndex int
	MatchedTalkMark  float64
	Request          []byte
	Peer             net.TCPAddr
}

func NewCallOutbound(peer net.TCPAddr, request []byte) *CallOutbound {
	return &CallOutbound{
		replayedAction: newReplayedAction("CallOutbound"),
		Peer: peer,
		Request: request,
	}
}

type CallFunction struct {
	replayedAction
	CallFromFile string
	CallFromLine int
	CallIntoFile string
	CallIntoLine int
	FuncName     string
	Args         map[string]interface{}
}

type ReturnFunction struct {
	replayedAction
	CallFunctionId string
	ReturnValue    interface{}
}

type AppendFile struct {
	replayedAction
	FileName string
	Content  []byte
}
