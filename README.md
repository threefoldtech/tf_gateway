# TFGateway ![Tests and Build](https://github.com/threefoldtech/tfgateway/workflows/Tests%20and%20Build/badge.svg)

## Content
- [Architecture](#architecture)
- [Installation](#Installation)
  - [Prerequisities and software packages needed](#software-prerequisites)
  - [Configuration of the software packages](#configuring-the-software)


## Architecture

The high level architecture of the TF Gateway can be visualised as follows:
![architecture_overview](docs/asset/overview.png)

The TFGateway works by reading reservation details from the TF Explorer. It then converts this information into a configuration readable by the TCP Router server and CoreDNS and stores these in a redis server.

Both CoreDNS and the TCP router are checking for changes to their configuration information in redis.  Everytime the configuration stored in redis changes this configuration is made active.


## Installation

### Software prerequisites 

Bofore we begin there are a number of requirements that need to be met before we can install:
- installation requires a linux operating system installed on a (bare metal) server.
- IPv4 to IPv6 masquerading. The host needs to masquerade the ipv6 traffic going out `ip6tables -t nat -A POSTROUTING -o eth0 -j MASQUERADE`
- Installed redis server (linux distribution)
- coredns-redis software [(GitHub)](https://github.com/threefoldtech/coredns-redis)
 - tcprouter software [(GitHub)](https://github.com/threefoldtech/tcprouter)
- tfgateway software [(GitHub)](https://github.com/threefoldtech/tfgateway)

After installed the minimal server version of the linuc distribution of your choice, the following steps need to be taken in order to get the basic software installed.  Here we use the [ubuntu server 20.10](https://ubuntu.com/download/server).

1. After installing the ubuntu server create a systemd driven system, update and upgrade it / install some stuff (archlinux users and other distros supported know what to do ;-) )

2. Update the installed system software database and upgrade all packages updated.
 
   ```shell
    # When used in a shell script
    export DEBIAN_FRONTEND=noninteractive

    # Update and upgrade
    apt -y update
    apt -y upgrade
    apt -y autoremove
    ```

3. Configure remote access to the server on 2 ports:

    ```shell
    # Enable ssh access on port 22 and 34022
    sed -ie "s/^#Port.*/Port 22\nPort 34022\n/" /etc/ssh/sshd_config
    systemctl restart sshd
    ```

4. Install ```ufw`` if not installed by the defualt base os installation

    ```shell
    apt -o Dpkg::Options::='--force-confold' --force-yes -fuy install ufw
    ufw allow 34022/tcp
    echo | ufw enable
    systemctl enable ufw --now
    ```

5. You might want to install `docker` for testbedding or other reasons. This is not needed for bare TFGateway operations

    ```shell
    # OPTIONAL: This is not required for general operation of a TF Gateway.

    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -
    add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"

    apt update
    apt-get install docker-ce docker-ce-cli containerd.io -y
    systemctl enable docker --now
    ```

6. Install `mongodb`.  The required server version is not part yet of the mainstream (ubuntu) server distrinbution, so we add the mongo repository and get ids from there.
    ```
    wget -qO - https://www.mongodb.org/static/pgp/server-4.2.asc | sudo apt-key add -
    echo "deb [ arch=amd64,arm64 ] https://repo.mongodb.org/apt/ubuntu bionic/mongodb-org/4.2 multiverse" | sudo tee /etc/apt/sources.list.d/mongodb-org-4.2.list
    apt-get update
    apt-get install -y mongodb-org
    ```

7. Install `redis`, the key value store.
    ```shell
    apt install redis
    ```

8. For networking, I assume you know your way ;-).  This is all for basic installation of required software components.


### Configuring the software

Any container (or containers) running a particular a service if the TF Grid is (are) not directly accessible from the Internet The gateway allows you to expose serivces on the internet by doing several things.

First, you want your service to be reachable by a DNS name (fqdn) in a domain.
Secondly, the service you want to expose needs to be reachable.

TFGateway manages a dns server (coredns) and a tcp proxy that accepts connections from clients an forwards it to your container / application. Configuration of both services happen by a redis backend that is used as a queue, so that coredns and the tcp proxy can add/remove entries in their live configuration.

So all 4 [pieces of software](#software-prerequisites) need to be running on a node, with some ip/tcp ports properly opened up to handle traffic.


#### Configuring of the services

Your system might be running resolved and using port 53. This port is needed for coredns, so we have to stop resolved.  Also the port 6379 (redis) will be used as soon as the default redis starts.
1. Update the system and create configration file directories

    ```shell
    apt update -y
    apt install redis-server redis-tools -y
    mkdir -p /etc/coredns
    mkdir -p /etc/tcprouter
    ```

2. Create the identity seed file

    File: ```/etc/identity.seed```

    ```shell
    cat << EOF > /etc/identity.seed
    "1.1.0"{"mnemonic":"$$3BOT_WORDS","threebotid":$$3BOT_ID}
    EOF
    ```
  
    Make sure to replace the `$$3BOT_WORDS` and `$$3BOT_ID` with your 3bot identity that manages (owns)your farm.

    Identity creation (configuration) done.

3. Before changing the configuration of resolved, coredns and redis we have to:
    - stop and disable redis unit 
    ```shell
    systemctl stop redis
    ```
    - stop and disable systemd-resolved unit 
    ```shell
    systemctl stop systemd-resolved
    ```
    Standard Redis service stopped and disabled

4. Then we will have to edit the redis service.

    a. Configuration file: `/etc/tfredis.conf`.  

    If this file does not exists, here's how you can create it:
    ```shell
    cat << EOF > /etc/tfredis.conf
    # START of the content of the file
      bind 127.0.0.1
    # END of the content of the file
    EOF
    ```

    b. Service configuration: `/etc/systemd/system/tfredis.service`

    If this file does not exists, here's how you can create it:
    ```shell
    cat << EOF > /etc/systemd/system/tfredis.service
    # START of the content of the file
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
    # END of the content of the file
    EOF
    ```
    Redis service configuration done.

5. Next is the `coreDNS` service. Edit the config file: `/etc/coredns/Corefile`

    a. Configuration file: `/etc/coredns/Corefile`.  
    ```shell
    cat << EOF > /etc/coredns/Corefile
    # START of the content of the file
    {
      redis
      {
        address 127.0.0.1:6379
      }
    }
    # END of the content of the file
    EOF
    ```
    b. Service configuration: `/etc/systemd/system/coredns.service`

    If this file does not exists, here's how you can create it:
    ```shell
    cat << EOF > /etc/systemd/system//etc/systemd/system/coredns.service
    # START of the content of the file
    [Unit]
    Description=CoreDNS
    After=network.target
    Requires=tfredis.service

    [Service]
    ExecStart=/usr/local/bin/coredns -conf /etc/coredns/Corefile
    Type=simple
    Restart=on-failure
    MemoryAccounting=true
    MemoryHigh=800M
    MemoryMax=1G

    [Install]
    WantedBy=multi-user.target
    # END of the content of the file
    EOF
    ```
    CoreDNS service configuration done.

6. Next if the `tcprouter` service
  
    a. Configuration file: `/etc/tcprouter/router.toml`

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

    b. Service configuration: `/etc/systemd/system/tcprouter.service`
    
    ```shell
    cat << EOF > /etc/systemd/system//etc/systemd/system/tcprouter.service
    # START of the content of the file
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
    # END of the content of the file
    EOF
    ```
    
    TCPRouter service configuration done.

7. Last thing that needs doing:  the `tfgateway` sevice.  

    a. Service configuration: `/etc/systemd/system/tcprouter.service`
    
    ```shell
    cat << EOF > /etc/systemd/system//etc/systemd/system/tcprouter.service
    # START of the content of the file
    [Unit]
    Description=TCP router server
    After=network.target

    [Service]
    ExecStartPre=/bin/bash -c "/bin/systemctl set-environment ip=$(/sbin/ip r get 1.1.1.1 | awk '{print $7}')"
    ExecStartPre=/bin/bash -c "/bin/systemctl set-environment hostname=$(/bin/hostname)"
    ExecStartPre=/bin/bash -c "/bin/systemctl set-environment subdom=tfgw$${hostname/tf-gateway}"
    ExecStart=/usr/local/bin/tfgateway --seed /etc/identity.seed --nameservers ${hostname}.gateway.tf --endpoint ${ip}:3443 --domains ${subdom}.gateway.tf --domains ${subdom}.3x0.me --domains ${subdom}.ava.tf --domains ${subdom}.base.tf --farm 1

    Type=simple
    Restart=on-failure
    MemoryAccounting=true
    MemoryHigh=800M
    MemoryMax=1G

    [Install]
    WantedBy=multi-user.target
    # END of the content of the file
    EOF
    ```

  TCPRouter condiguration done!



```shell
systemctl stop redis
systemctl stop systemd-resolved
systemctl start tfredis && systemctl start coredns && systemctl start tcprouter && systemctl start tfgateway

```

### Delegation of domains

If you want people to be able to delegate domain to the TFGateway. User needs to create a `NS record` pointing to the a domain of the TFGateway. Which means you need to have an `A record` pointing to the IP of the TFGateway and use the `--nameservers` flag when starting the TFGateway.

```
## Core TFGateway  nodes

There are 7 nodes that have the tfgateway installed, 6 of them are DO nodes, one is a separate machine in the freefarm env.

<!-- URL is only accesible by ThreeFold staff, shoul we mentioned it? -->

https://cloud.digitalocean.com/projects/92a99fbe-5fa1-48f0-b088-1d93a56ac817/resources?i=68c689

  -  tf-gateway-prod-01: ssh root@159.89.181.109 -p 34022 (Ubuntu 18.04)
  -  tf-gateway-prod-02: ssh root@167.71.58.136 -p 34022 (Ubuntu 18.04)
  -  tf-gateway-prod-03: ssh root@161.35.35.103 -p 34022 (Ubuntu 18.04)
  -  tf-gateway-prod-04: ssh root@161.35.88.77 -p 34022 (Ubuntu 18.04)
  -  tf-gateway-prod-05: ssh root@64.225.33.77 -p 34022 (Ubuntu 18.04)
  -  tf-gateway-prod-06: ssh root@159.89.181.109 -p 34022 (Ubuntu 18.04)
  -  tf-gateway-prod-07: ssh root@185.69.166.121 -p 34022 (ArchLinux)

the nodes are all configured the same way, where systemd handles the daemons necessary to run the nodes.

### Supported primitives

- Delegation of domain
- Creation of A or AAAA DNS records
- HTTP(S) proxy
- Reverse tunnel TCP proxy: https://github.com/threefoldtech/tcprouter#reverse-tunneling
- Gateway IPv4 to IPv6