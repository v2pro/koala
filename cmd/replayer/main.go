package main

import (
	_ "github.com/v2pro/koala/gateway/gw4libc"
	"github.com/v2pro/koala/envarg"
	"github.com/v2pro/plz/witch"
	"github.com/v2pro/plz/countlog"
)

func init() {
	envarg.SetupLogging()
	countlog.LogWriters = append(countlog.LogWriters, witch.TheEventQueue)
	witch.StartViewer(":8318")
}

func main() {
}
