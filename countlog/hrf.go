package countlog

import (
	"fmt"
	"net"
	"encoding/base64"
	"context"
)

type HumanReadableFormat struct {
	ContextPropertyNames []string
}

func (hrf *HumanReadableFormat) FormatLog(event Event) string {
	msg := []byte{}
	ctx := hrf.describeContext(event)
	if len(ctx) == 0 {
		msg = append(msg, fmt.Sprintf(
			"=== %s ===\n", event.Event)...)
	} else {
		msg = append(msg, fmt.Sprintf(
			"=== [%s] %s ===\n", string(ctx), event.Event)...)
	}
	for i := 0; i < len(event.Properties); i += 2 {
		k, _ := event.Properties[i].(string)
		switch k {
		case "", "ctx", "timestamp":
			continue
		}
		v := event.Properties[i+1]
		formattedV := ""
		switch typedV := v.(type) {
		case []byte:
			buf := typedV
			if isBinary(buf) {
				formattedV = base64.StdEncoding.EncodeToString(buf)
			} else {
				formattedV = string(buf)
			}
		case net.TCPAddr:
			formattedV = typedV.String()
		case *net.TCPAddr:
			formattedV = typedV.String()
		default:
			formattedV = fmt.Sprintf("%v", typedV)
		}
		msg = append(msg, k...)
		msg = append(msg, ": "...)
		msg = append(msg, formattedV...)
		msg = append(msg, '\n')
	}
	return string(msg)
}

func (hrf *HumanReadableFormat) describeContext(event Event) []byte {
	msg := []byte{}
	ctx, _ := event.Get("ctx").(context.Context)
	for _, propName := range hrf.ContextPropertyNames {
		propValue := event.Get(propName)
		if propValue == nil && ctx != nil {
			propValue = ctx.Value(propName)
		}
		if propValue != nil {
			if len(msg) > 0 {
				msg = append(msg, ',')
			}
			msg = append(msg, propName...)
			msg = append(msg, '=')
			msg = append(msg, fmt.Sprintf("%v", propValue)...)
		}
	}
	return msg
}

func isBinary(buf []byte) bool {
	for _, b := range buf {
		if b == '\r' || b == '\n' {
			continue
		}
		if b < 32 || b > 127 {
			return true
		}
	}
	return false
}
