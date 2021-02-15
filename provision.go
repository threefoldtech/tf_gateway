package tfgateway

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/threefoldtech/zos/pkg/crypto"
	"github.com/threefoldtech/zos/pkg/gridtypes"

	"github.com/threefoldtech/tfgateway/dns"
	"github.com/threefoldtech/tfgateway/proxy"
	"github.com/threefoldtech/tfgateway/wg"
	"github.com/threefoldtech/zos/pkg/identity"
	"github.com/threefoldtech/zos/pkg/provision"
)

// ErrUnsupportedWorkload is return when a workload of a type not supported by
// provisiond is received from the explorer

// ReservationType enum list all the supported primitives but the tfgateway

type provisionFn func(ctx context.Context, wl *gridtypes.Workload) (interface{}, error)
type decommissionFn func(ctx context.Context, wl *gridtypes.Workload) error

// Provisioner hold all the logic responsible to provision and decomission
// the different primitives workloads defined by this package
type Provisioner struct {
	kp identity.KeyPair

	proxy *proxy.Mgr
	dns   *dns.Mgr
	wg    *wg.Mgr

	Provisioners    map[gridtypes.WorkloadType]provisionFn
	Decommissioners map[gridtypes.WorkloadType]decommissionFn
}

// NewProvisioner creates a new 0-OS provisioner
func NewProvisioner(proxy *proxy.Mgr, dns *dns.Mgr, wg *wg.Mgr, kp identity.KeyPair) provision.Provisioner {
	p := &Provisioner{
		kp:    kp,
		proxy: proxy,
		dns:   dns,
		wg:    wg,
	}
	provisioners := map[gridtypes.WorkloadType]provision.DeployFunction{
		GatewayProxyType:          p.proxyProvision,
		GatewayReverseProxyType:   p.reverseProxyProvision,
		GatewaySubdomainType:      p.subDomainProvision,
		GatewayDomainDelegateType: p.domainDeleateProvision,
	}

	decommissioners := map[gridtypes.WorkloadType]provision.RemoveFunction{
		GatewayProxyType:          p.proxyDecomission,
		GatewayReverseProxyType:   p.reverseProxyDecomission,
		GatewaySubdomainType:      p.subDomainDecomission,
		GatewayDomainDelegateType: p.domainDeleateDecomission,
	}

	if wg != nil {
		provisioners[Gateway4To6Type] = p.gateway4To6Provision
		decommissioners[Gateway4To6Type] = p.gateway4To6Decomission
	}

	return provision.NewMapProvisioner(provisioners, decommissioners)
}

func (p *Provisioner) decryptSecret(ctx context.Context, user gridtypes.ID, secret string, version int) (string, error) {
	if len(secret) == 0 {
		return "", nil
	}

	engine := provision.GetEngine(ctx)

	bytes, err := hex.DecodeString(secret)
	if err != nil {
		return "", err
	}

	var (
		out []byte
	)
	// now only one version is supported
	switch version {
	default:
		userPubKey := engine.Users().GetKey(user)
		if userPubKey == nil {
			return "", fmt.Errorf("failed to retrieve user %s public key", user)
		}
		out, err = crypto.DecryptECDH(bytes, p.kp.PrivateKey, userPubKey)
	}

	return string(out), err
}
