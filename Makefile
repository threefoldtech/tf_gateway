OUT = $(shell realpath -m bin)
GOPATH := $(shell go env GOPATH)
branch = $(shell git symbolic-ref -q --short HEAD || git describe --tags --exact-match)
revision = $(shell git rev-parse HEAD)
dirty = $(shell test -n "`git diff --shortstat 2> /dev/null | tail -n1`" && echo "*")
version = github.com/threefoldtech/zos/pkg/version
ldflags = '-w -s -X $(version).Branch=$(branch) -X $(version).Revision=$(revision) -X $(version).Dirty=$(dirty)'

all: build

getdeps:
	@echo "Installing golint" && go get -u golang.org/x/lint/golint
	@echo "Installing gocyclo" && go get -u github.com/fzipp/gocyclo
	@echo "Installing deadcode" && go get -u github.com/remyoudompheng/go-misc/deadcode
	@echo "Installing misspell" && go get -u github.com/client9/misspell/cmd/misspell
	@echo "Installing ineffassign" && go get -u github.com/gordonklaus/ineffassign

verifiers: vet fmt lint cyclo spelling staticcheck

vet:
	@echo "Running $@"
	@go vet -atomic -bool -copylocks -nilfunc -printf -rangeloops -unreachable -unsafeptr -unusedresult ./...

fmt:
	@echo "Running $@"
	@gofmt -d .

lint:
	@echo "Running $@"
	@${GOPATH}/bin/golint -set_exit_status $(shell go list ./... | grep -v stubs)

ineffassign:
	@echo "Running $@"
	@${GOPATH}/bin/ineffassign .

cyclo:
	@echo "Running $@"
	@${GOPATH}/bin/gocyclo -over 100 .

deadcode:
	@echo "Running $@"
	@${GOPATH}/bin/deadcode -test $(shell go list ./...) || true

spelling:
	@${GOPATH}/bin/misspell -i monitord -error `find .`

staticcheck:
	go run honnef.co/go/tools/cmd/staticcheck -- ./...

test: verifiers build
	# we already ran vet separately, so safe to turn it off here
	@echo "Running unit tests with GOFLAGS=${GOFLAGS}"	
	go test -v -vet=off ./...

testrace: verifiers build
	@echo "Running unit tests with GOFLAGS=${GOFLAGS}"
	# we already ran vet separately, so safe to turn it off here
	go test -v -vet=off -race ./...

build:
	cd cmd/tfgateway && go build -ldflags $(ldflags) -o $(OUT)/tfgateway
