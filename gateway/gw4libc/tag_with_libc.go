// +build koala_libc

package gw4libc

// #cgo CFLAGS: -DKOALA_LIBC -DPTHREAD -DPTHREAD_SINGLETHREADED_TIME
// #cgo CXXFLAGS: -DKOALA_LIBC -DPTHREAD -DPTHREAD_SINGLETHREADED_TIME --std=c++11 -Wno-ignored-attributes
import "C"