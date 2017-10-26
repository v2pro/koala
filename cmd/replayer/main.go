package main

import (
	_ "github.com/v2pro/koala/gateway/gw4libc"
	"github.com/v2pro/koala/envarg"
)

func init() {
	envarg.SetupLogging()
}

func main() {
}
