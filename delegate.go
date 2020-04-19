package tfgateway

import (
	"context"
	"encoding/json"

	"github.com/threefoldtech/zos/pkg/provision"

	"github.com/rs/zerolog/log"
)

type Delegate struct {
	Domain string `json:"domain"`
}

func (p *Provisioner) domainDeleateProvision(ctx context.Context, r *provision.Reservation) (interface{}, error) {
	data := Delegate{}
	if err := json.Unmarshal(r.Data, &data); err != nil {
		return nil, err
	}
	log.Info().Str("id", r.ID).Msgf("provision Delegate %+v", data)

	return nil, p.dns.AddDomainDelagate(r.User, data)
}

func (p *Provisioner) domainDeleateDecomission(ctx context.Context, r *provision.Reservation) error {
	data := Delegate{}
	if err := json.Unmarshal(r.Data, &data); err != nil {
		return err
	}
	log.Info().Str("id", r.ID).Msgf("decomission Delegate %+v", data)

	return p.dns.RemoveDomainDelagate(r.User, data)
}
