module github.com/threefoldtech/tfgateway

go 1.14

require (
	github.com/alicebob/miniredis/v2 v2.11.4
	github.com/cenkalti/backoff/v3 v3.2.2
	github.com/containerd/continuity v0.0.0-20200710164510-efbc4488d8fe // indirect
	github.com/containerd/ttrpc v1.0.1 // indirect
	github.com/containernetworking/plugins v0.8.4
	github.com/gogo/googleapis v1.4.0 // indirect
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/onsi/ginkgo v1.11.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.19.0
	github.com/shirou/gopsutil v2.19.11+incompatible
	github.com/stretchr/testify v1.6.1
	github.com/threefoldtech/tfexplorer v0.3.2-0.20200714142941-bf454397bab9
	github.com/threefoldtech/zos v0.4.0-rc2.0.20200714104826-f9fc1e4b069b
	github.com/urfave/cli/v2 v2.2.0
	github.com/vishvananda/netlink v1.1.0
	golang.zx2c4.com/wireguard/wgctrl v0.0.0-20191219145116-fa6499c8e75f
	gopkg.in/yaml.v2 v2.2.8 // indirect
)

replace github.com/docker/distribution v2.7.1+incompatible => github.com/docker/distribution v2.7.1-0.20190205005809-0d3efadf0154+incompatible
