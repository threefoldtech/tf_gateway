## Functionality

So how does all that fit together?

There are 2 parts:
  - resolve the dns name to an ip address, so that the client request arrives at the gatway that has been configured to accept requests for the specified fqdn (e.g. `www.myspecialname.gateway.tf`)
  - forwarding traffic from a client to the service that you want to expose.

### DNS

Let's start by playing a dns server for your client, that wants to reach www.myspecialname.gateway.tf.

Your client has a dns server (Recursive DNS server,Rdns) configured in it's IP configuration.
A Recursive DNS server handles the queries for your client, ultimately returning an IP address for and FQDN.
That RDSN server ahs a list of all Top-Level-Domain nameservers it can send queries to.

Client -> Rdns: hey, Rdns, I want the IP address of `www.myspecialname.gateway.tf`

Rdns -> TLD server : Hey, `tld` server, what nameserver can I reach for a `.tf` domain

TLD Server -> Rdns : here is an ip address where you can ask

Rdns -> `.tf` nameserver: Hey, `.tf` nameserver, I'm looking for a nameserver that can answer questions about `gateway.tf`

`.tf` nameserver -> Rdns : you can go ask the server with that ip.

Rdns -> `gateway.tf` nameserver: hey `gateway.tf` server, I want the IP address of `www.myspecialname.gateway.tf`.

`gateway.tf` nameserver -> Rdns: Uh-oh, I'm not authoritative for that fqdn, actually `myspecialname` is a subdomain with it's own nameserver, go ask him, here is the ip address

Rdns -> `myspecialname.gateway.tf` nameserver : Hey `myspecialname.gateway.tf`, can you give me the IP address of `www` in your domain ?

`myspecialname.gateway.tf` nameserver -> Rdns: well of course, Rdns, here is the IP address.


Now the DNS server for `myspecialname.gateway.tf` is that coredns server on the tfgatways. Entries are added by a reservation on the explorer, which the gets picked up by the gateway and set up, zo that above process for resolving the fqdn (`www.myspecialname.gateway.tf`) to an ip address for the client.


## TCP Proxy

The TCP Proxy is special in the sense that it contains 2 parts:
  - a server to accept connections from external clients (like your web browser)
  - and a client-server part that effectively forwards the above connection stream towards the exposed service

The client-server part seems a bit convoluted, but please bear with me:

The moment you want to expose a service, the grid adds a container in the same user network of the running service container, and starts a proxy, that is also a client towards the tfgateway server.
The tfgateway's tcp proxy connects the outside listener with the the proxy client in that addon container, and as such the the proxy in the addon container can forward the queries towards the service.