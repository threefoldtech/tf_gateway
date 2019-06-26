from unittest import TestCase
from uuid import uuid4
from zeroos.core0.client import Client
from loguru import logger
import configparser
import redis
import base64
import json
import redis

logger.add("tf_gateway.log", format="{time} {level} {message}", level="INFO")


class BaseTest(TestCase):
    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        config = configparser.ConfigParser()
        config.optionxform = str
        config.read("config.ini")
        self.jwt = config["main"]["jwt"]
        self.deploy = config["main"]["deploy"]
        self.destroy = config["main"]["destroy"]
        self.coredns_node_ip = config["tf_gateway"]["node_ip"]
        self.coredns_redis_port = config["tf_gateway"]["redis_port"]

    @classmethod
    def setUpClass(cls):
        self = cls()
        if self.deploy == "True":
            # deploy tf gateway container
            cl = Client(host=self.coredns_node_ip, password=self.jwt)
            cls.tf_gateway_id = cl.container.create(
                name="test_tf_gateway",
                root_url="https://hub.grid.tf/tf-autobuilder/threefoldtech-tf_gateway-tf-gateway-master.flist",
                nics=[{"type": "default", "name": "defaultnic", "id": " None"}],
                port={"53|udp": 53, "443": 443, self.coredns_redis_port: 6379},
            ).get()

    @classmethod
    def tearDownClass(cls):
        self = cls()
        if self.destroy == "True":
            cl = Client(host=self.coredns_node_ip, password=self.jwt)
            cl.container.terminate(self.tf_gateway_id)

    def setUp(self):
        pass

    def tearDown(self):
        pass

    def log(self, msg):
        logger.info(msg)

    def random_string(self):
        return str(uuid4())[:10]

    def backup_file(self, path):
        with open(path, "r") as f:
            backup = f.read()
        return backup

    def wirte_file(self, path, content, rw="w+"):
        with open(path, rw) as f:
            f.write(content)

    def delete_redis_record(self, name):
        r = redis.Redis()
        r.delete(name)

    def delete_tcp_record(self, name):
        if not name.startswith("/tcprouter/service/"):
            name = "/tcprouter/service/{}".format(name)
        self.delete_redis_record(name)

    def delete_all_tcp_records(self):
        r = redis.Redis()
        keys = r.keys()
        for key in keys:
            try:
                key = key.decode()
            except Exception:
                continue
            if key.startswith("/tcprouter/service"):
                self.delete_redis_record(key)

    def delete_coredns_record(self, name):
        r = redis.Redis()
        r.hdel("bots.grid.tf.", name)

    def delete_all_coredns_records(self):
        self.delete_redis_record("bots.grid.tf.")
