package recording

import (
	"time"
	"strconv"
	"net"
)

type action struct {
	ActionId   string
	OccurredAt int64
	ActionType string
}

type Action interface {
	GetActionId() string
	GetOccurredAt() int64
	GetActionType() string
}

func (action *action) GetActionType() string {
	return action.ActionType
}

func (action *action) GetActionId() string {
	return action.ActionId
}

func (action *action) GetOccurredAt() int64 {
	return action.OccurredAt
}

func newAction(actionType string) action {
	occurredAt := time.Now().UnixNano()
	actionId := strconv.FormatInt(occurredAt, 10)
	return action{
		ActionId:   actionId,
		OccurredAt: occurredAt,
		ActionType: actionType,
	}
}

type CallFromInbound struct {
	action
	Peer    net.TCPAddr
	Request []byte
}

type ReturnInbound struct {
	action
	Response     []byte
}

type CallOutbound struct {
	action
	Peer         net.TCPAddr
	Request      []byte
	ResponseTime int64
	Response     []byte
}

type AppendFile struct {
	action
	FileName string
	Content  []byte
}