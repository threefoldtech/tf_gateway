module github.com/threefoldtech/tfgateway

go 1.14

require (
	github.com/cenkalti/backoff/v3 v3.0.0
	github.com/containerd/continuity v0.0.0-20200413184840-d3ef23f19fbb // indirect
	github.com/containerd/ttrpc v1.0.0 // indirect
	github.com/containernetworking/cni v0.7.2-0.20190807151350-8c6c47d1c7fc
	github.com/containernetworking/plugins v0.8.4
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/go-events v0.0.0-20190806004212-e31b211e4f1c // indirect
	github.com/gogo/googleapis v1.3.2 // indirect
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/onsi/ginkgo v1.11.0 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/opencontainers/runc v0.1.1 // indirect
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.18.0
	github.com/shirou/gopsutil v2.19.11+incompatible
	github.com/stretchr/testify v1.5.1
	github.com/syndtr/gocapability v0.0.0-20180916011248-d98352740cb2 // indirect
	github.com/threefoldtech/tfexplorer v0.2.5
	github.com/threefoldtech/zos v0.2.5
	github.com/urfave/cli/v2 v2.2.0
	github.com/vishvananda/netlink v1.0.0
	golang.zx2c4.com/wireguard/wgctrl v0.0.0-20191219145116-fa6499c8e75f
	gopkg.in/yaml.v2 v2.2.8 // indirect
	gotest.tools v2.2.0+incompatible
)

replace github.com/threefoldtech/tfexplorer => ../tfexplorer

replace github.com/threefoldtech/zos => ../zos

replace github.com/docker/distribution v2.7.1+incompatible => github.com/docker/distribution v2.7.1-0.20190205005809-0d3efadf0154+incompatible
