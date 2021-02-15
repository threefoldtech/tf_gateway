package tfgateway

import (
	"context"
	"encoding/json"

	"github.com/threefoldtech/zos/pkg/gridtypes"

	"github.com/rs/zerolog/log"
)

func (p *Provisioner) domainDeleateProvision(ctx context.Context, wl *gridtypes.Workload) (interface{}, error) {
	data := GatewayDelegate{}
	if err := json.Unmarshal(wl.Data, &data); err != nil {
		return nil, err
	}
	log.Info().Stringer("id", wl.ID).Msgf("provision Delegate %+v", data)

	return nil, p.dns.AddDomainDelagate(p.kp.Identity(), wl.User.String(), data.Domain)
}

func (p *Provisioner) domainDeleateDecomission(ctx context.Context, wl *gridtypes.Workload) error {
	data := GatewayDelegate{}
	if err := json.Unmarshal(wl.Data, &data); err != nil {
		return err
	}
	log.Info().Stringer("id", wl.ID).Msgf("decomission Delegate %+v", data)

	return p.dns.RemoveDomainDelagate(wl.User.String(), data.Domain)
}
