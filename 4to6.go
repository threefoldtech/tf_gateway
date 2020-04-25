package tfgateway

import (
	"context"
	"encoding/json"

	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/tfgateway/wg"
	"github.com/threefoldtech/zos/pkg/provision"
)

// Gateway4to6 is a primitive that allow client to have a ipv6 gateway
type Gateway4to6 struct {
	PublicKey string `json:"public_key"`
}

// Gateway4to6Result contains the configuration required by the user to
// start its 4to6 gateway tunnel
type Gateway4to6Result struct {
	IPs   []string  `json:"ips"`
	Peers []wg.Peer `json:"peers"`
}

func (p *Provisioner) gateway4To6Provision(ctx context.Context, r *provision.Reservation) (interface{}, error) {
	return p.gateway4To6ProvisionImpl(ctx, r)
}

func (p *Provisioner) gateway4To6ProvisionImpl(ctx context.Context, r *provision.Reservation) (Gateway4to6Result, error) {
	data := Gateway4to6{}
	if err := json.Unmarshal(r.Data, &data); err != nil {
		return Gateway4to6Result{}, err
	}
	log.Info().Str("id", r.ID).Msgf("provision gateway4to6 %+v", data)

	cfg, err := p.wg.AddPeer(r.User, data.PublicKey)
	if err != nil {
		return Gateway4to6Result{}, err
	}

	return Gateway4to6Result{
		IPs:   cfg.IPs,
		Peers: cfg.Peers,
	}, nil
}

func (p *Provisioner) gateway4To6Decomission(ctx context.Context, r *provision.Reservation) error {
	data := Gateway4to6{}
	if err := json.Unmarshal(r.Data, &data); err != nil {
		return err
	}
	log.Info().Str("id", r.ID).Msgf("decomission gateway4to6 %+v", data)

	return p.wg.RemovePeer(data.PublicKey)
}
