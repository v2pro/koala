package main

// #include <stddef.h>
// #include <netinet/in.h>
// #include <sys/types.h>
// #include <sys/socket.h>
import "C"
import (
	"reflect"
	"unsafe"
	"github.com/v2pro/koala/ch"
)

var sockaddr_in_type = reflect.TypeOf((*C.struct_sockaddr_in)(nil)).Elem()
var sockaddr_in_sin_family_field = ch.FieldOf(sockaddr_in_type, "sin_family")
var sockaddr_in_sin_port_field = ch.FieldOf(sockaddr_in_type, "sin_port")
var sockaddr_in_sin_addr_field = ch.FieldOf(sockaddr_in_type, "sin_addr")
var in_addr_type = reflect.TypeOf((*C.struct_in_addr)(nil)).Elem()
var in_addr_s_addr_field = ch.FieldOf(in_addr_type, "s_addr")

func init() {
	ch.Dump(sockaddr_in_type)
	ch.Dump(ch.FieldOf(sockaddr_in_type, "sin_addr").Type)
}

func sockaddr_in_sin_family_get(ptr *C.struct_sockaddr_in) uint16 {
	return ch.GetUint16(unsafe.Pointer(ptr), sockaddr_in_sin_family_field)
}

func sockaddr_in_sin_port_get(ptr *C.struct_sockaddr_in) uint16 {
	return ch.GetUint16(unsafe.Pointer(ptr), sockaddr_in_sin_port_field)
}

func sockaddr_in_sin_addr_get(ptr *C.struct_sockaddr_in) uint32 {
	sin_addr := ch.GetPtr(unsafe.Pointer(ptr), sockaddr_in_sin_addr_field)
	return ch.GetUint32(sin_addr, in_addr_s_addr_field)
}
