package countlog


import (
	"runtime"
	"fmt"
)

const TRACE = 10
const DEBUG = 20
const INFO = 30
const WARN = 40
const ERROR = 50
const FATAL = 60

func Trace(event string, properties ...interface{}) {
	log(TRACE, event, properties)
}
func Debug(event string, properties ...interface{}) {
	log(DEBUG, event, properties)
}
func Info(event string, properties ...interface{}) {
	log(INFO, event, properties)
}
func Warn(event string, properties ...interface{}) {
	log(WARN, event, properties)
}
func Error(event string, properties ...interface{}) {
	log(ERROR, event, properties)
}
func Fatal(event string, properties ...interface{}) {
	log(FATAL, event, properties)
}
func Log(level int, event string, properties ...interface{}) {
	log(level, event, properties)
}
func log(level int, event string, properties []interface{}) {
	var expandedProperties []interface{}
	for _, logWriter := range LogWriters {
		if !logWriter.ShouldLog(level, event, properties) {
			continue
		}
		if expandedProperties == nil {
			expandedProperties = []interface{}{}
			for _, prop := range properties {
				propProvider, _ := prop.(func() interface{})
				if propProvider == nil {
					expandedProperties = append(expandedProperties, prop)
				} else {
					expandedProperties = append(expandedProperties, propProvider())
				}
			}
		}
		_, file, line, ok := runtime.Caller(1)
		if ok {
			expandedProperties = append(expandedProperties, "lineNumber")
			expandedProperties = append(expandedProperties, fmt.Sprintf("%s:%d", file, line))
		}
		logWriter.WriteLog(level, event, expandedProperties)
	}
}
