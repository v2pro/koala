package envarg

import (
	"os"
	"github.com/v2pro/plz/countlog"
	"github.com/v2pro/plz/countlog/output/hrf"
	"github.com/v2pro/plz/countlog/output"
	"github.com/v2pro/plz/countlog/output/compact"
	"io"
	"github.com/v2pro/plz/countlog/spi"
)

func SetupLogging() {
	countlog.SetMinLevel(LogLevel())
	countlog.EventWriter = output.NewEventWriter(output.EventWriterConfig{
		Format: createLogFormat(),
		Writer: openLogFile(),
	})
}

func openLogFile() io.Writer {
	fileName := LogFile()
	switch fileName {
	case "STDOUT":
		return os.Stdout
	case "STDERR":
		return os.Stderr
	default:
		file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			spi.OnError(err)
			return os.Stderr
		}
		return file
	}
}

func createLogFormat() output.Format {
	switch LogFormat() {
	case "HumanReadableFormat":
		return &hrf.Format{}
	case "CompactFormat":
		return &compact.Format{}
	default:
		os.Stderr.WriteString("unknown LogFormat: " + LogFormat() + "\n")
		os.Stderr.Sync()
		return &compact.Format{}
	}
}
