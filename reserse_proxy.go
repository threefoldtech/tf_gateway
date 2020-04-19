package tfgateway

import (
	"context"

	"github.com/threefoldtech/zos/pkg/provision"

	"github.com/polydawn/refmt/json"
	"github.com/rs/zerolog/log"
)

// ReverseProxy define a reverse tunnel TCP proxy
type ReverseProxy struct {
	Domain string `json:"domain"`
	Secret string `json:"secret"`
}

func (p *Provisioner) reverseProxyProvision(ctx context.Context, r *provision.Reservation) (interface{}, error) {
	data := ReverseProxy{}
	if err := json.Unmarshal(r.Data, &data); err != nil {
		return nil, err
	}
	log.Info().Str("id", r.ID).Msgf("provision proxy %+v", data)

	return nil, p.proxy.AddReverseProxy(r.User, data.Domain, data.Secret)
}

func (p *Provisioner) reverseProxyDecomission(ctx context.Context, r *provision.Reservation) error {
	data := ReverseProxy{}
	if err := json.Unmarshal(r.Data, &data); err != nil {
		return err
	}
	log.Info().Str("id", r.ID).Msgf("decomission proxy %+v", data)

	return p.proxy.RemoveReverseProxy(r.User, data.Domain)
}
