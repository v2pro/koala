package countlog


import (
	"runtime"
	"fmt"
)

const LEVEL_TRACE = 10
const LEVEL_DEBUG = 20
const LEVEL_INFO = 30
const LEVEL_WARN = 40
const LEVEL_ERROR = 50
const LEVEL_FATAL = 60

func Trace(event string, properties ...interface{}) {
	log(LEVEL_TRACE, event, properties)
}
func Debug(event string, properties ...interface{}) {
	log(LEVEL_DEBUG, event, properties)
}
func Info(event string, properties ...interface{}) {
	log(LEVEL_INFO, event, properties)
}
func Warn(event string, properties ...interface{}) {
	log(LEVEL_WARN, event, properties)
}
func Error(event string, properties ...interface{}) {
	log(LEVEL_ERROR, event, properties)
}
func Fatal(event string, properties ...interface{}) {
	log(LEVEL_FATAL, event, properties)
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
