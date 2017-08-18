package recording

type Recorder interface {
	Record(session *Session)
}

var Recorders = []Recorder{}
