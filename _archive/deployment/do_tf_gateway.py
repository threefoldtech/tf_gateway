import os
import gevent
import traceback
from Jumpscale import j


class Do_tf:
    def __init__(self):
        token = os.environ.get("TOKEN")
        self.do_client = j.clients.digitalocean.get("DO_CL", token_=token)

    def create_vm(self, name):
        create = self.do_client.droplet_create(name="", sshkey="Peter", image="zero_os", size_slug="s-1vcpu-1gb")
        ip = create[1].addr
        return ip

    def create_container(self, name, ip, flist, nics, ports, env=None, host_network=True):
        zos_node = j.clients.zos.get(name="zos_cl", host=ip)
        container_id = zos_node.client.container.create(
            name=name, root_url=flist, host_network=host_network, nics=nics, port=ports, env=env
        ).get()
        return container_id

    def deploy_tf_gateway(self, name, env=None):
        ip = self.create_vm(name=name)
        container_id = self.create_container(
            name=name,
            ip=ip,
            flist="https://hub.grid.tf/tf-autobuilder/threefoldtech-tf_gateway-tf-gateway-master.flist",
            nics=[{"type": "default", "name": "defaultnic", "id": " None"}],
            ports={"53|udp": 53, "443": 443, "4000": 6379},
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
    do_tf = Do_tf()
    master_ip = do_tf.deploy_tf_gateway(name="master")

    futures = []
    for i in range(1, 6):
        name = "slave_{}".format(i)
        f = gevent.spawn(do_tf.deploy_tf_gateway, name=name, env={"MASTER_REDIS_IP": master_ip})
        f.link_exception(do_tf.on_error)
        f.link_value(do_tf.on_value)
        futures.append(f)

    gevent.joinall(futures, raise_error=False)
