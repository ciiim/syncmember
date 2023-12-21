package syncmember

import (
	"net"
	"strconv"
)

type Address struct {
	IP         net.IP
	addrString string
	Port       int
	Name       string
}

func (a Address) WithName(name string) Address {
	a.Name = name
	return a
}

func (a Address) String() string {
	if a.addrString == "" {
		a.addrString = net.JoinHostPort(a.IP.String(), strconv.Itoa(a.Port))
	}
	return a.addrString
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
