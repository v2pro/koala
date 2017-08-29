package trace

import (
	"bytes"
	"strings"
	"regexp"
	"fmt"
	"io/ioutil"
	"github.com/v2pro/koala/countlog"
)

var re_function *regexp.Regexp

func init() {
	var err error
	re_function, err = regexp.Compile(`((static[\s\r\n]+)?(public[\s\r\n]+)?(private[\s\r\n]+)?(static[\s\r\n]+)?)function[\s\r\n]*(\w+)\(([\s\r\n\w$&,='"]*)\)[\s\r\n]*{`)
	if err != nil {
		panic(fmt.Sprintf("failed to compile regex: %s", err.Error()))
	}
}

func MockFile(fileName string) []byte {
	if !strings.HasSuffix(fileName, ".php") {
		return nil
	}
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		countlog.Trace("event!trace.failed to read file", "err", err, "fileName", fileName)
		return nil
	}
	return []byte(Instrument(fileName, string(content)))
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
	if strings.IndexByte(lastLine, '#') != -1 {
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
$callFunction = "to-koala:call-function\n" . json_encode([
	'ActionId' => $actionId,
	'CallIntoFile' => '%s',
	'CallIntoLine' => %v,
	'FuncName' => '%s',
	'Args' => [%s],
], JSON_PARTIAL_OUTPUT_ON_ERROR);
socket_sendto($sock, $callFunction, strlen($callFunction), 0, '127.127.127.127', 127);
%s
$returnFunction = "to-koala:return-function\n" . json_encode([
	'CallFunctionId' => $actionId,
	'ReturnValue' => $response,
], JSON_PARTIAL_OUTPUT_ON_ERROR);
socket_sendto($sock, $returnFunction, strlen($returnFunction), 0, '127.127.127.127', 127);
return $response;
	`, fileName, lineNo, functionName, invokeArgs, delegateInvocation)
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
