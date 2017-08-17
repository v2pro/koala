package main

// #include <stddef.h>
// #include <netinet/in.h>
// #include <sys/types.h>
// #include <sys/socket.h>
// #include "span.h"
import "C"
import "math"

func ch_span_to_bytes(span C.struct_ch_span) []byte {
	buf := (*[math.MaxInt32]byte)(span.Ptr)[:span.Len]
	copyOfBuf := make([]byte, len(buf))
	copy(copyOfBuf, buf)
	return copyOfBuf
}