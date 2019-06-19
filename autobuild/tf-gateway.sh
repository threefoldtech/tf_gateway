#!/bin/bash
set -ex

apt-get update
apt-get install git gcc make wget -y

# make output directory
ARCHIVE=/tmp/archives
GATEWAY_FLIST=/tmp/gateway_flist_dir

mkdir -p $ARCHIVE
mkdir -p $GATEWAY_FLIST/bin

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
go build -ldflags "-linkmode external -s -w -extldflags -static" -o $GATEWAY_FLIST/bin/tcprouter
popd

# make sure binary is executable
chmod +x $GATEWAY_FLIST/bin/*
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

    cp coredns $GATEWAY_FLIST/bin/

popd


# copy /etc/ssl too
cp /etc/ssl -R $GATEWAY_FLIST


# make sure binary is executable
chmod +x $GATEWAY_FLIST/bin/*


cat << EOF > $GATEWAY_FLIST/Corefile
. {
    redis  {
        address 127.0.0.1:6379
    }
    forward 8.8.8.8 9.9.9.9

}
EOF
cat << EOF > $GATEWAY_FLIST/router.toml
[server]
addr = "0.0.0.0"
port = 443

[server.dbbackend]
type 	 = "redis"
addr     = "127.0.0.1"
port     = 6379
refresh  = 10
EOF



cat << EOF > $GATEWAY_FLIST/bin/redis.sh
#!/bin/sh
if $MASTER_REDIS_IP
then
echo "slaveof $MASTER_REDIS_IP 6379" >>  $GATEWAY_FLIST/etc/redis/redis.conf
fi
EOF




cat << EOF > $GATEWAY_FLIST/.startup.toml

[startup.redis]
name = "core.system"
protected = true

[startup.redis.args]
name = "./bin/redis.sh"

[startup.coredns]
name = "core.system"
after = ["redis"]
protected = true

[startup.coredns.args]
name = "coredns"
args = ["-conf", "/Corefile"]

[startup.tcprouter]
name = "core.system"
after = ["redis"]
protected = true

[startup.tcprouter.args]
name = "tcprouter"
args = ["/router.toml"]



EOF

mkdir -p $GATEWAY_FLIST/etc/redis

cat << EOF > $GATEWAY_FLIST/etc/redis/redis.conf

bind 0.0.0.0
protected-mode yes
port 6379

tcp-backlog 511
timeout 0
tcp-keepalive 300
daemonize no
supervised no
pidfile /var/run/redis_6379.pid

loglevel notice
logfile ""
always-show-logo yes
save 900 1
save 300 10
save 60 10000
stop-writes-on-bgsave-error yes

rdbcompression yes
rdbchecksum yes
dbfilename dump.rdb
dir ./
replica-serve-stale-data yes
replica-read-only yes
repl-diskless-sync no
repl-diskless-sync-delay 5
repl-disable-tcp-nodelay no
replica-priority 100
lazyfree-lazy-eviction no
lazyfree-lazy-expire no
lazyfree-lazy-server-del no
replica-lazy-flush no

appendonly no
appendfilename "appendonly.aof"
appendfsync everysec
no-appendfsync-on-rewrite no

auto-aof-rewrite-percentage 100
auto-aof-rewrite-min-size 64mb
aof-load-truncated yes
aof-use-rdb-preamble yes
lua-time-limit 5000
slowlog-log-slower-than 10000
slowlog-max-len 128
latency-monitor-threshold 0
notify-keyspace-events ""
hash-max-ziplist-entries 512
hash-max-ziplist-value 64

list-max-ziplist-size -2
list-compress-depth 0

set-max-intset-entries 512
zset-max-ziplist-entries 128
zset-max-ziplist-value 64
hll-sparse-max-bytes 3000

stream-node-max-bytes 4096
stream-node-max-entries 100
activerehashing yes

client-output-buffer-limit normal 0 0 0
client-output-buffer-limit replica 256mb 64mb 60
client-output-buffer-limit pubsub 32mb 8mb 60

EOF


pushd /tmp
wget https://gist.githubusercontent.com/xmonader/5d1fc6134f1f65acd0d10f71453adb27/raw/2190cef40e75dda44112ac9d31840c958980cd16/copy-chroot.sh
chmod +x copy-chroot.sh

apt install -y redis-server redis-tools

./copy-chroot.sh  `which redis-server` $GATEWAY_FLIST
./copy-chroot.sh  `which redis-cli` $GATEWAY_FLIST

popd

tar -czf "/tmp/archives/gateway.tar.gz" -C $GATEWAY_FLIST .