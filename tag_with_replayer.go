// +build koala_replayer

package main

// #cgo CFLAGS: -DUSE_LIB_FAKETIME -DPTHREAD -DPTHREAD_SINGLETHREADED_TIME
import "C"