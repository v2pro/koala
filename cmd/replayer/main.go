package main

import (
	_ "github.com/v2pro/koala/gateway/gw4libc"
	"github.com/v2pro/koala/countlog"
	"github.com/v2pro/koala/envarg"
)

func init() {
	startLogging()
}

func startLogging() {
	if len(countlog.LogWriters) != 0 {
		// extension already setup alternative log writers
		return
	}
	logWriter := countlog.NewAsyncLogWriter(envarg.LogLevel(), countlog.NewFileLogOutput(envarg.LogFile()))
	logWriter.LogFormatter = &countlog.HumanReadableFormat{
		ContextPropertyNames: []string{"threadID", "outboundSrc"},
		StringLengthCap:      512,
	}
	logWriter.EventWhitelist["event!replaying.talks_scored"] = true
	//logWriter.EventWhitelist["event!sut.opening_file"] = true
	logWriter.Start()
	countlog.LogWriters = append(countlog.LogWriters, logWriter)
}

func main() {
}