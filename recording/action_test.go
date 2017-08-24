package recording

import (
	"testing"
	"github.com/stretchr/testify/require"
	"encoding/json"
)

func Test_marshal_append_file(t *testing.T) {
	should := require.New(t)
	bytes, err := json.Marshal(&AppendFile{
		FileName: "/abc",
		Content:  []byte("hello"),
	})
	should.Nil(err)
	should.Contains(string(bytes), "hello")
}

func Test_marshal_call_outbound(t *testing.T) {
	should := require.New(t)
	bytes, err := json.Marshal(&CallOutbound{
		Request:  []byte("hello"),
		Response: []byte("world"),
	})
	should.Nil(err)
	should.Contains(string(bytes), "hello")
	should.Contains(string(bytes), "world")
}

func Test_marshal_return_inbound(t *testing.T) {
	should := require.New(t)
	bytes, err := json.Marshal(&ReturnInbound{
		Response: []byte("hello"),
	})
	should.Nil(err)
	should.Contains(string(bytes), "hello")
}

func Test_marshal_call_from_inbound(t *testing.T) {
	should := require.New(t)
	bytes, err := json.Marshal(&CallFromInbound{
		Request: []byte("hello"),
	})
	should.Nil(err)
	should.Contains(string(bytes), "hello")
}

func Test_marshal_session(t *testing.T) {
	session := Session{
		CallFromInbound: &CallFromInbound{
			Request: []byte("hello"),
		},
		ReturnInbound: &ReturnInbound{
			Response: []byte("hello"),
		},
		Actions: []Action{
			&CallOutbound{
				Request:  []byte("hello"),
				Response: []byte("world"),
			},
			&AppendFile{
				FileName: "/abc",
				Content:  []byte("hello"),
			},
		},
	}
	bytes, err := json.MarshalIndent(session, "", "  ")
	should := require.New(t)
	should.Nil(err)
	should.NotContains(string(bytes), "=") // no base64
}
