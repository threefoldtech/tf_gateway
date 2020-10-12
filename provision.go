package tfgateway

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net"
	"strconv"

	"github.com/threefoldtech/zos/pkg/crypto"

	"github.com/threefoldtech/tfexplorer/client"

	"github.com/threefoldtech/tfexplorer/models/generated/workloads"
	"github.com/threefoldtech/tfexplorer/schema"
	"github.com/threefoldtech/tfgateway/dns"
	"github.com/threefoldtech/tfgateway/proxy"
	"github.com/threefoldtech/tfgateway/wg"
	"github.com/threefoldtech/zos/pkg/identity"
	"github.com/threefoldtech/zos/pkg/provision"
)

// ErrUnsupportedWorkload is return when a workload of a type not supported by
// provisiond is received from the explorer
var ErrUnsupportedWorkload = errors.New("workload type not supported")

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

	explorer *client.Client

	Provisioners    map[provision.ReservationType]provision.ProvisionerFunc
	Decommissioners map[provision.ReservationType]provision.DecomissionerFunc
}

// NewProvisioner creates a new 0-OS provisioner
func NewProvisioner(proxy *proxy.Mgr, dns *dns.Mgr, wg *wg.Mgr, kp identity.KeyPair, explorer *client.Client) *Provisioner {
	p := &Provisioner{
		kp:       kp,
		proxy:    proxy,
		dns:      dns,
		wg:       wg,
		explorer: explorer,
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

func (p *Provisioner) decrypt(msg, userID string, reservationVersion int) (string, error) {
	if len(msg) == 0 {
		return "", nil
	}

	bytes, err := hex.DecodeString(msg)
	if err != nil {
		return "", err
	}

	var (
		out        []byte
		userPubKey ed25519.PublicKey
	)

	switch reservationVersion {
	case 0:
		out, err = crypto.Decrypt(bytes, p.kp.PrivateKey)
	case 1:
		userPubKey, err = p.fetchUserPublicKey(userID)
		if err != nil {
			return "", fmt.Errorf("failed to retrieve user %s public key: %w", userID, err)
		}
		out, err = crypto.DecryptECDH(bytes, p.kp.PrivateKey, userPubKey)
	}

	return string(out), err

}

func (p *Provisioner) fetchUserPublicKey(userID string) (ed25519.PublicKey, error) {
	iid, err := strconv.Atoi(userID)
	if err != nil {
		return nil, err
	}

	user, err := p.explorer.Phonebook.Get(schema.ID(iid))
	if err != nil {
		return nil, err
	}

	b, err := hex.DecodeString(user.Pubkey)
	if err != nil {
		return nil, err
	}

	return ed25519.PublicKey(b), nil
}

func proxyConverter(w workloads.Workloader) (Proxy, string, error) {
	p, ok := w.(*workloads.GatewayProxy)
	if !ok {
		return Proxy{}, "", fmt.Errorf("failed to convert proxy workload, wrong format")
	}
	return Proxy{
		Domain:  p.Domain,
		Addr:    p.Addr,
		Port:    p.Port,
		PortTLS: p.PortTLS,
	}, p.NodeId, nil
}

func reserveproxyConverter(w workloads.Workloader) (ReverseProxy, string, error) {
	p, ok := w.(*workloads.GatewayReverseProxy)
	if !ok {
		return ReverseProxy{}, "", fmt.Errorf("failed to convert reverse proxy workload, wrong format")
	}

	return ReverseProxy{
		Domain: p.Domain,
		Secret: p.Secret,
	}, p.NodeId, nil
}

func subdomainConverter(w workloads.Workloader) (Subdomain, string, error) {
	s, ok := w.(*workloads.GatewaySubdomain)
	if !ok {
		return Subdomain{}, "", fmt.Errorf("failed to convert subdomain workload, wrong format")
	}

	subdomain := Subdomain{
		Domain: s.Domain,
		IPs:    make([]net.IP, len(s.IPs)),
	}
	for i := range s.IPs {
		subdomain.IPs[i] = net.ParseIP(s.IPs[i])
	}

	return subdomain, s.NodeId, nil
}

func delegateConverter(w workloads.Workloader) (Delegate, string, error) {
	d, ok := w.(*workloads.GatewayDelegate)
	if !ok {
		return Delegate{}, "", fmt.Errorf("failed to convert delegate domain workload, wrong format")
	}

	return Delegate{
		Domain: d.Domain,
	}, d.NodeId, nil
}

func gateway4To6Converter(w workloads.Workloader) (Gateway4to6, string, error) {
	t, ok := w.(*workloads.Gateway4To6)
	if !ok {
		return Gateway4to6{}, "", fmt.Errorf("failed to convert gateway 4to6 workload, wrong format")
	}

	return Gateway4to6{
		PublicKey: t.PublicKey,
	}, t.NodeId, nil
}

// WorkloadToProvisionType TfgridReservationWorkload1 to provision.Reservation
func WorkloadToProvisionType(w workloads.Workloader) (*provision.Reservation, error) {

	reservation := &provision.Reservation{
		ID:        fmt.Sprintf("%d-%d", w.GetID(), w.WorkloadID()),
		User:      fmt.Sprintf("%d", w.GetCustomerTid()),
		Type:      provision.ReservationType(w.GetWorkloadType().String()),
		Created:   w.GetEpoch().Time,
		Duration:  math.MaxInt64,
		Signature: []byte(w.GetCustomerSignature()),
		ToDelete:  w.GetNextAction() == workloads.NextActionDelete,
		Reference: w.GetReference(),
		Result:    resultFromSchemaType(w.GetResult()),
	}

	// to ensure old reservation workload that are already running
	// keeps running as it is, we use the reference as new workload ID
	if reservation.Reference != "" {
		reservation.ID = reservation.Reference
	}

	var (
		data interface{}
		err  error
	)

	switch w.GetWorkloadType() {
	case workloads.WorkloadTypeProxy:
		data, reservation.NodeID, err = proxyConverter(w)
		if err != nil {
			return nil, err
		}
	case workloads.WorkloadTypeReverseProxy:
		data, reservation.NodeID, err = reserveproxyConverter(w)
		if err != nil {
			return nil, err
		}
	case workloads.WorkloadTypeSubDomain:
		data, reservation.NodeID, err = subdomainConverter(w)
		if err != nil {
			return nil, err
		}
	case workloads.WorkloadTypeDomainDelegate:
		data, reservation.NodeID, err = delegateConverter(w)
		if err != nil {
			return nil, err
		}
	case workloads.WorkloadTypeGateway4To6:
		data, reservation.NodeID, err = gateway4To6Converter(w)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("%w (%s) (%T)", ErrUnsupportedWorkload, w.GetWorkloadType().String(), w)
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

func resultFromSchemaType(r workloads.Result) provision.Result {

	result := provision.Result{
		Type:      provision.ReservationType(r.Category.String()),
		Created:   r.Epoch.Time,
		Data:      r.DataJson,
		Error:     r.Message,
		ID:        r.WorkloadId,
		State:     provision.ResultState(r.State),
		Signature: r.Signature,
	}

	return result
}
