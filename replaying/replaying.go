package replaying

import (
	"net"
	"sync"
	"github.com/v2pro/koala/countlog"
	"syscall"
	"os"
)

var isReplaying = false

func init() {
	isReplaying = os.Getenv("KOALA_MODE") == "REPLAYING"
}

func IsReplaying() bool {
	return isReplaying
}

func IsRecording() bool {
	return !isReplaying
}

var tmp = map[string]*ReplayingSession{}
var tmpMutex = &sync.Mutex{}

func StoreTmp(inboundAddr net.TCPAddr, session *ReplayingSession) {
	tmpMutex.Lock()
	defer tmpMutex.Unlock()
	tmp[inboundAddr.String()] = session
}

func RetrieveTmp(inboundAddr net.TCPAddr) *ReplayingSession {
	tmpMutex.Lock()
	defer tmpMutex.Unlock()
	key := inboundAddr.String()
	session := tmp[key]
	delete(tmp, key)
	return session
}

func ResolveAddresses(targetIpPort string) (*net.TCPAddr, *net.TCPAddr, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0") // ask for new port
	if err != nil {
		countlog.Error("failed to resolve local tcp addr port", "err", err)
		return nil, nil, err
	}
	localAddr := listener.Addr().(*net.TCPAddr)
	err = listener.Close()
	if err != nil {
		countlog.Error("failed to close", "err", err)
		return nil, nil, err
	}
	remoteAddr, err := net.ResolveTCPAddr("tcp", targetIpPort)
	if err != nil {
		countlog.Error("failed to resolve remote tcp addr", "err", err)
		return nil, nil, err
	}
	return localAddr, remoteAddr, nil
}

func BindLocalAddr(socketFD int) (*net.TCPAddr, error) {
	localAddr, err := syscall.Getsockname(int(socketFD))
	if err != nil {
		return nil, err
	}
	localInet4Addr := localAddr.(*syscall.SockaddrInet4)
	if localInet4Addr.Port != 0 && localInet4Addr.Addr != [4]byte{} {
		return &net.TCPAddr{
			IP:   localInet4Addr.Addr[:],
			Port: localInet4Addr.Port,
		}, nil
	}
	err = syscall.Bind(socketFD, &syscall.SockaddrInet4{
		Addr: [4]byte{127, 0, 0, 1},
		Port: 0,
	})
	if err != nil {
		return nil, err
	}
	localAddr, err = syscall.Getsockname(int(socketFD))
	if err != nil {
		return nil, err
	}
	localInet4Addr = localAddr.(*syscall.SockaddrInet4)
	return &net.TCPAddr{
		IP:   localInet4Addr.Addr[:],
		Port: localInet4Addr.Port,
	}, nil
}
