### Before running the tests:
1- Install [coredns](https://github.com/coredns/coredns#compilation-from-source), [tcprouter](https://github.com/xmonader/tcprouter#install), redis and tmux.

2- Install requirement packages `pip3 install -r requirements`.

**Note:** it's preferable to run this test inside docker container and take backup for `/etc/hosts`and `/etc/resolv.conf` in case of running local tests.

### How to run:
#### Local tests:
```bash
nosetests-3.4  local/testcases.py
```
#### Rmote tests:
**Note:** it's not done yet.
```bash
nosetests-3.4  remote/testcases.py
```