package main

import (
	_ "github.com/v2pro/koala/gateway/gw4libc"
	"github.com/v2pro/koala/envarg"
	"github.com/v2pro/plz/witch"
)

func init() {
	envarg.SetupLogging()
	witch.StartViewer(":8318")
}

func main() {
}
