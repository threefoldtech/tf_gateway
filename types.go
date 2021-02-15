package tfgateway

import (
	"fmt"
	"io"

	"github.com/threefoldtech/tfgateway/wg"
	"github.com/threefoldtech/zos/pkg/gridtypes"
	"github.com/threefoldtech/zos/pkg/provision"
)

const (
	//GatewayProxyType type
	GatewayProxyType gridtypes.WorkloadType = "gateway-proxy"
	//GatewayReverseProxyType type
	GatewayReverseProxyType gridtypes.WorkloadType = "gateway-reverse-proxy"
	// GatewaySubdomainType type
	GatewaySubdomainType gridtypes.WorkloadType = "gateway-subdomain"
	// GatewayDomainDelegateType type
	GatewayDomainDelegateType gridtypes.WorkloadType = "gateway-domain-delegate"
	// Gateway4To6Type type
	Gateway4To6Type gridtypes.WorkloadType = "gateway-4to6"
)

// Order is preferred order of workload type deployment on boot
var Order = provision.WithStartupOrder(
	GatewayDomainDelegateType,
	GatewaySubdomainType,
	GatewayProxyType,
	GatewayReverseProxyType,
	Gateway4To6Type,
)

func init() {
	// register types for engine
	gridtypes.RegisterType(GatewayProxyType, GatewayProxy{})
	gridtypes.RegisterType(GatewayReverseProxyType, GatewayReverseProxy{})
	gridtypes.RegisterType(GatewaySubdomainType, GatewaySubdomain{})
	gridtypes.RegisterType(GatewayDomainDelegateType, GatewayDelegate{})
	gridtypes.RegisterType(Gateway4To6Type, Gateway4To6{})
}

var _ gridtypes.WorkloadData = GatewayProxy{}

// GatewayProxy type
type GatewayProxy struct {
	Domain  string `bson:"domain" json:"domain"`
	Addr    string `bson:"addr" json:"addr"`
	Port    uint32 `bson:"port" json:"port"`
	PortTLS uint32 `bson:"port_tls" json:"port_tls"`
}

// Valid implements WorkloadData
func (g GatewayProxy) Valid() error {
	//TODO:
	return nil
}

// Challenge implements WokloadData
func (g GatewayProxy) Challenge(w io.Writer) error {

	if _, err := fmt.Fprintf(w, "%s", g.Domain); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "%s", g.Addr); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "%d", g.Port); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "%d", g.PortTLS); err != nil {
		return err
	}
	return nil
}

var _ gridtypes.WorkloadData = GatewayReverseProxy{}

// GatewayReverseProxy type
type GatewayReverseProxy struct {
	Domain string `bson:"domain" json:"domain"`
	Secret string `bson:"secret" json:"secret"`
}

// Valid implementation
func (g GatewayReverseProxy) Valid() error {
	if g.Domain == "" {
		return fmt.Errorf("domain cannot be empty")
	}

	if g.Secret == "" {
		return fmt.Errorf("secret cannot be empty")
	}

	return nil
}

//Challenge implementation
func (g GatewayReverseProxy) Challenge(b io.Writer) error {
	if _, err := fmt.Fprintf(b, "%s", g.Domain); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(b, "%s", g.Secret); err != nil {
		return err
	}

	return nil
}

var _ gridtypes.WorkloadData = GatewaySubdomain{}

// GatewaySubdomain  type
type GatewaySubdomain struct {
	Domain string   `bson:"domain" json:"domain"`
	IPs    []string `bson:"ips" json:"ips"`
}

//Valid implementation
func (s GatewaySubdomain) Valid() error {
	//TODO
	return nil
}

// Challenge implementation
func (s GatewaySubdomain) Challenge(b io.Writer) error {
	if _, err := fmt.Fprintf(b, "%s", s.Domain); err != nil {
		return err
	}
	for _, ip := range s.IPs {
		if _, err := fmt.Fprintf(b, "%s", ip); err != nil {
			return err
		}
	}

	return nil
}

var _ gridtypes.WorkloadData = GatewayDelegate{}

// GatewayDelegate type
type GatewayDelegate struct {
	Domain string `bson:"domain" json:"domain"`
}

// Valid implementation
func (d GatewayDelegate) Valid() error {
	// TODO
	return nil
}

// Challenge implementation
func (d GatewayDelegate) Challenge(w io.Writer) error {
	if _, err := fmt.Fprintf(w, "%s", d.Domain); err != nil {
		return err
	}

	return nil
}

var _ gridtypes.WorkloadData = Gateway4To6{}

// Gateway4To6 type
type Gateway4To6 struct {
	PublicKey string `bson:"public_key" json:"public_key"`
}

// Valid implementation
func (g Gateway4To6) Valid() error {
	return nil
}

// Challenge implementation
func (g Gateway4To6) Challenge(b io.Writer) error {
	if _, err := fmt.Fprintf(b, "%s", g.PublicKey); err != nil {
		return err
	}

	return nil
}

// Gateway4to6Result contains the configuration required by the user to
// start its 4to6 gateway tunnel
type Gateway4to6Result struct {
	IPs   []string  `json:"ips"`
	Peers []wg.Peer `json:"peers"`
}
