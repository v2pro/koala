package countlog

type LogFormatter interface {
	FormatLog(event Event) []byte
}