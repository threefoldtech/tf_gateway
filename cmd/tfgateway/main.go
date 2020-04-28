package main

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/cenkalti/backoff/v3"

	"github.com/shirou/gopsutil/host"
	"github.com/threefoldtech/tfgateway"
	"github.com/threefoldtech/tfgateway/cache"
	"github.com/threefoldtech/tfgateway/dns"
	"github.com/threefoldtech/tfgateway/proxy"
	"github.com/threefoldtech/tfgateway/redis"
	"github.com/threefoldtech/tfgateway/wg"
	"github.com/threefoldtech/zos/pkg/app"
	"github.com/threefoldtech/zos/pkg/crypto"
	"github.com/threefoldtech/zos/pkg/geoip"
	"github.com/threefoldtech/zos/pkg/provision"
	"github.com/threefoldtech/zos/pkg/provision/explorer"
	"github.com/threefoldtech/zos/pkg/utils"
	"github.com/threefoldtech/zos/pkg/version"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/threefoldtech/tfexplorer/client"
	"github.com/threefoldtech/tfexplorer/models/generated/directory"
	"github.com/threefoldtech/zos/pkg/identity"
	"github.com/urfave/cli/v2"
)

var appCLI = cli.App{
	Version: version.Current().String(),
	Usage:   "",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "seed",
			Value: "identity.seed",
			Usage: "path to the file containing the identity seed",
		},
		&cli.StringFlag{
			Name:  "explorer",
			Value: "https://explorer.grid.tf/explorer",
			Usage: "URL to the explorer API used to poll reservations",
		},
		&cli.StringFlag{
			Name:  "redis",
			Value: "tcp://localhost:6379",
			Usage: "address of the redis configuration server",
		},
		&cli.BoolFlag{
			Name:  "debug",
			Usage: "enable debug logging",
		},
		&cli.StringSliceFlag{
			Name:  "nameservers",
			Usage: "list of DNS nameserver used by this TFGateway. User use this value to know where to point there NS record in order to delegate domain to the TFGateway",
		},
		&cli.StringSliceFlag{
			Name:  "domains",
			Usage: "list of domain managed by this TFGateway. User can create free subdomain of any domain managed by TFGateway",
		},
		&cli.Int64Flag{
			Name:  "tcp-client-port",
			Usage: "the listening port on which the TCP router client needs to connect to in order to initiate a reverse tunnel",
			Value: 18000,
		},
		&cli.StringFlag{
			Name:  "endpoint",
			Usage: "listening address of the wireguard interface, format: host:port",
		},
		&cli.StringFlag{
			Name:  "wg-iface",
			Usage: "name of the wireguard interface created for the 4to6 tunnel primitive",
			Value: "wg-tfgateway",
		},
	},
	Before: func(c *cli.Context) error {
		app.Initialize()

		// Default level for this example is info, unless debug flag is present
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		if c.Bool("debug") {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		}

		return nil
	},
	Action: run,
}

func main() {
	err := appCLI.Run(os.Args)
	if err != nil {
		log.Fatal().Err(err).Send()
	}
}

func run(c *cli.Context) error {
	pool, err := redis.NewPool(c.String("redis"))
	if err != nil {
		return fmt.Errorf("failed to connec to redis configuration server: %w", err)
	}

	kp, err := ensureID(c.String("seed"))
	if err != nil {
		return err
	}

	staster := &tfgateway.Counters{}
	localStore := cache.NewRedis(pool)
	if err := localStore.Sync(staster); err != nil {
		return fmt.Errorf("failed to sync statser with reservation from cache: %w", err)
	}

	wgID := gwIdentity{kp}
	e, err := client.NewClient(c.String("explorer"), wgID)
	if err != nil {
		return fmt.Errorf("failed to instantiate explorer client: %w", err)
	}

	loc, err := geoip.Fetch()
	if err != nil {
		return fmt.Errorf("failed to fetch location: %w", err)
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
		ManagedDomains: c.StringSlice("domains"),
		DnsNameserver:  c.StringSlice("nameservers"),
		TcpRouterPort:  c.Int64("tcp-client-port"),
	}

	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = 0
	if err := backoff.Retry(registerID(gw, e), bo); err != nil {
		return fmt.Errorf("failed to register gateway in the explorer: %w", err)
	}

	var wgMgr *wg.Mgr
	if is4To6Enabled(c) {

		// the Gateway4To6 workloads doesn't not survice gateway restart, so we clear all reservations from the
		// cache so they are always re-provisionned
		if err := localStore.ClearByType([]provision.ReservationType{tfgateway.Gateway4To6Reservation}); err != nil {
			return err
		}

		var (
			endpoint = c.String("endpoint")
			wgIface  = c.String("wg-iface")
		)
		log.Info().
			Str("endpoint", endpoint).
			Str("wg-iface", wgIface).
			Msg("gateway 4 to 6 enabled")

		wgMgr, err = wg.New(kp, wg.NewIPPool(kp), endpoint, wgIface)
		if err != nil {
			return err
		}

		defer func() {
			if err := wgMgr.Close(); err != nil {
				log.Error().Err(err).Msgf("failed to cleanup network namespace")
			}
		}()
	}

	provisioner := tfgateway.NewProvisioner(proxy.New(pool), dns.New(pool), wgMgr)

	engine := provision.New(provision.EngineOps{
		NodeID: kp.Identity(),
		Cache:  localStore,
		Source: provision.CombinedSource(
			provision.PollSource(explorer.NewPoller(e, tfgateway.WorkloadToProvisionType, tfgateway.ProvisionOrder), kp),
			provision.NewDecommissionSource(localStore),
		),
		Provisioners:   provisioner.Provisioners,
		Decomissioners: provisioner.Decommissioners,
		Feedback:       tfgateway.NewFeedback(e, tfgateway.ResultToSchemaType),
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
	return nil
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

func is4To6Enabled(c *cli.Context) bool {
	for _, s := range []string{c.String("endpoint"), c.String("wg-iface")} {
		if s == "" {
			return false
		}
	}
	return true
}
