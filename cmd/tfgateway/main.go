package main

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"flag"
	"os"
	"path/filepath"

	"github.com/threefoldtech/zos/pkg/crypto"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/threefoldtech/tfexplorer/client"
	"github.com/threefoldtech/tfexplorer/models/generated/directory"
	"github.com/threefoldtech/tfgateway"
	"github.com/threefoldtech/zos/pkg/app"
	"github.com/threefoldtech/zos/pkg/geoip"
	"github.com/threefoldtech/zos/pkg/identity"
	"github.com/threefoldtech/zos/pkg/provision"
	"github.com/threefoldtech/zos/pkg/utils"
	"github.com/threefoldtech/zos/pkg/version"
)

const (
	module = "provision"
)

func main() {
	app.Initialize()

	var (
		explorer   string
		redisAddr  string
		seed       string
		storageDir string
		debug      bool
		ver        bool
	)

	flag.StringVar(&seed, "seed", "identity.seed", "path to the file containing the identity seed")
	flag.StringVar(&explorer, "explorer", "https://explorer.grid.tf/explorer", "URL to the explorer API used to poll reservations")
	flag.StringVar(&redisAddr, "redis", "tcp://localhost:6379", "Addr of the redis configuration server")
	flag.StringVar(&storageDir, "data-dir", "tfgateway_datadir", "directory used by tfgateway to store temporary data")
	flag.BoolVar(&debug, "debug", false, "enable debug logging")
	flag.BoolVar(&ver, "v", false, "show version and exit")

	flag.Parse()
	if ver {
		version.ShowAndExit(false)
	}

	// Default level for this example is info, unless debug flag is present
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	flag.Parse()

	pool, err := tfgateway.NewRedisPool(redisAddr)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connec to redis configuration server")
	}

	if err := os.MkdirAll(storageDir, 0770); err != nil {
		log.Fatal().Err(err).Msg("failed to create cache directory")
	}

	app.BootedPath = filepath.Join(storageDir, "booted")

	var kp identity.KeyPair
	kp, err = identity.LoadKeyPair(seed)
	if err != nil && !os.IsNotExist(err) {
		log.Fatal().Err(err).Msg("failed to read identity seed")
	}
	if os.IsNotExist(err) {
		if kp, err = identity.GenerateKeyPair(); err != nil {
			log.Fatal().Err(err).Msgf("failed to generate new key pair from seed")
		}

		if err := kp.Save(seed); err != nil {
			log.Fatal().Err(err).Msgf("failed to save new key pair into file %s", seed)
		}
	}

	loc, err := geoip.Fetch()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to fetch location")
	}

	wgID := gwIdentity{kp}
	cl, err := client.NewClient(explorer, wgID)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to instantiate explorer client")
	}

	if err := cl.Directory.GatewayRegister(directory.Gateway{
		NodeId:       kp.Identity(),
		PublicKeyHex: hex.EncodeToString(kp.PublicKey),
		OsVersion:    version.Current().Short(),
		// Created: ,
		// Updated: ,
		// Uptime: ,
		// Address: ,
		Location: directory.Location{
			Continent: loc.Continent,
			Country:   loc.Country,
			City:      loc.City,
			Longitude: loc.Longitute,
			Latitude:  loc.Latitude,
		},
		// Workloads: ,
		// ManagedDomains: ,
		// TcpRouterPort: ,
		// DnsNameserver: ,
	}); err != nil {
		log.Fatal().Err(err).Msg("failed to register gateway in the explorer")
	}

	// to store reservation locally on the gateway
	localStore, err := provision.NewFSStore(filepath.Join(storageDir, "reservations"))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create local reservation store")
	}

	// create context and add middlewares
	ctx := context.Background()

	dns, err := tfgateway.NewCoreDNS(pool, "_dns")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create DNS manager")
	}

	proxy, err := tfgateway.NewTCPRouter(pool)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create TCP proxy manager")
	}

	provisioner := tfgateway.NewProvisioner(proxy, dns)

	engine := provision.New(provision.EngineOps{
		NodeID: kp.Identity(),
		Cache:  localStore,
		Source: provision.CombinedSource(
			provision.PollSource(provision.ReservationPollerFromWorkloads(cl.Workloads, tfgateway.WorkloadToProvisionType), kp),
			provision.NewDecommissionSource(localStore),
		),
		Feedback:       provision.NewExplorerFeedback(cl, provision.ToSchemaType),
		Signer:         wgID,
		Provisioners:   provisioner.Provisioners,
		Decomissioners: provisioner.Decommissioners,
	})

	log.Info().Str("identity", kp.Identity()).Msg("starting gateway")

	ctx, _ = utils.WithSignal(ctx)
	utils.OnDone(ctx, func(_ error) {
		log.Info().Msg("shutting down")
	})

	if err := engine.Run(ctx); err != nil {
		log.Error().Err(err).Msg("unexpected error")
	}
	log.Info().Msg("provision engine stopped")
}

type gwIdentity struct {
	kp identity.KeyPair
}

func (n gwIdentity) PrivateKey() ed25519.PrivateKey {
	return n.kp.PrivateKey
}

func (n gwIdentity) Identity() string {
	return n.kp.Identity()
}

func (n gwIdentity) Sign(b []byte) ([]byte, error) {
	return crypto.Sign(n.kp.PrivateKey, b)
}
