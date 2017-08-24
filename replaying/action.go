package replaying

import (
	"time"
	"strconv"
	"net"
	"github.com/v2pro/koala/recording"
)

type replayedAction struct {
	ActionId   string
	OccurredAt int64
	ActionType string
}

type ReplayedAction interface {
	GetActionId() string
	GetActionType() string
	GetOccurredAt() int64
}

func (action *replayedAction) GetActionType() string {
	return action.ActionType
}

func (action *replayedAction) GetActionId() string {
	return action.ActionId
}

func (action *replayedAction) GetOccurredAt() int64 {
	return action.OccurredAt
}

func newReplayedAction(actionType string) replayedAction {
	occurredAt := time.Now().UnixNano()
	actionId := strconv.FormatInt(occurredAt, 10)
	return replayedAction{
		ActionId:   actionId,
		OccurredAt: occurredAt,
		ActionType: actionType,
	}
}

type CallFromInbound struct {
	replayedAction
	OriginalRequestTime int64
	OriginalRequest []byte
}

type ReturnInbound struct {
	replayedAction
	OriginalResponse []byte
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
		Peer:           peer,
		Request:        request,
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
