package syncmember

import (
	"bytes"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/ciiim/syncmember/codec"
	"github.com/ciiim/syncmember/transport"
	"github.com/google/btree"
)

func SendPacket(transport *transport.UDPTransport, packet *Packet) error {
	b, err := codec.Marshal(packet)
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(b)
	return transport.SendRaw(buf, packet.To.UDPAddr())
}

func EqualAddress(a, b Address) bool {
	return a.IP.Equal(b.IP) && a.Port == b.Port
}

func kRamdonNodes(k int, nodes []*Node, exclude func(*Node) bool) []*Node {
	cloneNodes := make([]*Node, len(nodes))
	copy(cloneNodes, nodes)
	rand.Shuffle(len(cloneNodes), func(i, j int) {
		cloneNodes[i], cloneNodes[j] = cloneNodes[j], cloneNodes[i]
	})
	pickedNodes := make([]*Node, 0, k)
	if len(cloneNodes) < k {
		k = len(cloneNodes)
	}
	pickedNums := 0
	for i := 0; i < k; i++ {
		if pickedNums >= k {
			break
		}
		if exclude(cloneNodes[i]) {
			continue
		}
		pickedNodes = append(pickedNodes, cloneNodes[i])
		pickedNums++
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
	ips, err := net.LookupIP(addr)
	if err != nil {
		return Address{}
	}
	//只取第一个
	ip := ips[0]
	return Address{
		IP: ip,
		// FIXME:暂时没有解决方案-Port
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

func randSeq() uint64 {
	return rand.Uint64()
}

// lay out btree to disk
func layoutKVTree(tree *btree.BTree, mu *sync.Mutex) {
	mu.Lock()
	defer mu.Unlock()
}
