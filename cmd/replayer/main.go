package main

import (
	_ "github.com/v2pro/koala/gateway/gw4libc"
	"github.com/v2pro/koala/envarg"
	"github.com/v2pro/plz/witch"
)

func init() {
	envarg.SetupLogging()
	witch.Start(":8318")
}

func main() {
}
