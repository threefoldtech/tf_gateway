package tfgateway

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/tfexplorer/models/generated/workloads"
	"github.com/threefoldtech/tfgateway/dns"
	"github.com/threefoldtech/tfgateway/proxy"
	"github.com/threefoldtech/zos/pkg/provision"
)

// ReservationType enum list all the supported primitives but the tfgateway
var (
	ProxyReservation         provision.ReservationType = "proxy"
	ReverseProxyReservation  provision.ReservationType = "reserve-proxy"
	SubDomainReservation     provision.ReservationType = "subdomain"
	DomainDeleateReservation provision.ReservationType = "domain-delegate"
)

type Provisioner struct {
	proxy *proxy.ProxyMgr
	dns   *dns.DNSMgr

	Provisioners    map[provision.ReservationType]provision.ProvisionerFunc
	Decommissioners map[provision.ReservationType]provision.DecomissionerFunc
}

func NewProvisioner(proxy *proxy.ProxyMgr, dns *dns.DNSMgr) *Provisioner {
	p := &Provisioner{
		proxy: proxy,
		dns:   dns,
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
	return p
}

func proxyConverter(w workloads.GatewayProxy) (Proxy, string, error) {
	return Proxy{
		Domain:  w.Domain,
		Addr:    w.Addr,
		Port:    w.Port,
		PortTLS: w.PortTLS,
	}, w.NodeId, nil
}

func reserveproxyConverter(w workloads.GatewayReserveProxy) (ReverseProxy, string, error) {
	return ReverseProxy{
		Domain: w.Domain,
		Secret: w.Secret,
	}, w.NodeId, nil
}

func SubdomainConverter(w workloads.GatewaySubdomain) (Subdomain, string, error) {
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

func WorkloadToProvisionType(w workloads.ReservationWorkload) (*provision.Reservation, error) {
	log.Info().Msgf("tfgatway converter %+v", w)
	reservation := &provision.Reservation{
		ID:        w.WorkloadId,
		User:      w.User,
		Type:      provision.ReservationType(w.Type.String()),
		Created:   w.Created.Time,
		Duration:  time.Duration(w.Duration) * time.Second,
		Signature: []byte(w.Signature),
		// Data:      w.Content,
		ToDelete: w.ToDelete,
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
	case workloads.GatewayReserveProxy:
		data, reservation.NodeID, err = reserveproxyConverter(tmp)
		if err != nil {
			return nil, err
		}
	case workloads.GatewaySubdomain:
		data, reservation.NodeID, err = SubdomainConverter(tmp)
		if err != nil {
			return nil, err
		}
	case workloads.GatewayDelegate:
		data, reservation.NodeID, err = delegateConverter(tmp)
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
