package tfgateway

import (
	"context"
	"fmt"
	"strings"

	"github.com/threefoldtech/zos/pkg/gridtypes"

	"encoding/json"

	"github.com/rs/zerolog/log"
)

func (p *Provisioner) reverseProxyProvision(ctx context.Context, wl *gridtypes.Workload) (interface{}, error) {
	data := GatewayReverseProxy{}
	if err := json.Unmarshal(wl.Data, &data); err != nil {
		return nil, err
	}
	log.Info().Stringer("id", wl.ID).Msgf("provision proxy %+v", data)

	secret, err := p.decryptSecret(ctx, wl.User, data.Secret, wl.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt secret: %w", err)
	}

	if !strings.HasPrefix(secret, fmt.Sprintf("%s:", wl.User.String())) {
		return nil, fmt.Errorf("secret must follow the format 'threebotID:random'")
	}

	return nil, p.proxy.AddReverseProxy(wl.User.String(), data.Domain, secret)
}

func (p *Provisioner) reverseProxyDecomission(ctx context.Context, wl *gridtypes.Workload) error {
	data := GatewayReverseProxy{}
	if err := json.Unmarshal(wl.Data, &data); err != nil {
		return err
	}
	log.Info().Stringer("id", wl.ID).Msgf("decomission proxy %+v", data)

	return p.proxy.RemoveReverseProxy(wl.User.String(), data.Domain)
}
