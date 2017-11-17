package recording

import (
	"math"
	"github.com/v2pro/plz/countlog"
	"encoding/binary"
	"bytes"
)

type TraceHeader []byte

type TraceHeaderKey []byte

type TraceHeaderValue []byte

var TraceHeaderKeyTraceId = TraceHeaderKey("ti")

var TraceHeaderKeySpanId = TraceHeaderKey("si")

func (header TraceHeader) Next() (TraceHeaderKey, TraceHeaderValue, TraceHeader) {
	if len(header) < 2 {
		countlog.Error("event!trace_header.malformed header", "header", header)
		return nil, nil, nil
	}
	keySize := binary.BigEndian.Uint16(header)
	if len(header) < 2+int(keySize)+2 {
		countlog.Error("event!trace_header.malformed header", "header", header)
		return nil, nil, nil
	}
	key := header[2:2+keySize]
	header = header[2+keySize:]
	valueSize := binary.BigEndian.Uint16(header)
	if len(header) < 2+int(valueSize) {
		countlog.Error("event!trace_header.malformed header", "header", header)
		return nil, nil, nil
	}
	value := header[2:2+valueSize]
	header = header[2+valueSize:]
	return TraceHeaderKey(key), TraceHeaderValue(value), header
}

func (header TraceHeader) Set(key TraceHeaderKey, value TraceHeaderValue) TraceHeader {
	if len(key) > math.MaxUint16 || len(value) > math.MaxUint16 {
		countlog.Error("event!trace_header.size overflow", "key", key, "value", value)
		return header
	}
	newHeader := make([]byte, 0, len(header)+len(key)+len(value)+4)
	newHeader = append(newHeader, byte(len(key)>>8), byte(len(key)))
	newHeader = append(newHeader, key...)
	newHeader = append(newHeader, byte(len(value)>>8), byte(len(value)))
	newHeader = append(newHeader, value...)
	var tmpKey TraceHeaderKey
	for len(header) > 0 {
		minSize := 2
		headerSize := len(header)
		if headerSize < minSize {
			countlog.Error("event!trace_header.malformed header", "header", header)
			return newHeader
		}
		keySize := binary.BigEndian.Uint16(header)
		minSize += int(keySize) + 2
		if headerSize < minSize {
			countlog.Error("event!trace_header.malformed header", "header", header)
			return newHeader
		}
		tmpKey = TraceHeaderKey(header[2:2+keySize])
		valueSize := binary.BigEndian.Uint16(header[2+keySize:])
		minSize += int(valueSize)
		if headerSize < minSize {
			countlog.Error("event!trace_header.malformed header", "header", header)
			return newHeader
		}
		if !bytes.Equal(tmpKey, key) {
			newHeader = append(newHeader, header[:minSize]...)
		}
		header = header[minSize:]
	}
	return newHeader
}

func (header TraceHeader) MarshalJSON() ([]byte, error) {
	return EncodeAnyByteArray(header), nil
}
