#!/bin/bash
set -ex

apt-get update
apt-get install git gcc wget -y

# make output directory
ARCHIVE=/tmp/archives
TCPROUTER_FLIST=/tmp/tcprouter_dir


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



pushd /tmp
wget https://gist.githubusercontent.com/xmonader/5d1fc6134f1f65acd0d10f71453adb27/raw/2190cef40e75dda44112ac9d31840c958980cd16/copy-chroot.sh
chmod +x copy-chroot.sh

apt install -y redis-server redis-tools

./copy-chroot.sh  `which redis-server` $TCPROUTER_FLIST
./copy-chroot.sh  `which redis-cli` $TCPROUTER_FLIST

popd



cat << EOF > $TCPROUTER_FLIST/router.toml
[server]
addr = "0.0.0.0"
port = 443

[server.dbbackend]
type 	 = "redis"
addr     = "127.0.0.1"
port     = 6379
refresh  = 10
EOF


cat << EOF > $TCPROUTER_FLIST/.startup.toml

[startup.tcprouter]
name = "core.system"
after = "redis"
protected = true

[startup.tcprouter.args]
name = "tcprouter"
args = ["/router.toml"]

[startup.redis]
name = "core.system"
protected = true

[startup.redis.args]
name = "redis-server"

EOF


tar -czf "/tmp/archives/tcprouter.tar.gz" -C $TCPROUTER_FLIST .