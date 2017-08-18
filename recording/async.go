package recording

import (
	"github.com/v2pro/koala/countlog"
	"context"
)

type AsyncRecorder struct {
	Context      context.Context
	realRecorder Recorder
	recordChan   chan *Session
}

func NewAsyncRecorder(realRecorder Recorder) *AsyncRecorder {
	return &AsyncRecorder{
		recordChan:   make(chan *Session, 100),
		realRecorder: realRecorder,
	}
}

func (recorder *AsyncRecorder) Start() {
	go recorder.backgroundRecord()
}

func (recorder *AsyncRecorder) backgroundRecord() {
	defer func() {
		recovered := recover()
		if recovered != nil {
			countlog.Error("event!boss_recorder.panic",
				"err", recovered,
				"ctx", recorder.Context,
				"stacktrace", countlog.ProvideStacktrace)
		}
	}()
	for {
		session := <-recorder.recordChan
		countlog.Debug("event!boss_recorder.record_session",
			"ctx", recorder.Context,
			"session", session)
		recorder.realRecorder.Record(session)
	}
}
