## How to test New WebGateway :-

### In this test we will use 3 flists (all flists are merged with ubuntu flist ):-

##### Coredns_flist =  <https://hub.grid.tf/thabet/corednredisubuntu.flist>

##### TCPRouter_flist = <https://hub.grid.tf/thabet/routerubuntu.flist>

##### Caddy_flist = <https://hub.grid.tf/nashaatp/generic_caddy.flist>

### we can use the coredns we need to point nx record to it, here is how to do this :-

- Create a A record point to the public ip of the coredns like (ovh2.grid.tf)
![](https://github.com/threefoldtech/tf_gateway/blob/master/newg_1.png)

- Create NS record with new domain that will point to the domain (ovh2.grid.tf)
![](https://github.com/threefoldtech/tf_gateway/blob/master/newg_2.png)

#### Now Creating the containers :-

- creating CoreDns container :
  - we need coredns to have `udp|53` port open 

`zos container new  --ports='udp|53:53' --name=coredns_3bot_test --hostname=coredns --root=https://hub.grid.tf/thabet/corednredisubuntu.flis`

 - Creating TCPRouter container
  - `zos container new  --ports=80:80,443:443 --name=tcprouter_3bot_test --hostname=tcprouter --root=https://hub.grid.tf/thabet/routerubuntu.flist`
 
 - Creating websites Container :-  (with portforward to 443 to the container)
    - `zos container new  --ports=5071:443 --name=caddy1_3bot_test --hostname=caddy1 --root=https://hub.grid.tf/nashaatp/generic_caddy.flist --env='REPO_URL:https://github.com/threefoldtech/www_threefold.tech.git,REPO_BRANCH:production'`

   - `zos container new  --ports=5073:443 --name=caddy2_3bot_test --hostname=caddy2 --root=https://hub.grid.tf/nashaatp/generic_caddy.flist --env='REPO_URL:https://github.com/Incubaid/www_incubaid,REPO_BRANCH:production'`
   
   ``` 
   NOTE the webserver(caddy) has to creat it's own certificate 
   Like in caddy we add in caddyfile option to create the cert 
   ```
   -  caddyfile example :- 

   ```

   https://site1.bot.testbots.grid.tf {
    bind 0.0.0.0
    gzip
    root .
    tls enter_email_addr_here
    git {
        repo https://github.com/Incubaid/www_incubaid.git
        branch production
        hook /webhook "zU3687$aJEb6"
        key ~/.ssh/id_rsa
        path . 
       }
   }

- now adding info to new CoreDns Container :-
   
  - ssh to the coredns continer, and download script that will add keys to redis 
  
`wget https://raw.githubusercontent.com/threefoldtech/tf_gateway/master/scripts/create_coredns_site.py`
 
```
root@coredns:/tmp# python3
Python 3.6.5 (default, Apr  1 2018, 05:46:30) 
[GCC 7.3.0] on linux
Type "help", "copyright", "credits" or "license" for more information.
>>> 
>>> import create_coredns_site as c
>>> c.create_a_record("site1.bot", [{"ip":"188.165.218.205"}])
>>> c.create_a_record("site2.bot", [{"ip":"188.165.218.205"}])


```

 - Adding info to TCPRouter Contianer
   - ssh to the TCPRouter continer, and download script that will add keys to redis
```
 wget https://raw.githubusercontent.com/threefoldtech/tf_gateway/master/scripts/create_service.py

root@tcprouter:~# python3 create_service.py site1bot site2.bot.testbots.grid.tf 10.102.90.219:5552
root@tcprouter:~# python3 create_service.py site2bot site2.bot.testbots.grid.tf 10.102.90.219:5552

```


##### Make sure that redis is running before adding the keys
