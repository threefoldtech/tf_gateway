#!/bin/bash
set -ex

apt-get update
apt-get install git gcc wget -y

# make output directory
ARCHIVE=/tmp/archives
TCPROUTER_FLIST=/tmp/tfchain


mkdir -p $ARCHIVE
mkdir -p $TCPROUTER_FLIST/bin

# install go
GOFILE=go1.12.linux-amd64.tar.gz
wget https://dl.google.com/go/$GOFILE
tar -C /usr/local -xzf $GOFILE
mkdir -p /root/go
export GOPATH=/root/go
export PATH=$PATH:/usr/local/go/bin:$GOPATH/go/bin
go get -u github.com/xmonader/tcprouter

TCPROUTER=$GOPATH/src/github.com/xmonader/tcprouter


pushd $TCPROUTER
go build -ldflags "-linkmode external -s -w -extldflags -static" -o $TCPROUTER_FLIST/bin/tcprouter
popd

# make sure binary is executable
chmod +x $TCPROUTER_FLIST/bin/*



tar -czf "/tmp/archives/tcprouter.tar.gz" -C $TCPROUTER_FLIST .