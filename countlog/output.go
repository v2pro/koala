package countlog

type LogOutput interface {
	OutputLog(timestamp int64, formattedEvent []byte)
	Close()
}