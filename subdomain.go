package tfgateway

import (
	"context"
	"fmt"
	"net"

	"github.com/threefoldtech/zos/pkg/gridtypes"

	"encoding/json"

	"github.com/rs/zerolog/log"
)

func (p *Provisioner) subDomainProvision(ctx context.Context, wl *gridtypes.Workload) (interface{}, error) {
	data := GatewaySubdomain{}
	if err := json.Unmarshal(wl.Data, &data); err != nil {
		return nil, err
	}
	log.Info().Stringer("id", wl.ID).Msgf("provision Sudbomain %+v", data)
	var ips []net.IP
	for _, ip := range data.IPs {
		parsed := net.ParseIP(ip)
		if parsed == nil {
			return nil, fmt.Errorf("invalid ip '%s'", ip)
		}
		ips = append(ips, parsed)
	}

	return nil, p.dns.AddSubdomain(wl.User.String(), data.Domain, ips)
}

func (p *Provisioner) subDomainDecomission(ctx context.Context, wl *gridtypes.Workload) error {
	data := GatewaySubdomain{}
	if err := json.Unmarshal(wl.Data, &data); err != nil {
		return err
	}
	log.Info().Stringer("id", wl.ID).Msgf("decomission Sudbomain %+v", data)

	var ips []net.IP
	for _, ip := range data.IPs {
		parsed := net.ParseIP(ip)
		if parsed == nil {
			return fmt.Errorf("invalid ip '%s'", ip)
		}
		ips = append(ips, parsed)
	}

	return p.dns.RemoveSubdomain(wl.User.String(), data.Domain, ips)
}
