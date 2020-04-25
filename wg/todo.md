# vpn-server

## boot

1. generate identity seed
2. derivate ipv6 fd::/8 prefix from seed
3. create network namespace for router
3.1 configure nftable/iptable in router namespace
4. create wg interface, move into router namespace

## add new user

1. receive public key from client
2. allocate fd:: address from pool
3. add peer to router config
4. generate wg-config for user
5. reload router wg iface
6. send result
