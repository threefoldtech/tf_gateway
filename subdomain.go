package tfgateway

import (
	"context"
	"net"

	"github.com/threefoldtech/zos/pkg/provision"

	"github.com/polydawn/refmt/json"
	"github.com/rs/zerolog/log"
)

type Subdomain struct {
	Domain string   `json:"domain"`
	IPs    []net.IP `json:"destination"`
}

func (p *Provisioner) subDomainProvision(ctx context.Context, r *provision.Reservation) (interface{}, error) {
	data := Subdomain{}
	if err := json.Unmarshal(r.Data, &data); err != nil {
		return nil, err
	}
	log.Info().Str("id", r.ID).Msgf("provision Sudbomain %+v", data)

	return nil, p.dns.AddSubdomain(r.User, data)
}

func (p *Provisioner) subDomainDecomission(ctx context.Context, r *provision.Reservation) error {
	data := Subdomain{}
	if err := json.Unmarshal(r.Data, &data); err != nil {
		return err
	}
	log.Info().Str("id", r.ID).Msgf("provision Sudbomain %+v", data)

	return p.dns.RemoveSubdomain(r.User, data)
}
