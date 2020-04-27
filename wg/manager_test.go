package wg

import (
	"testing"

	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/threefoldtech/zos/pkg/identity"
	"github.com/threefoldtech/zos/pkg/network/namespace"
	"github.com/threefoldtech/zos/pkg/network/wireguard"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func createMgr(t *testing.T) *Mgr {
	kp, err := identity.FromSeed([]byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"))
	require.NoError(t, err)

	mgr, err := New(kp, NewIPPool(kp), "192.168.0.100:1234", "wg-test")
	require.NoError(t, err)
	return mgr
}

func getWgDevice(name, nsName string) (*wgtypes.Device, error) {
	netNS, err := namespace.GetByName(nsName)
	if err != nil {
		return nil, err
	}
	defer netNS.Close()

	var device *wgtypes.Device
	err = netNS.Do(func(_ ns.NetNS) error {
		link, err := wireguard.GetByName(name)
		if err != nil {
			return err
		}
		device, err = link.Device()
		if err != nil {
			return err
		}
		return nil
	})
	return device, err
}

func TestAddRemovePeer(t *testing.T) {
	t.Skip()

	mgr := createMgr(t)
	defer t.Cleanup(func() {
		mgr.Close()
	})

	// 1st Peer
	privateyKeyUser1, err := wgtypes.GeneratePrivateKey()
	require.NoError(t, err)

	cfg, err := mgr.AddPeer("user1", privateyKeyUser1.PublicKey().String())
	require.NoError(t, err)

	assert.Equal(t, []string{"fd5e:ca9b:d3eb:7c0:b51a:eb:8acd:14a5"}, cfg.IPs)
	assert.Equal(t, []Peer{
		{
			Endpoint:   mgr.endpoint,
			AllowedIPs: []string{"::/0"},
			PublicKey:  mgr.priv.PublicKey().String(),
		},
	}, cfg.Peers)

	d, err := getWgDevice(mgr.wgIface, mgr.nsName)
	require.NoError(t, err)
	assert.Equal(t, 1, len(d.Peers))

	// 2nd Peer
	privateyKeyUser2, err := wgtypes.GeneratePrivateKey()
	require.NoError(t, err)

	cfg, err = mgr.AddPeer("user2", privateyKeyUser2.PublicKey().String())
	require.NoError(t, err)

	assert.Equal(t, []string{"fd5e:ca9b:d3eb:7c0:17a2:2514:f5d3:c8c3"}, cfg.IPs)
	assert.Equal(t, []Peer{
		{
			Endpoint:   mgr.endpoint,
			AllowedIPs: []string{"::/0"},
			PublicKey:  mgr.priv.PublicKey().String(),
		},
	}, cfg.Peers)

	d, err = getWgDevice(mgr.wgIface, mgr.nsName)
	require.NoError(t, err)
	assert.Equal(t, 2, len(d.Peers))

	// remove 1st peer
	err = mgr.removePeer(privateyKeyUser1.PublicKey().String())
	require.NoError(t, err)

	d, err = getWgDevice(mgr.wgIface, mgr.nsName)
	require.NoError(t, err)
	assert.Equal(t, 1, len(d.Peers))
}
