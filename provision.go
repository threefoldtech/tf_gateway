package tfgateway

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/threefoldtech/zos/pkg/crypto"

	"github.com/threefoldtech/tfexplorer/models/generated/workloads"
	"github.com/threefoldtech/tfexplorer/schema"
	"github.com/threefoldtech/tfgateway/dns"
	"github.com/threefoldtech/tfgateway/proxy"
	"github.com/threefoldtech/tfgateway/wg"
	"github.com/threefoldtech/zos/pkg/identity"
	"github.com/threefoldtech/zos/pkg/provision"
)

// ReservationType enum list all the supported primitives but the tfgateway
var (
	ProxyReservation         = provision.ReservationType(workloads.WorkloadTypeProxy.String())
	ReverseProxyReservation  = provision.ReservationType(workloads.WorkloadTypeReverseProxy.String())
	SubDomainReservation     = provision.ReservationType(workloads.WorkloadTypeSubDomain.String())
	DomainDeleateReservation = provision.ReservationType(workloads.WorkloadTypeDomainDelegate.String())
	Gateway4To6Reservation   = provision.ReservationType(workloads.WorkloadTypeGateway4To6.String())
)

// ProvisionOrder is used to sort the workload type
// in the right order for provision engine
var ProvisionOrder = map[provision.ReservationType]int{
	DomainDeleateReservation: 0,
	SubDomainReservation:     1,
	ProxyReservation:         2,
	ReverseProxyReservation:  3,
	Gateway4To6Reservation:   4,
}

// Provisioner hold all the logic responsible to provision and decomission
// the different primitives workloads defined by this package
type Provisioner struct {
	kp identity.KeyPair

	proxy *proxy.Mgr
	dns   *dns.Mgr
	wg    *wg.Mgr

	Provisioners    map[provision.ReservationType]provision.ProvisionerFunc
	Decommissioners map[provision.ReservationType]provision.DecomissionerFunc
}

// NewProvisioner creates a new 0-OS provisioner
func NewProvisioner(proxy *proxy.Mgr, dns *dns.Mgr, wg *wg.Mgr, kp identity.KeyPair) *Provisioner {
	p := &Provisioner{
		kp:    kp,
		proxy: proxy,
		dns:   dns,
		wg:    wg,
	}
	p.Provisioners = map[provision.ReservationType]provision.ProvisionerFunc{
		ProxyReservation:         p.proxyProvision,
		ReverseProxyReservation:  p.reverseProxyProvision,
		SubDomainReservation:     p.subDomainProvision,
		DomainDeleateReservation: p.domainDeleateProvision,
	}
	p.Decommissioners = map[provision.ReservationType]provision.DecomissionerFunc{
		ProxyReservation:         p.proxyDecomission,
		ReverseProxyReservation:  p.reverseProxyDecomission,
		SubDomainReservation:     p.subDomainDecomission,
		DomainDeleateReservation: p.domainDeleateDecomission,
	}

	if wg != nil {
		p.Provisioners[Gateway4To6Reservation] = p.gateway4To6Provision
		p.Decommissioners[Gateway4To6Reservation] = p.gateway4To6Decomission
	}

	return p
}

func (p *Provisioner) decrypt(msg string) (string, error) {
	if len(msg) == 0 {
		return "", nil
	}

	bytes, err := hex.DecodeString(msg)
	if err != nil {
		return "", err
	}

	out, err := crypto.Decrypt(bytes, p.kp.PrivateKey)
	return string(out), err

}

func proxyConverter(w workloads.GatewayProxy) (Proxy, string, error) {
	return Proxy{
		Domain:  w.Domain,
		Addr:    w.Addr,
		Port:    w.Port,
		PortTLS: w.PortTLS,
	}, w.NodeId, nil
}

func reserveproxyConverter(w workloads.GatewayReverseProxy) (ReverseProxy, string, error) {
	return ReverseProxy{
		Domain: w.Domain,
		Secret: w.Secret,
	}, w.NodeId, nil
}

func subdomainConverter(w workloads.GatewaySubdomain) (Subdomain, string, error) {
	s := Subdomain{
		Domain: w.Domain,
		IPs:    make([]net.IP, len(w.IPs)),
	}
	for i := range w.IPs {
		s.IPs[i] = net.ParseIP(w.IPs[i])
	}

	return s, w.NodeId, nil
}

func delegateConverter(w workloads.GatewayDelegate) (Delegate, string, error) {
	return Delegate{
		Domain: w.Domain,
	}, w.NodeId, nil
}

func gateway4To6Converter(w workloads.Gateway4To6) (Gateway4to6, string, error) {
	return Gateway4to6{
		PublicKey: w.PublicKey,
	}, w.NodeId, nil
}

// WorkloadToProvisionType TfgridReservationWorkload1 to provision.Reservation
func WorkloadToProvisionType(w workloads.ReservationWorkload) (*provision.Reservation, error) {

	reservation := &provision.Reservation{
		ID:        w.WorkloadId,
		User:      w.User,
		Type:      provision.ReservationType(w.Type.String()),
		Created:   w.Created.Time,
		Duration:  time.Duration(w.Duration) * time.Second,
		Signature: []byte(w.Signature),
		ToDelete:  w.ToDelete,
	}

	var (
		data interface{}
		err  error
	)

	switch tmp := w.Content.(type) {
	case workloads.GatewayProxy:
		data, reservation.NodeID, err = proxyConverter(tmp)
		if err != nil {
			return nil, err
		}
	case workloads.GatewayReverseProxy:
		data, reservation.NodeID, err = reserveproxyConverter(tmp)
		if err != nil {
			return nil, err
		}
	case workloads.GatewaySubdomain:
		data, reservation.NodeID, err = subdomainConverter(tmp)
		if err != nil {
			return nil, err
		}
	case workloads.GatewayDelegate:
		data, reservation.NodeID, err = delegateConverter(tmp)
		if err != nil {
			return nil, err
		}
	case workloads.Gateway4To6:
		data, reservation.NodeID, err = gateway4To6Converter(tmp)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown workload type (%s) (%T)", w.Type.String(), tmp)
	}

	reservation.Data, err = json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return reservation, nil
}

// ResultToSchemaType converts result to schema type
func ResultToSchemaType(r provision.Result) (*workloads.Result, error) {

	var rType workloads.WorkloadTypeEnum
	switch r.Type {
	case ProxyReservation:
		rType = workloads.WorkloadTypeProxy
	case ReverseProxyReservation:
		rType = workloads.WorkloadTypeReverseProxy
	case SubDomainReservation:
		rType = workloads.WorkloadTypeSubDomain
	case DomainDeleateReservation:
		rType = workloads.WorkloadTypeDomainDelegate
	case Gateway4To6Reservation:
		rType = workloads.WorkloadTypeGateway4To6
	default:
		return nil, fmt.Errorf("unknown reservation type: %s", r.Type)
	}

	result := workloads.Result{
		Category:   rType,
		WorkloadId: r.ID,
		DataJson:   r.Data,
		Signature:  r.Signature,
		State:      workloads.ResultStateEnum(r.State),
		Message:    r.Error,
		Epoch:      schema.Date{Time: r.Created},
	}

	return &result, nil
}
