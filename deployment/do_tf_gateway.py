import os
import gevent
import traceback
from Jumpscale import j


class Do_tf:
    def __init__(self):
        do_token = os.environ.get("TOKEN")
        self.do_client = j.clients.digitalocean.get("DO_CL", token_=do_token)

    def create_vm(self, name):
        created = self.do_client.droplet_create(name="", sshkey="Peter", image="zero_os", size_slug="s-1vcpu-1gb")
        if created and len(created) >= 2:
            ip = created[1].addr
        return ip

    def create_container(self, name, ip, flist, nics, ports, env=None, host_network=True):
        zos_node = j.clients.zos.get(name="zos_cl", host=ip)
        container_id = zos_node.client.container.create(
            name=name, root_url=flist, host_network=host_network, nics=nics, port=ports, env=env
        ).get()
        return container_id

    def install_web_gw(self, name, flist, ports, env=None):
        ip = self.create_vm(name=name)
        container_id = self.create_container(
            name=name,
            ip=ip,
            flist=flist,
            nics=[{"type": "default", "name": "defaultnic", "id": " None"}],
            ports=ports,
            env=env
        )
        if container_id:
            print("{} has been deployed.".format(name))
        return ip

    def on_error(self, fut):
        print("something happened.")

        try:
            fut.get()
        except Exception as e:
            print("ERROR HAPPEEND: ", str(e) + traceback.format_exc())

    def on_value(self, fut):
        print("SUCCESS: ", fut.value)


if __name__ == "__main__":

    gw_flist = "https://hub.grid.tf/tf-autobuilder/threefoldtech-tf_gateway-tf-gateway-master.flist"
    jsx_flist = "https://hub.grid.tf/tf-autobuilder/threefoldtech-jumpscaleX-development.flist"
    do_tf = Do_tf()
    master_ip = do_tf.install_web_gw(name="jsx_master", flist=jsx_flist, ports={"80": 80, "443": 443, "4000": 6379})

    futures = []
    for i in range(5):
        name = "tf_gateway_{}".format(i)
        f = gevent.spawn(do_tf.install_web_gw(), name=name, flist=gw_flist,
                         ports={"53|udp": 53, "443": 443, "4000": 6379}, env={"MASTER_REDIS_IP": master_ip})
        f.link_exception(do_tf.on_error)
        f.link_value(do_tf.on_value)
        futures.append(f)

    gevent.joinall(futures, raise_error=False)
