package replaying

import (
	"net"
	"github.com/v2pro/koala/st"
	"sync"
	"github.com/v2pro/koala/countlog"
	"syscall"
)

var tmp = map[string]*st.Session{}
var tmpMutex = &sync.Mutex{}

func StoreTmp(inboundAddr net.TCPAddr, session *st.Session) {
	tmpMutex.Lock()
	defer tmpMutex.Unlock()
	tmp[inboundAddr.String()] = session
}

func RetrieveTmp(inboundAddr net.TCPAddr) *st.Session {
	tmpMutex.Lock()
	defer tmpMutex.Unlock()
	key := inboundAddr.String()
	session := tmp[key]
	delete(tmp, key)
	return session
}

func ResolveAddresses(targetIpPort string) (*net.TCPAddr, *net.TCPAddr, error) {
	conn, err := net.Dial("udp", targetIpPort)
	if err != nil {
		countlog.Error("failed to check route", "err", err)
		return nil, nil, err
	}
	localIp := conn.LocalAddr().(*net.UDPAddr).IP.String()
	if err != nil {
		countlog.Error("failed to resolve local tcp addr", "err", err)
		return nil, nil, err
	}
	listener, err := net.Listen("tcp", localIp+":0") // ask for new port
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
	remoteAddr, err := net.ResolveTCPAddr("tcp", ":9000")
	if err != nil {
		countlog.Error("failed to resolve remote tcp addr", "err", err)
		return nil, nil, err
	}
	return localAddr, remoteAddr, nil
}

func BindLocalAddr(socketFD int, targetAddr net.TCPAddr) (*net.TCPAddr, error) {
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
	localTcpAddr, _, err := ResolveAddresses(targetAddr.String())
	if err != nil {
		return nil, err
	}
	err = syscall.Bind(socketFD, &syscall.SockaddrInet4{
		Addr: [4]byte{localTcpAddr.IP[0], localTcpAddr.IP[1], localTcpAddr.IP[2], localTcpAddr.IP[3]},
		Port: localTcpAddr.Port,
	})
	if err != nil {
		return nil, err
	}
	return localTcpAddr, nil
}
