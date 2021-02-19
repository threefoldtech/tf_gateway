package main

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/rusart/muxprom"

	"github.com/threefoldtech/tfgateway"
	"github.com/threefoldtech/tfgateway/dns"
	"github.com/threefoldtech/tfgateway/proxy"
	"github.com/threefoldtech/tfgateway/redis"
	"github.com/threefoldtech/tfgateway/wg"
	"github.com/threefoldtech/zos/pkg/app"
	"github.com/threefoldtech/zos/pkg/crypto"
	"github.com/threefoldtech/zos/pkg/provision"
	"github.com/threefoldtech/zos/pkg/provision/api"
	"github.com/threefoldtech/zos/pkg/provision/storage"
	"github.com/threefoldtech/zos/pkg/substrate"
	"github.com/threefoldtech/zos/pkg/utils"
	"github.com/threefoldtech/zos/pkg/version"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/asaskevich/govalidator"
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
			Name:  "http",
			Usage: "http listen address",
			Value: ":2021",
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
		&cli.Int64Flag{
			Name:  "farm",
			Usage: "The farm ID of the tfgateway",
		},
		&cli.StringFlag{
			Name:     "farm-secret",
			Usage:    "The farm secret to join the farm",
			Required: true,
		},
		&cli.StringFlag{
			Name:  "storage",
			Usage: "Storage path for workload files",
			Value: "/var/cache/gateway",
		},
		&cli.StringFlag{
			Name:  "substrate",
			Usage: "url to substrate",
			Value: "wss://tfgrid.tri-fold.com", //TODO: this should change to production substrate
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
		&cli.BoolFlag{
			Name:  "free",
			Usage: "if specified, the gateway will be marked as free to use and capacity can be reserved using FreeTFT",
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

func validDomain(d string) bool {
	return govalidator.IsDNSName(d)
}

func validDomains(ds []string) bool {
	for _, d := range ds {
		if !validDomain(d) {
			return false
		}
	}
	return true
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
		return fmt.Errorf("failed to connect to redis configuration server: %w", err)
	}

	kp, err := ensureID(c.String("seed"))
	if err != nil {
		return err
	}

	domains := c.StringSlice("domains")
	nameservers := c.StringSlice("nameservers")
	if !validDomains(domains) {
		return fmt.Errorf("invalid domains: %v", domains)
	}

	if !validDomains(nameservers) {
		return fmt.Errorf("invalid nameservers: %v", nameservers)
	}

	dnsMgr := dns.New(pool, kp.Identity())
	if err := dnsMgr.Cleanup(); err != nil {
		log.Fatal().Err(err).Msg("failed to clean up coredns config")
	}

	for _, domain := range domains {
		log.Info().Msgf("gateway will manage domain %s", domain)
		if err := dnsMgr.AddDomainDelagate(kp.Identity(), kp.Identity(), domain); err != nil {
			return errors.Wrapf(err, "fail to manage domain %s", domain)
		}
	}

	var wgMgr *wg.Mgr
	if is4To6Enabled(c) {
		// // the Gateway4To6 workloads doesn't not survice gateway restart, so we clear all reservations from the
		// // cache so they are always re-provisionned
		// if err := localStore.ClearByType([]provision.ReservationType{tfgateway.Gateway4To6Type}); err != nil {
		// 	return err
		// }

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
	storagePath := c.String("storage")
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		return errors.Wrap(err, "failed to create storage directory")
	}

	substrateURL := c.String("substrate")
	storage, err := storage.NewFSStore(storagePath)
	users, err := substrate.NewSubstrateUsers(substrateURL)

	provisioner := tfgateway.NewProvisioner(proxy.New(pool), dnsMgr, wgMgr, kp)
	engine := provision.New(
		storage,
		provisioner,
		provision.WithUsers(users),
		tfgateway.Order,
	)

	log.Info().Str("identity", kp.Identity()).Msg("starting gateway")

	ctx := context.Background()

	ctx, _ = utils.WithSignal(ctx)
	utils.OnDone(ctx, func(_ error) {
		log.Info().Msg("shutting down")
	})

	go func() {
		if err := engine.Run(ctx); err != nil && err != context.Canceled {
			log.Fatal().Err(err).Msg("unexpected error")
		}
		log.Info().Msg("provision engine stopped")
	}()

	httpServer, err := getHTTPServer(engine)
	if err != nil {
		return errors.Wrap(err, "failed to create http server")
	}

	httpServer.Addr = c.String("http")
	utils.OnDone(ctx, func(_ error) {
		httpServer.Close()
	})

	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return errors.Wrap(err, "http api exited unexpectedely")
	}

	return nil
}

func getHTTPServer(engine provision.Engine) (*http.Server, error) {
	router := mux.NewRouter().StrictSlash(true)

	prom := muxprom.New(
		muxprom.Router(router),
		muxprom.Namespace("gateway"),
	)
	prom.Instrument()

	v1 := router.PathPrefix("/api/v1").Subrouter()

	_, err := api.NewWorkloadsAPI(v1, engine)
	if err != nil {
		return nil, errors.Wrap(err, "failed to setup workload api")
	}

	return &http.Server{
		Handler: router,
	}, nil
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

func is4To6Enabled(c *cli.Context) bool {
	for _, s := range []string{c.String("endpoint"), c.String("wg-iface")} {
		if s == "" {
			return false
		}
	}
	return true
}
