package tfgateway

import (
	"context"
	"encoding/json"

	"github.com/threefoldtech/zos/pkg/gridtypes"

	"github.com/rs/zerolog/log"
)

func (p *Provisioner) proxyProvision(ctx context.Context, wl *gridtypes.Workload) (interface{}, error) {
	log.Info().Stringer("id", wl.ID).Msgf("provision proxy %+v", wl)
	data := GatewayProxy{}
	if err := json.Unmarshal(wl.Data, &data); err != nil {
		return nil, err
	}
	log.Info().Stringer("id", wl.ID).Msgf("provision proxy %+v", data)

	return nil, p.proxy.AddProxy(wl.User.String(), data.Domain, data.Addr, int(data.Port), int(data.PortTLS))
}

func (p *Provisioner) proxyDecomission(ctx context.Context, wl *gridtypes.Workload) error {
	log.Info().Stringer("id", wl.ID).Msg("decomission proxy")

	data := GatewayProxy{}
	if err := json.Unmarshal(wl.Data, &data); err != nil {
		return err
	}

	return p.proxy.RemoveProxy(wl.User.String(), data.Domain)
}
