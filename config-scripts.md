## Configuration file creation scripts



### Identity seed file 

File: ```/etc/identity.seed```

```shell
cat << EOF > /etc/identity.seed
"1.1.0"{"mnemonic":"$$3BOT_WORDS","threebotid":$$3BOT_ID}
EOF
```

### CoreDNS 

Configuration file: ```/etc/coredns/corefile```

```shell
cat << EOF > /etc/coredns/Corefile
. {
    log
    errors
    redis  {
        address 127.0.0.1:6379
    }
    forward . 174.138.6.79 8.8.8.8 1.1.1.1
}
EOF
```

Service configuration: ```/etc/systemd/system/coredns.service```

```
cat << EOF > /etc/systemd/system/codedns.service
[Unit]
Description=CoreDNS
After=network.target
After=tfredis.target

[Service]
ExecStart=/usr/local/bin/coredns -conf /etc/coredns/Corefile
Type=simple
Restart=on-failure
MemoryAccounting=true
MemoryHigh=800M
MemoryMax=1G

[Install]
WantedBy=multi-user.target
EOF
```

### TF Redis 

Configuration file: `/etc/tfredis.conf`

```shell
cat << EOF > /etc/tfredis.conf
bind 127.0.0.1
EOF
```

Service configuration: `/etc/systemd/system/tfredis.service`

```shell
cat << EOF > /etc/systemd/system/tfredis.service
[Unit]
Description=The Redis server for TFGateway
After=network.target

[Service]
Type=simple
Environment=statedir=/run/redis
PIDFile=/run/redis/redis.pid
ExecStartPre=/bin/touch /var/log/redis.log
ExecStartPre=/bin/mkdir -p /run/redis
ExecStart=redis-server /etc/tfredis.conf
ExecReload=/bin/kill -USR2 $MAINPID
MemoryAccounting=true
MemoryHigh=800M
MemoryMax=1G
LimitNOFILE=10050

[Install]
WantedBy=multi-user.target
EOF
```

### TCPRouter 

Configuration file: ```/etc/tcprouter/router.toml```

```shell
cat << EOF > /etc/tcprouter/router.toml
[server]
addr = "0.0.0.0"
port = 443
httpport = 80
clientsport = 18000
[server.dbbackend]
type     = "redis"
addr     = "127.0.0.1"
port     = 6379
refresh  = 10
EOF
```

Service configuration: `/etc/systemd/system/tcprouter.service`

```shell
cat << EOF > /etc/systemd/system/tcprouter.service
[Unit]
Description=TCP router server
After=network.target
After=coredns.target

[Service]
ExecStart=/usr/local/bin/trs --config /etc/tcprouter/router.toml
Type=simple
Restart=on-failure
MemoryAccounting=true
MemoryHigh=800M
MemoryMax=1G

[Install]
WantedBy=multi-user.target
EOF
```

### TFGateway

Service configuration: ```/etc/systemd/system/tfgateway.service```

```shell
cat << EOF > /etc/systemd/system/tfgateway.service
[Unit]
Description=tfgateway server
After=network.target
After=tcprouter.target

[Service]
ExecStartPre=/bin/bash -c "/bin/systemctl set-environment ip=$(/sbin/ip r get 1.1.1.1 | awk '{print $7}')"
ExecStart=/usr/local/bin/tfgateway --seed /etc/identity.seed --explorer $$EXPLORER_URL/api/v1 --nameservers $$NAMESERVER --endpoint ${ip}:3443 --domains $$DOMAIN --farm $$FARM_ID

Type=simple
Restart=on-failure
MemoryAccounting=true
MemoryHigh=800M
MemoryMax=1G

[Install]
WantedBy=multi-user.target
EOF
```


Make sure to replace variables starting with `$$` in the above with the right values