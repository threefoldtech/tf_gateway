package wg

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/containernetworking/plugins/pkg/utils/sysctl"
	"github.com/threefoldtech/zos/pkg/identity"
	"github.com/threefoldtech/zos/pkg/network/namespace"
	"github.com/threefoldtech/zos/pkg/network/wireguard"
	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// Peer is a wireguard peer
type Peer struct {
	PublicKey  string   `json:"public_key"`
	Endpoint   string   `json:"endpoint"`
	AllowedIPs []string `json:"allowed_ips"`
}

// PeerConfig is the wireguard configuration elements returns to the user
// he can use it to create its wg config
type PeerConfig struct {
	IPs   []string `json:"ips"`
	Peers []Peer   `json:"peers"`
}

// Mgr is manager of the Gateway4To6
type Mgr struct {
	kp       identity.KeyPair
	ipAlloc  *IPPool
	priv     wgtypes.Key
	endpoint string
	wgIface  string
	nsName   string
}

// New creates a wireguard manager
func New(kp identity.KeyPair, ipAlloc *IPPool, endpoint string, wgIface string) (*Mgr, error) {

	priv, err := wgtypes.NewKey(kp.PrivateKey.Seed()[:32])
	if err != nil {
		return nil, err
	}

	mgr := &Mgr{
		kp:       kp,
		ipAlloc:  ipAlloc,
		priv:     priv,
		endpoint: endpoint,
		wgIface:  wgIface,
		nsName:   "tfgateway",
	}

	return mgr, mgr.setupWGNamespace()
}

// setupWGNamespace needs to be called when the manager starts
// it setup the network namespace and wireguard interface
func (m *Mgr) setupWGNamespace() error {
	iface, err := wireguard.New(m.wgIface)
	if err != nil {
		return err
	}

	netns, err := namespace.Create(m.nsName)
	if err != nil {
		return err
	}
	defer netns.Close()

	if err := netlink.LinkSetNsFd(iface, int(netns.Fd())); err != nil {
		return err
	}

	err = netns.Do(func(_ ns.NetNS) error {
		if _, err := sysctl.Sysctl("net.ipv6.conf.all.forwarding", "1"); err != nil {
			return err
		}

		lo, err := netlink.LinkByName("lo")
		if err != nil {
			return err
		}

		if err := netlink.LinkSetUp(lo); err != nil {
			return err
		}

		if err := iface.SetAddr(m.ipAlloc.Gateway().String()); err != nil {
			return err
		}

		cl, err := wgctrl.New()
		if err != nil {
			return err
		}
		defer cl.Close()

		_, sp, err := net.SplitHostPort(m.endpoint)
		if err != nil {
			return err
		}
		port, err := strconv.Atoi(sp)
		if err != nil {
			return err
		}

		err = cl.ConfigureDevice(m.wgIface, wgtypes.Config{
			PrivateKey: &m.priv,
			ListenPort: &port,
		})
		if err != nil {
			return err
		}

		return netlink.LinkSetUp(iface)
	})
	return err
}

// Close removes the network namespace and wireguard interface
func (m *Mgr) Close() error {
	netns, err := namespace.GetByName(m.nsName)
	if err != nil {
		return err
	}
	return namespace.Delete(netns)
}

// AddPeer addd a peer identified by pubkey to the wireguard network
// The peer address is allocated from the manager pool and returned to the caller
func (m *Mgr) AddPeer(user, pubKey string) (PeerConfig, error) {
	ip, err := m.ipAlloc.Get([]byte(user))
	if err != nil {
		return PeerConfig{}, err
	}

	cfg := PeerConfig{
		IPs: []string{ip.String()},
		Peers: []Peer{
			{
				PublicKey:  m.priv.PublicKey().String(),
				Endpoint:   m.endpoint,
				AllowedIPs: []string{"::/0"},
			},
		},
	}

	if err := m.appendPeer(Peer{
		PublicKey:  pubKey,
		AllowedIPs: []string{fmt.Sprintf("%s/128", ip.String())},
	}); err != nil {
		return PeerConfig{}, err
	}

	return cfg, nil
}

// RemovePeer removes a peer identified by pubkey from the wireguard network
func (m *Mgr) RemovePeer(pubKey string) error {
	return m.removePeer(pubKey)
}

func (m *Mgr) appendPeer(peer Peer) error {
	netNS, err := namespace.GetByName(m.nsName)
	if err != nil {
		return err
	}
	defer netNS.Close()

	return netNS.Do(func(_ ns.NetNS) error {
		cl, err := wgctrl.New()
		if err != nil {
			return err
		}
		defer cl.Close()

		k, err := wgtypes.ParseKey(peer.PublicKey)
		if err != nil {
			return err
		}

		aips := make([]net.IPNet, len(peer.AllowedIPs))
		for i := range peer.AllowedIPs {
			_, ipnet, err := net.ParseCIDR(peer.AllowedIPs[i])
			if err != nil {
				return err
			}
			aips[i] = *ipnet
		}

		interval := time.Second * 25
		cfg := wgtypes.Config{
			ReplacePeers: false,
			Peers: []wgtypes.PeerConfig{
				{
					PublicKey:                   k,
					AllowedIPs:                  aips,
					PersistentKeepaliveInterval: &interval,
				},
			},
		}

		return cl.ConfigureDevice(m.wgIface, cfg)
	})
}

func (m *Mgr) removePeer(pub string) error {
	netNS, err := namespace.GetByName(m.nsName)
	if err != nil {
		return err
	}
	defer netNS.Close()

	return netNS.Do(func(_ ns.NetNS) error {
		cl, err := wgctrl.New()
		if err != nil {
			return err
		}
		defer cl.Close()

		k, err := wgtypes.ParseKey(pub)
		if err != nil {
			return err
		}

		cfg := wgtypes.Config{
			ReplacePeers: false,
			Peers: []wgtypes.PeerConfig{
				{
					UpdateOnly: true,
					Remove:     true,
					PublicKey:  k,
				},
			},
		}

		return cl.ConfigureDevice(m.wgIface, cfg)
	})
}
