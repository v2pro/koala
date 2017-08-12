package countlog

import "fmt"

type StdoutLogWriter struct {
	MinLevel int
	msgChan  chan Event
	FormatLog func(event Event) string
}

func (logWriter *StdoutLogWriter) ShouldLog(level int, event string, properties []interface{}) bool {
	return level >= logWriter.MinLevel
}

func (logWriter *StdoutLogWriter) WriteLog(level int, event string, properties []interface{}) {
	select {
	case logWriter.msgChan <- Event{Event: event, Properties: properties}:
	default:
		// drop on the floor
	}
}

func (logWriter *StdoutLogWriter) Start() {
	LogWriters = append(LogWriters, logWriter)
	go func() {
		for {
			event := <-logWriter.msgChan
			if logWriter.FormatLog != nil {
				fmt.Println(logWriter.FormatLog(event))
			} else {
				fmt.Println(event.Event, event.Properties)
			}
		}
	}()
}

func NewStdoutLogWriter(minLevel int) *StdoutLogWriter {
	return &StdoutLogWriter{
		MinLevel: minLevel,
		msgChan:  make(chan Event, 1024),
	}
}
