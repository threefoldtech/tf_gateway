module github.com/threefoldtech/tfgateway

go 1.14

require (
	github.com/alicebob/miniredis/v2 v2.11.4
	github.com/asaskevich/govalidator v0.0.0-20200108200545-475eaeb16496
	github.com/cenkalti/backoff/v3 v3.2.2
	github.com/containernetworking/plugins v0.8.4
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.19.0
	github.com/rusart/muxprom v0.0.0-20200323164249-36ea051efbe6
	github.com/shirou/gopsutil v2.20.6+incompatible // indirect
	github.com/stretchr/testify v1.6.1
	github.com/threefoldtech/zos v0.4.9-0.20210118140854-23f2d049c270
	github.com/urfave/cli/v2 v2.3.0
	github.com/vishvananda/netlink v1.1.0
	github.com/vishvananda/netns v0.0.0-20200520041808-52d707b772fe // indirect
	golang.zx2c4.com/wireguard/wgctrl v0.0.0-20191219145116-fa6499c8e75f
	google.golang.org/protobuf v1.25.0 // indirect
)

replace github.com/docker/distribution v2.7.1+incompatible => github.com/docker/distribution v2.7.1-0.20190205005809-0d3efadf0154+incompatible

replace github.com/threefoldtech/zos => ../zos
