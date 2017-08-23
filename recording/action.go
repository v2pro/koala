package recording

import (
	"time"
	"strconv"
	"net"
)

type action struct {
	actionId   string
	occurredAt int64
	actionType string
}

type Action interface {
	ActionType() string
}

func (action *action) ActionType() string {
	return action.actionType
}

func (action *action) ActionId() string {
	return action.actionId
}

func (action *action) OccurredAt() int64 {
	return action.occurredAt
}

func newAction(actionType string) action {
	occurredAt := time.Now().UnixNano()
	actionId := strconv.FormatInt(occurredAt, 10)
	return action{
		actionId:   actionId,
		occurredAt: occurredAt,
		actionType: actionType,
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