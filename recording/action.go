package recording

import (
	"net"
	"encoding/json"
)

type action struct {
	ActionIndex int
	OccurredAt  int64
	ActionType  string
}

type Action interface {
	GetActionIndex() int
	GetOccurredAt() int64
	GetActionType() string
}

func (action *action) GetActionType() string {
	return action.ActionType
}

func (action *action) GetActionIndex() int {
	return action.ActionIndex
}

func (action *action) GetOccurredAt() int64 {
	return action.OccurredAt
}

type CallFromInbound struct {
	action
	Peer    net.TCPAddr
	Request []byte
}

func (callFromInbound *CallFromInbound) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		CallFromInbound
		Request string
	}{
		CallFromInbound: *callFromInbound,
		Request:         string(callFromInbound.Request),
	})
}

type ReturnInbound struct {
	action
	Response []byte
}

func (returnInbound *ReturnInbound) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ReturnInbound
		Response string
	}{
		ReturnInbound: *returnInbound,
		Response:      string(returnInbound.Response),
	})
}

type CallOutbound struct {
	action
	Peer         net.TCPAddr
	Request      []byte
	ResponseTime int64
	Response     []byte
}

func (callOutbound *CallOutbound) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		CallOutbound
		Request  string
		Response string
	}{
		CallOutbound: *callOutbound,
		Request:      string(callOutbound.Request),
		Response:     string(callOutbound.Response),
	})
}

type AppendFile struct {
	action
	FileName string
	Content  []byte
}

func (appendFile *AppendFile) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		AppendFile
		Content string
	}{
		AppendFile: *appendFile,
		Content:    string(appendFile.Content),
	})
}

type SendUDP struct {
	action
	Peer    net.UDPAddr
	Content []byte
}

func (sendUDP *SendUDP) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		SendUDP
		Content string
	}{
		SendUDP: *sendUDP,
		Content: string(sendUDP.Content),
	})
}
