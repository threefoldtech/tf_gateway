module github.com/threefoldtech/tfgateway

go 1.14

require (
	github.com/cenkalti/backoff/v3 v3.0.0
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/onsi/ginkgo v1.11.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.18.0
	github.com/shirou/gopsutil v2.19.11+incompatible
	github.com/stretchr/testify v1.5.1
	github.com/threefoldtech/tfexplorer v0.2.5
	github.com/threefoldtech/zos v0.2.5
	github.com/urfave/cli/v2 v2.2.0
	gopkg.in/yaml.v2 v2.2.8 // indirect
)

//replace github.com/threefoldtech/zos => ../zos

replace github.com/docker/distribution v2.7.1+incompatible => github.com/docker/distribution v2.7.1-0.20190205005809-0d3efadf0154+incompatible
