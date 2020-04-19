package tfgateway

import (
	"context"
	"encoding/json"

	"github.com/threefoldtech/zos/pkg/provision"

	"github.com/rs/zerolog/log"
)

type Proxy struct {
	Domain  string `json:"domain"`
	Addr    string `json:"addr"`
	Port    uint32 `json:"port"`
	PortTLS uint32 `json:"port_tls"`
}

type ProxyResult struct{}

func (p *Provisioner) proxyProvision(ctx context.Context, r *provision.Reservation) (interface{}, error) {
	return p.proxyProvisionImpl(ctx, r)
}

func (p *Provisioner) proxyProvisionImpl(ctx context.Context, r *provision.Reservation) (result ProxyResult, err error) {
	log.Info().Str("id", r.ID).Msgf("provision proxy %+v", r)
	data := Proxy{}
	if err := json.Unmarshal(r.Data, &data); err != nil {
		return result, err
	}
	log.Info().Str("id", r.ID).Msgf("provision proxy %+v", data)

	return result, p.proxy.AddProxy(r.User, data)
}

func (p *Provisioner) proxyDecomission(ctx context.Context, r *provision.Reservation) error {
	log.Info().Str("id", r.ID).Msg("decomission proxy")

	data := Proxy{}
	if err := json.Unmarshal(r.Data, &data); err != nil {
		return err
	}

	return p.proxy.RemoveProxy(r.User, data)
}
