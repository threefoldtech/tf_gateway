package wg

import (
	"crypto/md5"
	"net"

	"github.com/threefoldtech/zos/pkg/identity"
)

// IPPool allows to generate IPv6 address from a /64 subnet
type IPPool struct {
	kp identity.KeyPair
}

// NewIPPool creates a IP pool allocator
func NewIPPool(kp identity.KeyPair) *IPPool {
	return &IPPool{
		kp: kp,
	}
}

// Get generates a IP that is part of the subnet of the pool by
// hashing the seed of the pool and and ID.
// the hash is then used to generated the last 8 byte of the ipv6 address
func (p *IPPool) Get(id []byte) (net.IP, error) {

	h := md5.New()
	if _, err := h.Write(p.kp.PrivateKey.Seed()); err != nil {
		return nil, err
	}
	if _, err := h.Write(id); err != nil {
		return nil, err
	}

	b := h.Sum(nil)

	ip := net.IP(make([]byte, net.IPv6len))
	copy(ip, p.Subnet().IP)
	for i := 8; i <= 15; i++ {
		ip[i] = b[i-8]
	}

	return ip, nil
}

// Gateway returns the Gateway address of the subnet managed by the pool
func (p *IPPool) Gateway() *net.IPNet {
	gateway := net.IP(make([]byte, net.IPv6len))
	copy(gateway, p.Subnet().IP)
	gateway[len(p.Subnet().IP)-1] = 0x01
	return &net.IPNet{
		IP:   gateway,
		Mask: net.CIDRMask(64, 128),
	}
}

// Subnet generates a /64 IPv6 prefix by hashing the seed
// then using the hash value to generate IP
// the idea is the same seed always generate the same prefix
func (p *IPPool) Subnet() net.IPNet {
	h := md5.Sum(p.kp.PrivateKey.Seed())

	ip := net.IP(make([]byte, net.IPv6len))
	ip[0] = 0xfd
	for i := 1; i <= 7; i++ {
		ip[i] = h[i-1]
	}

	return net.IPNet{
		IP:   ip,
		Mask: net.CIDRMask(64, 128),
	}
}
