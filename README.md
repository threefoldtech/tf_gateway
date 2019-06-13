# tf_gateway
tcp/http(s) proxy + dns server


# tcprouter

a down to earth tcp router based on traefik tcp streaming and supports multiple backends using [valkyrie](https://github.com/abronan/valkeyrie)


## Build

```
git clone https://github.com/xmonader/tcprouter 
cd tcprouter
go build
```

## Install

```bash
go get -u github.com/xmonader/tcprouter
```



## Running

configfile: router.toml
```toml
[server]
addr = "0.0.0.0"
port = 443

[server.dbbackend]
type 	 = "redis"
addr     = "127.0.0.1"
port     = 6379
refresh  = 10
```
then 
`./tcprouter router.toml`


Please notice if you are using low numbered port like 80 or 443 you can use sudo or setcap before running the binary.
- `sudo ./tcprouter router.toml`
- setcap: `sudo setcap CAP_NET_BIND_SERVICE=+eip PATH_TO_TCPROUTER`



### router.toml
We have two toml sections so far

#### [server]

```toml
[server]
addr = "0.0.0.0"
port = 443
```
in `[server]` section we define the listening interface/port the tcprouter intercepting: typically that's 443 for TLS connections.

#### [server.dbbackend]
```toml
[server.dbbackend]
type 	 = "redis"
addr     = "127.0.0.1"
port     = 6379
refresh  = 10
```
in `server.dbbackend` we define the backend kv store and its connection information `addr,port` and how often we want to reload the data from the kv store using `refresh` key in seconds.



## Data representation in KV

```
127.0.0.1:6379> KEYS *
1) "/tcprouter/services/bing"
2) "/tcprouter/services/google"
3) "/tcprouter/services/facebook"

127.0.0.1:6379> get /tcprouter/services/google
"{\"Key\":\"tcprouter/services/google\",\"Value\":\"eyJhZGRyIjoiMTcyLjIxNy4xOS40Njo0NDMiLCJzbmkiOiJ3d3cuZ29vZ2xlLmNvbSJ9\",\"LastIndex\":75292246}"

```

### Decoding data from python

```ipython

In [64]: res = r.get("/tcprouter/service/google")     

In [65]: decoded = json.loads(res)                    

In [66]: decoded                                      
Out[66]: 
{'Key': '/tcprouter/service/google',
 'Value': 'eyJhZGRyIjogIjE3Mi4yMTcuMTkuNDY6NDQzIiwgInNuaSI6ICJ3d3cuZ29vZ2xlLmNvbSIsICJuYW1lIjogImdvb2dsZSJ9'}


```
`Value` payload is base64 encoded because of how golang is marshaling.

```ipython
In [67]: base64.b64decode(decoded['Value'])           
Out[67]: b'{"addr": "172.217.19.46:443", "sni": "www.google.com", "name": "google"}'

```

## Examples

### Go

This example can be found at [examples/main.go](./examples/main.go)
```go

package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/abronan/valkeyrie"
	"github.com/abronan/valkeyrie/store"

	"github.com/abronan/valkeyrie/store/redis"
)

func init() {
	redis.Register()
}

type Service struct {
	Addr string `json:"addr"`
	SNI  string `json:"sni"`
	Name string `json:"bing"`
}

func main() {

	// Initialize a new store with redis
	kv, err := valkeyrie.NewStore(
		store.REDIS,
		[]string{"127.0.0.1:6379" },
		&store.Config{
			ConnectionTimeout: 10 * time.Second,
		},
	)
	if err != nil {
		log.Fatal("Cannot create store redis")
	}
	google := &Service{Addr:"172.217.19.46:443", SNI:"www.google.com", Name:"google"}
	encGoogle, _ := json.Marshal(google)
	bing := &Service{Addr:"13.107.21.200:443", SNI:"www.bing.com", Name:"bing"}
	encBing, _ := json.Marshal(bing)

	kv.Put("/tcprouter/services/google", encGoogle, nil)
	kv.Put("/tcprouter/services/bing", encBing, nil)


}

```




### Python
```python3
import base64
import json
import redis

r = redis.Redis()

def create_service(name, sni, addr):
    service = {}
    service['Key'] = '/tcprouter/service/{}'.format(name)
    record = {"addr":addr, "sni":sni, "name":name}
    json_dumped_record_bytes = json.dumps(record).encode()
    b64_record = base64.b64encode(json_dumped_record_bytes).decode()
    service['Value'] = b64_record
    r.set(service['Key'], json.dumps(service))
    
create_service('facebook', "www.facebook.com", "102.132.97.35:443")
create_service('google', 'www.google.com', '172.217.19.46:443')
create_service('bing', 'www.bing.com', '13.107.21.200:443')
            

```


If you want to test that locally you can modify `/etc/hosts`

```


127.0.0.1 www.google.com
127.0.0.1 www.bing.com
127.0.0.1 www.facebook.com

```
So your browser go to your `127.0.0.1:443` on requesting google or bing.


# dns server

Based on coredns + [redis](https://coredns.io/explugins/redis/) plugin

## build

To build we will need to add `redis` plugin to `plugin.cfg` of coredns

```bash
git clone https://github.com/coredns/coredns
cd coredns
echo 'redis:github.com/arvancloud/redis' >> plugin.cfg
make
chmod +x coredns
```

## Corefile

```
. {
    redis  {
        address 127.0.0.1:6379
    }
    forward 8.8.8.8 9.9.9.9 

}
```
We tell coredns to use `redis` plugin on address `127.0.0.1:6379` (can be customized check [redis](https://coredns.io/explugins/redis/) plugin )


## Running
`./coredns -conf Corefile`


### DNS records
 [redis](https://coredns.io/explugins/redis/) plugin

Here's an example of dns records for `example.net.`
```

# redis-cli> hgetall example.net.
#  1) "_ssh._tcp.host1"
#  2) "{\"srv\":[{\"ttl\":300, \"target\":\"tcp.example.com.\",\"port\":123,\"priority\":10,\"weight\":100}]}"
#  3) "*"
#  4) "{\"txt\":[{\"ttl\":300, \"text\":\"this is a wildcard\"}],\"mx\":[{\"ttl\":300, \"host\":\"host1.example.net.\",\"preference\": 10}]}"
#  5) "host1"
#  6) "{\"a\":[{\"ttl\":300, \"ip\":\"5.5.5.5\"}]}"
#  7) "sub.*"
#  8) "{\"txt\":[{\"ttl\":300, \"text\":\"this is not a wildcard\"}]}"
#  9) "_ssh._tcp.host2"
# 10) "{\"srv\":[{\"ttl\":300, \"target\":\"tcp.example.com.\",\"port\":123,\"priority\":10,\"weight\":100}]}"
# 11) "subdel"
# 12) "{\"ns\":[{\"ttl\":300, \"host\":\"ns1.subdel.example.net.\"},{\"ttl\":300, \"host\":\"ns2.subdel.example.net.\"}]}"
# 13) "@"
# 14) "{\"soa\":{\"ttl\":300, \"minttl\":100, \"mbox\":\"hostmaster.example.net.\",\"ns\":\"ns1.example.net.\",\"refresh\":44,\"retry\":55,\"expire\":66},\"ns\":[{\"ttl\":300, \"host\":\"ns1.example.net.\"},{\"ttl\":300, \"host\":\"ns2.example.net.\"}]}"
# redis-cli>
```

#### Example script to help with the creation of dns records on redis

```python
from sys import argv
import base64
import json
import redis

r = redis.Redis()


# redis-cli> hgetall example.net.
#  1) "_ssh._tcp.host1"
#  2) "{\"srv\":[{\"ttl\":300, \"target\":\"tcp.example.com.\",\"port\":123,\"priority\":10,\"weight\":100}]}"
#  3) "*"
#  4) "{\"txt\":[{\"ttl\":300, \"text\":\"this is a wildcard\"}],\"mx\":[{\"ttl\":300, \"host\":\"host1.example.net.\",\"preference\": 10}]}"
#  5) "host1"
#  6) "{\"a\":[{\"ttl\":300, \"ip\":\"5.5.5.5\"}]}"
#  7) "sub.*"
#  8) "{\"txt\":[{\"ttl\":300, \"text\":\"this is not a wildcard\"}]}"
#  9) "_ssh._tcp.host2"
# 10) "{\"srv\":[{\"ttl\":300, \"target\":\"tcp.example.com.\",\"port\":123,\"priority\":10,\"weight\":100}]}"
# 11) "subdel"
# 12) "{\"ns\":[{\"ttl\":300, \"host\":\"ns1.subdel.example.net.\"},{\"ttl\":300, \"host\":\"ns2.subdel.example.net.\"}]}"
# 13) "@"
# 14) "{\"soa\":{\"ttl\":300, \"minttl\":100, \"mbox\":\"hostmaster.example.net.\",\"ns\":\"ns1.example.net.\",\"refresh\":44,\"retry\":55,\"expire\":66},\"ns\":[{\"ttl\":300, \"host\":\"ns1.example.net.\"},{\"ttl\":300, \"host\":\"ns2.example.net.\"}]}"
# redis-cli>

def create_bot_record(domain="", record_type="a", records=None):
    """
    for every entry you need to comply with record format


    """
    data = {}
    records = records or []
    if r.hexists("bots.grid.tf.", domain):
        data = json.loads(r.hget("bots.grid.tf.", domain))
    if record_type in data:
        records.extend(data[record_type])
    data[record_type] = records
    r.hset("bots.grid.tf.", domain, json.dumps(data))
    
def create_a_record(domain, records):
    for rec in records:
        assert "ip" in rec
    
    return create_bot_record(domain, "a", records)

def create_aaaa_record(domain, records):
    for rec in records:
        assert "ip" in rec
    
    return create_bot_record(domain, "aaaa", records)


def create_txt_record(domain, records):
    for rec in records:
        assert "txt" in rec
    
    return create_bot_record(domain, "txt", records)


def create_ns_record(domain, records):
    for rec in records:
        assert "host" in rec
    
    return create_bot_record(domain, "ns", records)


```

## Scripts directory

### create_service.py

Helps with the creation of service for `tcprouter`

to register new service  `python3 create_service.py site1 site1.bot.testbots.grid.tf 172.17.2.5:443`


### create_coredns_site.py

```python
import create_coredns_site as c
c.create_a_record("site2.bot", [{"ip":"188.165.218.205"}])
```