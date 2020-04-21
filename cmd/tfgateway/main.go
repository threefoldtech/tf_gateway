package main

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cenkalti/backoff/v3"

	"github.com/shirou/gopsutil/host"
	"github.com/threefoldtech/tfgateway/dns"
	"github.com/threefoldtech/tfgateway/proxy"
	"github.com/threefoldtech/zos/pkg/crypto"
	"github.com/threefoldtech/zos/pkg/provision/explorer"
	"github.com/threefoldtech/zos/pkg/provision/primitives/cache"

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

func main() {
	app.Initialize()

	var (
		explorerAddr string
		redisAddr    string
		seed         string
		storageDir   string
		debug        bool
		ver          bool
	)

	flag.StringVar(&seed, "seed", "identity.seed", "path to the file containing the identity seed")
	flag.StringVar(&explorerAddr, "explorer", "https://explorer.grid.tf/explorer", "URL to the explorer API used to poll reservations")
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

	pool, err := newRedisPool(redisAddr)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connec to redis configuration server")
	}

	if err := os.MkdirAll(storageDir, 0770); err != nil {
		log.Fatal().Err(err).Msg("failed to create cache directory")
	}

	app.BootedPath = filepath.Join(storageDir, "booted")

	kp, err := ensureID(seed)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	staster := &tfgateway.Counters{}
	localStore, err := cache.NewFSStore(filepath.Join(storageDir, "reservations"))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create local reservation store")
	}
	if err := localStore.Sync(staster); err != nil {
		log.Fatal().Err(err).Msg("failed to sync statser with reservation from cache")
	}

	wgID := gwIdentity{kp}
	e, err := client.NewClient(explorerAddr, wgID)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to instantiate explorer client")
	}

	loc, err := geoip.Fetch()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to fetch location")
	}

	gw := directory.Gateway{
		NodeId:       kp.Identity(),
		PublicKeyHex: hex.EncodeToString(kp.PublicKey),
		OsVersion:    version.Current().Short(),
		Location: directory.Location{
			Continent: loc.Continent,
			Country:   loc.Country,
			City:      loc.City,
			Longitude: loc.Longitute,
			Latitude:  loc.Latitude,
		},
		// ManagedDomains: ,
		// TcpRouterPort: ,
		// DnsNameserver: ,
	}

	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = 0
	if err := backoff.Retry(registerID(gw, e), bo); err != nil {
		log.Fatal().Err(err).Msg("failed to register gateway in the explorer")
	}

	provisioner := tfgateway.NewProvisioner(proxy.New(pool), dns.New(pool, "_dns"))

	engine := provision.New(provision.EngineOps{
		NodeID: kp.Identity(),
		Cache:  localStore,
		Source: provision.CombinedSource(
			provision.PollSource(explorer.NewPoller(e, tfgateway.WorkloadToProvisionType, tfgateway.ProvisionOrder), kp),
			provision.NewDecommissionSource(localStore),
		),
		Provisioners:   provisioner.Provisioners,
		Decomissioners: provisioner.Decommissioners,
		Feedback:       explorer.NewFeedback(e, tfgateway.ResultToSchemaType),
		Signer:         wgID,
		Statser:        staster,
	})

	log.Info().Str("identity", kp.Identity()).Msg("starting gateway")

	ctx := context.Background()

	(&uptimeTicker{
		nodeID:   kp.Identity(),
		explorer: e,
	}).Start(ctx)

	ctx, _ = utils.WithSignal(ctx)
	utils.OnDone(ctx, func(_ error) {
		log.Info().Msg("shutting down")
	})

	if err := engine.Run(ctx); err != nil {
		log.Error().Err(err).Msg("unexpected error")
	}
	log.Info().Msg("provision engine stopped")
}

func registerID(gw directory.Gateway, expl *client.Client) func() error {
	return func() error {
		log.Info().Str("ID", gw.NodeId).Msg("trying to register to the explorer")
		return expl.Directory.GatewayRegister(gw)
	}
}

func ensureID(seed string) (kp identity.KeyPair, err error) {
	kp, err = identity.LoadKeyPair(seed)
	if err != nil && !os.IsNotExist(err) {
		return kp, fmt.Errorf("failed to read identity seed: %w", err)
	}
	if os.IsNotExist(err) {
		if kp, err = identity.GenerateKeyPair(); err != nil {
			return kp, fmt.Errorf("failed to generate new key pair from seed: %w", err)
		}

		if err := kp.Save(seed); err != nil {
			return kp, fmt.Errorf("failed to save new key pair into file %s: %w", seed, err)
		}
	}

	return kp, nil
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

type uptimeTicker struct {
	nodeID   string
	explorer *client.Client
}

// Uptime returns the uptime of the node
func (u *uptimeTicker) uptime() (uint64, error) {
	info, err := host.Info()
	if err != nil {
		return 0, err
	}
	return info.Uptime, nil
}

func (u *uptimeTicker) Start(ctx context.Context) {
	sendUptime := func() error {
		uptime, err := u.uptime()
		if err != nil {
			log.Error().Err(err).Msgf("failed to read uptime")
			return err
		}

		log.Info().Msg("send heart-beat to BCDB")
		if err := u.explorer.Directory.GatewayUpdateUptime(u.nodeID, uptime); err != nil {
			log.Error().Err(err).Msgf("failed to send heart-beat to BCDB")
			return err
		}
		return nil
	}

	_ = backoff.Retry(sendUptime, backoff.NewExponentialBackOff())

	tick := time.NewTicker(time.Minute * 10)

	go func() {
		defer tick.Stop()

		for {
			select {
			case <-tick.C:
				backoff.Retry(sendUptime, backoff.NewExponentialBackOff())
			case <-ctx.Done():
				return
			}
		}
	}()

}
