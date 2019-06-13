#!/bin/bash
set -ex

apt-get update
apt-get install git gcc make wget -y

# make output directory
ARCHIVE=/tmp/archives
COREDNS_REDIS_FLIST=/tmp/redis_coredns_flist_dir


mkdir -p $ARCHIVE
mkdir -p $COREDNS_REDIS_FLIST/bin

# install go
GOFILE=go1.12.linux-amd64.tar.gz
wget https://dl.google.com/go/$GOFILE
tar -C /usr/local -xzf $GOFILE
mkdir -p /root/go
export GOPATH=/root/go
export PATH=$PATH:/usr/local/go/bin:$GOPATH/go/bin


git clone https://github.com/coredns/coredns /tmp/coredns

pushd /tmp/coredns
    echo 'redis:github.com/arvancloud/redis' >> plugin.cfg
    make
    chmod +x coredns

    cp coredns $COREDNS_REDIS_FLIST/bin/

popd


# copy /etc/ssl too 
cp /etc/ssl -R $COREDNS_REDIS_FLIST


# make sure binary is executable
chmod +x $COREDNS_REDIS_FLIST/bin/*


cat << EOF > $COREDNS_REDIS_FLIST/Corefile
. {
    redis  {
        address 127.0.0.1:6379
    }
    forward 8.8.8.8 9.9.9.9 

}
EOF

cat << EOF > $COREDNS_REDIS_FLIST/.startup.toml

[startup.coredns]
name = "core.system"
after = "redis"
protected = true

[startup.coredns.args]
name = "coredns"
args = ["-conf", "/Corefile"]

[startup.redis]
name = "core.system"
protected = true

[startup.redis.args]
name = "redis-server"

EOF


pushd /tmp
wget https://gist.githubusercontent.com/xmonader/5d1fc6134f1f65acd0d10f71453adb27/raw/2190cef40e75dda44112ac9d31840c958980cd16/copy-chroot.sh
chmod +x copy-chroot.sh

apt install -y redis-server redis-tools

./copy-chroot.sh  `which redis-server` $COREDNS_REDIS_FLIST
./copy-chroot.sh  `which redis-cli` $COREDNS_REDIS_FLIST

popd



tar -czf "/tmp/archives/redis_coredns.tar.gz" -C $COREDNS_REDIS_FLIST .