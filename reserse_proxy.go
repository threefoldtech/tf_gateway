package tfgateway

import (
	"context"
	"fmt"
	"strings"

	"github.com/threefoldtech/zos/pkg/provision"

	"encoding/json"

	"github.com/rs/zerolog/log"
)

// ReverseProxy define a reverse tunnel TCP proxy
type ReverseProxy struct {
	Domain string `json:"domain"`
	Secret string `json:"secret"`
}

func (r ReverseProxy) validate(user string) error {
	if r.Domain == "" {
		return fmt.Errorf("domain cannot be empty")
	}

	if r.Secret == "" {
		return fmt.Errorf("secret cannot be empty")
	}

	if !strings.HasPrefix(r.Secret, fmt.Sprintf("%s:", user)) {
		return fmt.Errorf("secret must follow the format 'threebotID:random'")
	}

	return nil
}

func (p *Provisioner) reverseProxyProvision(ctx context.Context, r *provision.Reservation) (interface{}, error) {
	data := ReverseProxy{}
	if err := json.Unmarshal(r.Data, &data); err != nil {
		return nil, err
	}
	log.Info().Str("id", r.ID).Msgf("provision proxy %+v", data)

	var err error
	data.Secret, err = p.decrypt(data.Secret)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt secret: %w", err)
	}

	if err := data.validate(r.User); err != nil {
		return nil, err
	}

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
