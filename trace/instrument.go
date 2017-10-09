package trace

import (
	"bytes"
	"strings"
	"regexp"
	"fmt"
	"io/ioutil"
	"github.com/v2pro/plz/countlog"
	"sync"
	"os"
	"time"
	"crypto/sha1"
	"encoding/base32"
	"github.com/v2pro/koala/envarg"
)

var re_function *regexp.Regexp

type instrumentedFile struct {
	modTime              time.Time
	size                 int64
	instrumentedFileName string
}

var instrumentedFiles = map[string]*instrumentedFile{}
var instrumentedFilesMutex = &sync.Mutex{}

func init() {
	if !envarg.IsReplaying() {
		return
	}
	if _, err := os.Stat("/tmp/koala-instrumented-files"); err != nil {
		// dir not created yet, create
		err = os.Mkdir("/tmp/koala-instrumented-files", 0777)
		if err != nil {
			countlog.Error("event!trace.failed to create instrumented dir", "err", err)
		}
	}
	var err error
	re_function, err = regexp.Compile(`((static[\s\r\n]+)?(public[\s\r\n]+)?(private[\s\r\n]+)?(static[\s\r\n]+)?)function[\s\r\n]*(\w+)\(([\s\r\n\w$&,='"]*)\)[\s\r\n]*{`)
	if err != nil {
		countlog.Error("event!trace.failed to compile regex", "err", err)
	}
}

func InstrumentFile(fileName string) string {
	if !strings.HasSuffix(fileName, ".php") {
		return ""
	}
	fileInfo, err := os.Stat(fileName)
	if err != nil {
		countlog.Trace("event!trace.failed to stat file", "err", err, "fileName", fileName)
		return ""
	}
	cache := getInstrumentedFile(fileName)
	if cache != nil &&
		cache.modTime == fileInfo.ModTime() &&
		cache.size == fileInfo.Size() {
		return cache.instrumentedFileName
	}
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		countlog.Trace("event!trace.failed to read file", "err", err, "fileName", fileName)
		return ""
	}
	content = []byte(Instrument(fileName, string(content)))
	instrumentedFileName := "/tmp/koala-instrumented-files/" + hash(content)
	err = ioutil.WriteFile(instrumentedFileName + ".tmp", content, 0666)
	if err != nil {
		countlog.Error("event!sut.failed to write instrumented file",
			"instrumentedFileName", instrumentedFileName, "err", err)
		return ""
	}
	err = os.Rename(instrumentedFileName + ".tmp", instrumentedFileName)
	if err != nil {
		countlog.Error("event!sut.failed to rename instrumented file tmp",
			"instrumentedFileName", instrumentedFileName, "err", err)
		return ""
	}
	setInstrumentedFile(fileName, &instrumentedFile{
		modTime: fileInfo.ModTime(),
		size: fileInfo.Size(),
		instrumentedFileName: instrumentedFileName,
	})
	return instrumentedFileName
}

func getInstrumentedFile(fileName string) *instrumentedFile {
	instrumentedFilesMutex.Lock()
	defer instrumentedFilesMutex.Unlock()
	return instrumentedFiles[fileName]
}

func setInstrumentedFile(fileName string, cache *instrumentedFile) {
	instrumentedFilesMutex.Lock()
	defer instrumentedFilesMutex.Unlock()
	instrumentedFiles[fileName] = cache
}

func hash(content []byte) string {
	h := sha1.New()
	h.Write(content)
	return "g" + base32.StdEncoding.EncodeToString(h.Sum(nil))
}

func Instrument(fileName string, input string) string {
	buf := &bytes.Buffer{}
	lineNo := 0
	for {
		exitLineNo, leftIdx, rightIdx, substitution := subPhpFunctionDefinition(fileName, lineNo, input, phpTracepoint)
		lineNo = exitLineNo
		if substitution == "" {
			buf.WriteString(input)
			return buf.String()
		}
		buf.WriteString(input[:leftIdx])
		buf.WriteString(substitution)
		input = input[rightIdx:]
	}
}

func subPhpFunctionDefinition(fileName string, enterLineNo int, src string,
	bodyProvider func(fileName string, lineNo int, isDefinedInClass bool, functionName string, argumentsDef string) string) (exitLineNo int, leftIdx int, rightIdx int, substitution string) {
	idx := re_function.FindStringSubmatchIndex(src)
	if len(idx) == 0 {
		return -1, -1, -1, ""
	}
	leftIdx = idx[0]
	rightIdx = idx[1]
	functionSignature := src[leftIdx:rightIdx]
	functionNameLeftIdx := idx[12]
	functionNameRightIdx := idx[13]
	functionName := src[functionNameLeftIdx:functionNameRightIdx]
	underscoreFunctionName := fmt.Sprintf("orig_%v", functionName)
	functionModifier := src[idx[2]:idx[3]]
	definedInClass := !isBlank(functionModifier)
	newlineCount, lastLine := countNewLines(src[:functionNameLeftIdx])
	if strings.IndexByte(lastLine, '#') != -1 || strings.Index(lastLine, "//") != -1 {
		// if we add tracepoint after line comment
		// the newline added will escape the comment scope
		// this is a fix for this particular fix
		// for block comment, it is rare to encounter problem
		// however, regex is incapable of handling complex comment scopes
		// we are not trying to make this bullet-proof
		nextLineStart := strings.IndexByte(src[functionNameLeftIdx:], '\n')
		if nextLineStart == -1 {
			return -1, -1, -1, ""
		}
		nextLineStart += functionNameLeftIdx
		exitLineNo, leftIdx, rightIdx, substitution = subPhpFunctionDefinition(
			fileName, enterLineNo+newlineCount,
			src[nextLineStart+1:], bodyProvider)
		if leftIdx != -1 {
			leftIdx += nextLineStart + 1
		}
		if rightIdx != -1 {
			rightIdx += nextLineStart + 1
		}
		return
	}
	exitLineNo = enterLineNo + newlineCount
	functionBody := bodyProvider(fileName, exitLineNo, definedInClass, functionName, src[idx[14]:idx[15]])
	substitution = fmt.Sprintf(
		"%s %s }\n%s%s%s", functionSignature, functionBody,
		src[leftIdx:functionNameLeftIdx],
		underscoreFunctionName,
		src[functionNameRightIdx:rightIdx])
	return
}

func phpTracepoint(fileName string, lineNo int, isDefinedInClass bool, functionName string, argumentsDef string) string {
	underscoreFunctionName := fmt.Sprintf("orig_%v", functionName)
	buf := &bytes.Buffer{}
	for i, argDef := range strings.Split(argumentsDef, ",") {
		if i != 0 {
			buf.WriteByte(',')
		}
		dollarIdx := strings.IndexByte(argDef, '$')
		equalIdx := strings.IndexByte(argDef, '=')
		if dollarIdx == -1 {
			dollarIdx = 0
		}
		if equalIdx == -1 {
			equalIdx = len(argDef)
		}
		buf.WriteString(argDef[dollarIdx:equalIdx])
	}
	invokeArgs := buf.String()
	var delegateInvocation string
	if isDefinedInClass {
		delegateInvocation = fmt.Sprintf(`
$response = self::%s(%s);
`,
			underscoreFunctionName, invokeArgs)
	} else {
		delegateInvocation = fmt.Sprintf(`
$orig=function(%s){ return self::%s(%s); };
if (empty(__CLASS__)) { $response = %s(%s); }
else { $response = $orig(%s); }
`,
			argumentsDef, underscoreFunctionName, invokeArgs, // $orig=function(%s){ return self::%s(%s); };
			underscoreFunctionName, invokeArgs,               // if (empty(__CLASS__)) { return %s(%s); }
			invokeArgs) // else { return $orig(%s); }
	}
	return fmt.Sprintf(`
$sock = $GLOBALS['koala_helper_sock'] ?? socket_create(AF_INET, SOCK_DGRAM, SOL_UDP);
$actionId = sprintf('%%.0f', microtime(true) * 1000 * 1000 * 1000);
$args = [%s];
$encodedArgs = [];
foreach($args as $arg) {
	$encodedArg = json_encode($arg, JSON_PARTIAL_OUTPUT_ON_ERROR);
	if (strlen($encodedArg) > 4096) {
		$encodedArg = null;
	}
	$encodedArgs []= $encodedArg;
}
$callFunction = "to-koala!call-function\n" . json_encode([
	'ActionId' => $actionId,
	'CallIntoFile' => '%s',
	'CallIntoLine' => %v,
	'FuncName' => empty(__CLASS__) ? '%s' : strval(__CLASS__) . '::%s',
	'Args' => $encodedArgs,
], JSON_PARTIAL_OUTPUT_ON_ERROR);
socket_sendto($sock, $callFunction, strlen($callFunction), 0, '127.127.127.127', 127);
try {
	%s
	return $response;
} finally {
	$returnFunction = "to-koala!return-function\n" . json_encode([
		'CallFunctionId' => $actionId,
		'ReturnValue' => json_encode($response, JSON_PARTIAL_OUTPUT_ON_ERROR),
	], JSON_PARTIAL_OUTPUT_ON_ERROR);
	socket_sendto($sock, $returnFunction, strlen($returnFunction), 0, '127.127.127.127', 127);
}
	`,invokeArgs, fileName, lineNo, functionName, functionName, delegateInvocation)
}

func countNewLines(str string) (count int, lastLine string) {
	lastLine = str
	for {
		found := strings.IndexByte(lastLine, '\n')
		if found == -1 {
			return
		}
		count++
		lastLine = lastLine[found+1:]
	}
	return
}

func isBlank(str string) bool {
	if str == "" {
		return true
	}
	for _, c := range str {
		if c != ' ' {
			return false
		}
	}
	return true
}
