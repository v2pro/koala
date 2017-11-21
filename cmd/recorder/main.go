package main

import (
	_ "github.com/v2pro/koala/gateway/gw4libc"
	"github.com/v2pro/koala/envarg"
	"github.com/v2pro/plz/witch"
	"github.com/v2pro/koala/recording"
	"github.com/v2pro/plz/countlog"
	"encoding/json"
	"io/ioutil"
	"path"
)

func init() {
	envarg.SetupLogging()
	witch.StartViewer(":8318")
	dir := envarg.GetenvFromC("KOALA_RECORD_TO_DIR")
	if dir == "" {
		countlog.Fatal("event!recorder.pleases specify KOALA_RECORD_TO_DIR")
		return
	}
	recording.Recorders = append(recording.Recorders, &fileRecorder{dir:dir})
}

type fileRecorder struct {
	dir string
}

func (recorder *fileRecorder) Record(session *recording.Session) {
	bytes, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		countlog.Error("event!recorder.failed to marshal json", "err", err)
		return
	}
	ioutil.WriteFile(path.Join(recorder.dir, session.SessionId), bytes, 0666)
}

func main() {

}
