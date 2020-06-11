module github.com/threefoldtech/tfgateway

go 1.14

require (
	github.com/alicebob/miniredis/v2 v2.11.4
	github.com/cenkalti/backoff/v3 v3.2.2
	github.com/containernetworking/plugins v0.8.4
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/onsi/ginkgo v1.11.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.18.0
	github.com/shirou/gopsutil v2.19.11+incompatible
	github.com/stretchr/testify v1.5.1
	github.com/threefoldtech/tfexplorer v0.3.1
	github.com/threefoldtech/zos v0.3.4
	github.com/urfave/cli/v2 v2.2.0
	github.com/vishvananda/netlink v1.0.0
	golang.zx2c4.com/wireguard/wgctrl v0.0.0-20191219145116-fa6499c8e75f
	gopkg.in/yaml.v2 v2.2.8 // indirect
)

replace github.com/docker/distribution v2.7.1+incompatible => github.com/docker/distribution v2.7.1-0.20190205005809-0d3efadf0154+incompatible
