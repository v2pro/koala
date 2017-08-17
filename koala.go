// api for go application using koala, put Start() in the entry of main
package koala

import (
	"github.com/v2pro/koala/gateway/gw4go"
	"github.com/v2pro/koala/internal"
)

func Start() {
	gw4go.Start()
}

func SetDelegatedFromGoRoutineId(goid int64) {
	internal.SetDelegatedFromGoRoutineId(goid)
}

func GetCurrentGoRoutineId() int64 {
	return internal.GetCurrentGoRoutineId()
}
