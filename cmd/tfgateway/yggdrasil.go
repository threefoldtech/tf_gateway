package main

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"os"
	"os/exec"
	"runtime"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/zos/pkg/network/latency"
	"github.com/threefoldtech/zos/pkg/network/yggdrasil"
)

func ygg(ctx context.Context, sk ed25519.PrivateKey) error {
	bin, err := exec.LookPath("yggdrasil")
	if err != nil {
		return errors.Wrap(err, "yggdrasil is not installed")
	}

	list, err := yggdrasil.FetchPeerList()
	if err != nil {
		return errors.Wrap(err, "failed to get pubic peers list")
	}

	list = list.Ups()
	endpoints := make([]string, 0, len(list))
	for _, node := range list {
		endpoints = append(endpoints, node.Endpoint)
	}

	ls := latency.NewSorter(endpoints, runtime.NumCPU())
	results := ls.Run(ctx)
	// take top 3
	endpoints = endpoints[:0]
	for _, endpoint := range results {
		endpoints = append(endpoints, endpoint.Endpoint)
		if len(endpoints) == 3 {
			break
		}
	}

	cfg := yggdrasil.GenerateConfig(sk)
	cfg.Peers = endpoints

	const path = "/tmp/gateway-yggdrasil.conf"
	file, err := os.Create(path)
	if err != nil {
		return errors.Wrap(err, "failed to create yggdrasil config")
	}

	defer file.Close()
	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	if err := enc.Encode(cfg); err != nil {
		return errors.Wrap(err, "failed to write yggdrasil config")
	}
	file.Close()

	cmd := exec.CommandContext(ctx, bin, "-useconffile", path, "-loglevel", "trace")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return errors.Wrap(err, "failed to start yggdrasil")
	}

	go func() {
		if err := cmd.Wait(); err != nil {
			select {
			case <-ctx.Done():
			default:
				log.Fatal().Err(err).Msg("yggdrasil exited unexpectedly")
			}
		}
	}()

	return nil
}
