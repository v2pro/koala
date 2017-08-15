// +build !koala_go

package internal

import "syscall"

func SetCurrentGoRoutineIsKoala() {
}

func GetCurrentGoRoutineIsKoala() bool {
	return false
}

func GetCurrentGoRoutineId() int64 {
	return 0
}

func RegisterOnConnect(callback func(fd int, sa syscall.Sockaddr)) {
}

func RegisterOnAccept(callback func(serverSocketFD int, clientSocketFD int, sa syscall.Sockaddr)) {
}

func RegisterOnBind(callback func(fd int, sa syscall.Sockaddr)) {
}

func RegisterOnRecv(callback func(fd int, span []byte)) {
}

func RegisterOnSend(callback func(fd int, span []byte)) {

}