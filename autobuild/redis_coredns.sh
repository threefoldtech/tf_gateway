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


tar -czf "/tmp/archives/redis_coredns.tar.gz" -C $COREDNS_REDIS_FLIST .