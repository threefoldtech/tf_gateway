package tfgateway

import (
	"context"
	"encoding/json"

	"github.com/threefoldtech/zos/pkg/provision"

	"github.com/rs/zerolog/log"
)

// Proxy defines the configuration for a TCP proxy
type Proxy struct {
	Domain  string `json:"domain"`
	Addr    string `json:"addr"`
	Port    uint32 `json:"port"`
	PortTLS uint32 `json:"port_tls"`
}

func (p *Provisioner) proxyProvision(ctx context.Context, r *provision.Reservation) (interface{}, error) {
	log.Info().Str("id", r.ID).Msgf("provision proxy %+v", r)
	data := Proxy{}
	if err := json.Unmarshal(r.Data, &data); err != nil {
		return nil, err
	}
	log.Info().Str("id", r.ID).Msgf("provision proxy %+v", data)

	return nil, p.proxy.AddProxy(r.User, data.Domain, data.Addr, int(data.Port), int(data.PortTLS))
}

func (p *Provisioner) proxyDecomission(ctx context.Context, r *provision.Reservation) error {
	log.Info().Str("id", r.ID).Msg("decomission proxy")

	data := Proxy{}
	if err := json.Unmarshal(r.Data, &data); err != nil {
		return err
	}

	return p.proxy.RemoveProxy(r.User, data.Domain)
}
