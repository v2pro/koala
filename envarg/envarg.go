package envarg

import "os"

var isReplaying = false

func init() {
	isReplaying = os.Getenv("KOALA_MODE") == "REPLAYING"
}

func IsReplaying() bool {
	return isReplaying
}

func IsRecording() bool {
	return !isReplaying
}