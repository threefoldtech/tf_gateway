package tfgateway

import (
	"context"
	"encoding/json"

	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/zos/pkg/gridtypes"
)

func (p *Provisioner) gateway4To6Provision(ctx context.Context, r *gridtypes.Workload) (interface{}, error) {
	return p.gateway4To6ProvisionImpl(ctx, r)
}

func (p *Provisioner) gateway4To6ProvisionImpl(ctx context.Context, wl *gridtypes.Workload) (Gateway4to6Result, error) {
	data := Gateway4To6{}
	if err := json.Unmarshal(wl.Data, &data); err != nil {
		return Gateway4to6Result{}, err
	}
	log.Info().Stringer("id", wl.ID).Msgf("provision gateway4to6 %+v", data)

	cfg, err := p.wg.AddPeer(wl.User.String(), data.PublicKey)
	if err != nil {
		return Gateway4to6Result{}, err
	}

	return Gateway4to6Result{
		IPs:   cfg.IPs,
		Peers: cfg.Peers,
	}, nil
}

func (p *Provisioner) gateway4To6Decomission(ctx context.Context, wl *gridtypes.Workload) error {
	data := Gateway4To6{}
	if err := json.Unmarshal(wl.Data, &data); err != nil {
		return err
	}
	log.Info().Stringer("id", wl.ID).Msgf("decomission gateway4to6 %+v", data)

	return p.wg.RemovePeer(data.PublicKey)
}
