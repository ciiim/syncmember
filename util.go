package syncmember

import (
	"math/rand"
	"net"
	"strconv"
	"strings"
)

func kRamdonNodes(k int, nodes []*Node) []*Node {
	if len(nodes) <= k {
		return nodes
	}
	cloneNodes := make([]*Node, len(nodes))
	//FIXME: ADD MUTEX
	copy(cloneNodes, nodes)
	rand.Shuffle(len(cloneNodes), func(i, j int) {
		cloneNodes[i], cloneNodes[j] = cloneNodes[j], cloneNodes[i]
	})
	pickedNodes := make([]*Node, 0, k)
	for i := 0; i < k; i++ {
		pickedNodes = append(pickedNodes, cloneNodes[i])
	}
	return pickedNodes
}

func ResolveIP(ip string) net.IP {
	return net.ParseIP(ip)
}

func ResolveAddr(addr string) Address {
	if strings.Contains(addr, ":") {
		ip, port, err := net.SplitHostPort(addr)
		if err != nil {
			return Address{}
		}
		portInt, _ := strconv.Atoi(port)
		return Address{
			IP:   ResolveIP(ip),
			Port: portInt,
		}
	}
	return Address{
		IP:   ResolveIP(addr),
		Port: 0,
	}
}

func GetHostIP() net.IP {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil
	}
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP
			}
		}
	}
	return nil
}
