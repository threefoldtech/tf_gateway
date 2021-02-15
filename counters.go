package tfgateway

import (
	"github.com/threefoldtech/tfexplorer/models/generated/directory"
	"github.com/threefoldtech/zos/pkg/gridtypes"
	"github.com/threefoldtech/zos/pkg/provision/primitives"
)

// Counters tracks the amount of primitives workload deployed and
// the amount of resource unit used
type Counters struct {
	proxy          primitives.AtomicValue
	reverseProxy   primitives.AtomicValue
	subdomain      primitives.AtomicValue
	delegateDomain primitives.AtomicValue

	NRU primitives.AtomicValue // network units
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
func (c *Counters) Increment(r *gridtypes.Workload) error {

	switch r.Type {
	case GatewayProxyType:
		c.proxy.Increment(1)
	case GatewayReverseProxyType:
		c.reverseProxy.Increment(1)
	case GatewaySubdomainType:
		c.subdomain.Increment(1)
	case GatewayDomainDeleateType:
		c.delegateDomain.Increment(1)
	}

	return nil
}

// Decrement is called by the provision.Engine when a reservation has been decommissioned
func (c *Counters) Decrement(r *gridtypes.Workload) error {

	switch r.Type {
	case GatewayProxyType:
		c.proxy.Decrement(1)
	case GatewayReverseProxyType:
		c.reverseProxy.Decrement(1)
	case GatewaySubdomainType:
		c.subdomain.Decrement(1)
	case GatewayDomainDeleateType:
		c.delegateDomain.Decrement(1)
	}

	return nil
}
