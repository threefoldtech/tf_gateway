package tfgateway

import (
	"github.com/threefoldtech/tfexplorer/models/generated/directory"
	"github.com/threefoldtech/zos/pkg/provision"
	"github.com/threefoldtech/zos/pkg/provision/primitives"
)

// Counters tracks the amount of primitives workload deployed and
// the amount of resource unit used
type Counters struct {
	proxy          primitives.CounterUint64
	reverseProxy   primitives.CounterUint64
	subdomain      primitives.CounterUint64
	delegateDomain primitives.CounterUint64

	NRU primitives.CounterUint64 // network units
}

// CurrentWorkloads return the number of each workloads provisioned on the system
func (c *Counters) CurrentWorkloads() directory.WorkloadAmount {
	return directory.WorkloadAmount{
		Proxy:          uint16(c.proxy.Current()),
		ReverseProxy:   uint16(c.reverseProxy.Current()),
		Subdomain:      uint16(c.subdomain.Current()),
		DelegateDomain: uint16(c.delegateDomain.Current()),
	}
}

// CurrentUnits return the number of each resource units reserved on the system
func (c *Counters) CurrentUnits() directory.ResourceAmount {
	return directory.ResourceAmount{
		// NRU: c.NRU.Current(),
	}
}

// Increment is called by the provision.Engine when a reservation has been provisionned
func (c *Counters) Increment(r *provision.Reservation) error {

	// var (
	// 	u   resourceUnits
	// 	err error
	// )

	switch r.Type {
	case ProxyReservation:
		c.proxy.Increment(1)
		// u, err = processVolume(r)
	case ReverseProxyReservation:
		c.reverseProxy.Increment(1)
		// u, err = processContainer(r)
	case SubDomainReservation:
		c.subdomain.Increment(1)
		// u, err = processZdb(r)
	case DomainDeleateReservation:
		c.delegateDomain.Increment(1)
		// u, err = processKubernetes(r)
	}
	// if err != nil {
	// 	return err
	// }

	// c.CRU.Increment(u.CRU)
	// c.MRU.Increment(u.MRU)
	// c.SRU.Increment(u.SRU)
	// c.HRU.Increment(u.HRU)

	return nil
}

// Decrement is called by the provision.Engine when a reservation has been decommissioned
func (c *Counters) Decrement(r *provision.Reservation) error {

	// var (
	// 	u   resourceUnits
	// 	err error
	// )

	switch r.Type {
	case ProxyReservation:
		c.proxy.Decrement(1)
		// u, err = processVolume(r)
	case ReverseProxyReservation:
		c.reverseProxy.Decrement(1)
		// u, err = processContainer(r)
	case SubDomainReservation:
		c.subdomain.Decrement(1)
		// u, err = processZdb(r)
	case DomainDeleateReservation:
		c.delegateDomain.Decrement(1)
		// u, err = processKubernetes(r)
	}
	// if err != nil {
	// 	return err
	// }
	// c.CRU.Decrement(u.CRU)
	// c.MRU.Decrement(u.MRU)
	// c.SRU.Decrement(u.SRU)
	// c.HRU.Decrement(u.HRU)

	return nil
}

// type resourceUnits struct {
// 	SRU uint64 `json:"sru,omitempty"`
// 	HRU uint64 `json:"hru,omitempty"`
// 	MRU uint64 `json:"mru,omitempty"`
// 	CRU uint64 `json:"cru,omitempty"`
// }
