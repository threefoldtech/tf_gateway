package main

import (
	"context"
	"os"
	"time"

	"github.com/cenkalti/backoff/v3"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/zos/pkg/farmer"
	"github.com/threefoldtech/zos/pkg/geoip"
)

func registration(ctx context.Context, nodeID, substrate string, farmerID uint32) error {
	loc, err := geoip.Fetch()
	if err != nil {
		log.Fatal().Err(err).Msg("fetch location")
	}

	fm, err := farmer.NewClientFromSubstrate(substrate, farmerID)
	if err != nil {
		return errors.Wrap(err, "failed to create farmer client")
	}

	exp := backoff.NewExponentialBackOff()
	exp.MaxInterval = 2 * time.Minute
	bo := backoff.WithContext(exp, ctx)
	err = backoff.RetryNotify(func() error {
		return registerNode(fm, nodeID, farmerID, loc)
	}, bo, retryNotify)

	if err != nil {
		return errors.Wrap(err, "failed to register node")
	}

	log.Info().Msg("node has been registered")
	return nil
}

func retryNotify(err error, d time.Duration) {
	log.Warn().Err(err).Str("sleep", d.String()).Msg("registration failed")
}

func registerNode(cl *farmer.Client, nodeID string, farmerID uint32, loc geoip.Location) error {
	log.Info().Msg("registering at farmer bot")

	hostName, err := os.Hostname()
	if err != nil {
		hostName = "unknown"
	}

	return cl.GatewayRegister(farmer.Node{
		ID:       nodeID,
		HostName: hostName,
		FarmID:   farmerID,
		Secret:   "",
		Location: farmer.Location{
			Continent: loc.Continent,
			Country:   loc.Country,
			City:      loc.City,
			Longitude: loc.Longitute,
			Latitude:  loc.Latitude,
		},
	})
}
