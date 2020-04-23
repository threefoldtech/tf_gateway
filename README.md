# TFGateway ![Tests and Build](https://github.com/threefoldtech/tfgateway/workflows/Tests%20and%20Build/badge.svg)

## Supported primitives

- Delegation of domain
- Creation of A or AAAA DNS records
- HTTP(S) proxy
- Reverse tunnel TCP proxy: https://github.com/threefoldtech/tcprouter#reverse-tunneling

## Deployment

![arhcitecure_overview](docs/asset/overview.png)

The TFGateway works be reading the reservation detail from the TFExplorer. It then convert these reservation into configuration readable by the TCP Router server or CoreDNS and write them into a redis server.

Both CoreDNS and the TCP router are watching redis and reloads there internal configuration evrytime there is a change in redis.

### TCP Router server example configuration

```toml
[server]
addr = "0.0.0.0"
port = 443
httpport = 80
clientsport = 18000
[server.dbbackend]
type 	 = "redis"
addr     = "127.0.0.1"
port     = 6379
refresh  = 10
```

### CoreDNS example configuration


```
. {
    redis  {
        address 127.0.0.1:6379
    }
}
```