package countlog

import (
	"os"
	"unsafe"
	"path/filepath"
	"time"
)

type FileLogWriter struct {
	MinLevel            int
	EventWhitelist      map[string]bool
	msgChan             chan Event
	LogFormatter        LogFormatter
	writeLog            func(timestamp int64, formattedEvent []byte)
	openedFile          *os.File
	openedFileArchiveTo string
	isClosed            chan bool
}

func (logWriter *FileLogWriter) ShouldLog(level int, event string, properties []interface{}) bool {
	if logWriter.EventWhitelist[event] {
		return true
	}
	return level >= logWriter.MinLevel
}

func (logWriter *FileLogWriter) WriteLog(level int, event string, properties []interface{}) {
	select {
	case logWriter.msgChan <- Event{Event: event, Properties: properties}:
	default:
		// drop on the floor
	}
}

func (logWriter *FileLogWriter) Close() {
	close(logWriter.isClosed)
	if logWriter.openedFile != nil {
		logWriter.openedFile.Close()
	}
}

func (logWriter *FileLogWriter) Start() {
	LogWriters = append(LogWriters, logWriter)
	go func() {
		for {
			select {
			case event := <-logWriter.msgChan:
				formattedEvent := logWriter.LogFormatter.FormatLog(event)
				logWriter.writeLog(
					event.Properties[1].(int64),
					*(*[]byte)(unsafe.Pointer(&formattedEvent)))
			case <-logWriter.isClosed:
				return
			}
		}
	}()
}

func (logWriter *FileLogWriter) openLogFile(logFile string) {
	var err error
	logWriter.openedFile, err = os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		os.Stderr.Write([]byte("failed to open log file: " +
			logFile + ", " + err.Error() + "\n"))
		os.Stderr.Sync()
	}
	logWriter.openedFileArchiveTo = logFile + "." + time.Now().Format("200601021504")
}

func (logWriter *FileLogWriter) archiveLogFile(logFile string) {
	logWriter.openedFile.Close()
	logWriter.openedFile = nil
	err := os.Rename(logFile, logWriter.openedFileArchiveTo)
	if err != nil {
		os.Stderr.Write([]byte("failed to rename to archived log file: " +
			logWriter.openedFileArchiveTo + ", " + err.Error() + "\n"))
		os.Stderr.Sync()
	}
}

func NewFileLogWriter(minLevel int, logFile string) *FileLogWriter {
	writer := &FileLogWriter{
		MinLevel:     minLevel,
		msgChan:      make(chan Event, 1024),
		LogFormatter: &HumanReadableFormat{},
		writeLog: func(timestamp int64, formattedEvent []byte) {
			os.Stdout.Write(formattedEvent)
		},
		EventWhitelist: map[string]bool{},
	}
	switch logFile {
	case "STDOUT":
		writer.writeLog = func(timestamp int64, formattedEvent []byte) {
			os.Stdout.Write(formattedEvent)
		}
	case "STDERR":
		writer.writeLog = func(timestamp int64, formattedEvent []byte) {
			os.Stderr.Write(formattedEvent)
		}
	default:
		err := os.MkdirAll(filepath.Dir(logFile), 0755)
		if err != nil {
			os.Stderr.Write([]byte("failed to create dir for log file: " +
				filepath.Dir(logFile) + ", " + err.Error() + "\n"))
			os.Stderr.Sync()
		}
		writer.openLogFile(logFile)
		windowSize := int64(time.Hour)
		rotateAfter := (int64(time.Now().UnixNano()/windowSize) + 1) * windowSize
		writer.writeLog = func(timestamp int64, formattedEvent []byte) {
			if timestamp > rotateAfter {
				now := time.Now()
				rotateAfter = (int64(now.UnixNano()/windowSize) + 1) * windowSize
				writer.archiveLogFile(logFile)
				writer.openLogFile(logFile)
			}
			if writer.openedFile != nil {
				writer.openedFile.Write(formattedEvent) // silently ignore error
			}
		}
	}
	return writer
}
