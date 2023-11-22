package syncmember

import (
	"net"
	"strconv"
)

type Address struct {
	IP   net.IP
	Port int
	Name string
}

func (a Address) String() string {
	return net.JoinHostPort(a.IP.String(), strconv.Itoa(a.Port))
}

func (a Address) UDPAddr() *net.UDPAddr {
	return &net.UDPAddr{
		IP:   a.IP,
		Port: a.Port,
	}
}

func (a Address) TCPAddr() *net.TCPAddr {
	return &net.TCPAddr{
		IP:   a.IP,
		Port: a.Port,
	}
}
