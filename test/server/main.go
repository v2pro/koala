package main

import (
	"net/http"
	"github.com/v2pro/koala/koala"
	"github.com/v2pro/koala/countlog"
	"github.com/v2pro/koala/internal"
)

func main() {
	koala.Start()
	http.HandleFunc("/", func(respWriter http.ResponseWriter, req *http.Request) {
		countlog.Info("event!test_server.enter_handler",
			"threadID", internal.GetCurrentGoRoutineId(),
			"url", req.URL.String())
		_, err := http.Get("http://127.0.0.1:1/not-exist")
		if err != nil {
			respWriter.Write([]byte(err.Error()))
			return
		}
		respWriter.Write([]byte("good day"))
	})
	http.ListenAndServe("127.0.0.1:2515", nil)
}
