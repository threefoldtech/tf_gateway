module github.com/threefoldtech/tfgateway

go 1.14

replace github.com/threefoldtech/zos => ../zos

replace github.com/threefoldtech/tfexplorer => ../tfexplorer

replace github.com/docker/distribution v2.7.1+incompatible => github.com/docker/distribution v2.7.1-0.20190205005809-0d3efadf0154+incompatible

require (
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/go-events v0.0.0-20190806004212-e31b211e4f1c // indirect
	github.com/go-ole/go-ole v1.2.4 // indirect
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/onsi/ginkgo v1.11.0 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/opencontainers/runc v0.1.1 // indirect
	github.com/pkg/errors v0.9.1
	github.com/polydawn/refmt v0.0.0-20190807091052-3d65705ee9f1
	github.com/rs/zerolog v1.18.0
	github.com/shirou/w32 v0.0.0-20160930032740-bb4de0191aa4 // indirect
	github.com/stretchr/testify v1.5.1
	github.com/syndtr/gocapability v0.0.0-20180916011248-d98352740cb2 // indirect
	github.com/threefoldtech/tfexplorer v0.0.0-00010101000000-000000000000
	github.com/threefoldtech/zos v0.2.3
	github.com/warpfork/go-wish v0.0.0-20200122115046-b9ea61034e4a // indirect
	go.etcd.io/bbolt v1.3.1-etcd.8 // indirect
	gopkg.in/yaml.v2 v2.2.8 // indirect
	upspin.io v0.0.0-20200124224713-5ddde7b8e6ff // indirect
)
