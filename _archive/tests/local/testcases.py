from ..base_test import BaseTest
from tf_gateway.scripts import create_coredns_site as coredns
from tf_gateway.scripts import create_service as tcprouter
import unittest
import redis, requests, time
import os


class TcpRouterTests(BaseTest):
    @classmethod
    def setUpClass(cls):
        os.system("tmux new-session -d -s redis 'redis-server';")
        os.system("tmux new-session -d -s tcprouter 'tcprouter router.toml';")
        time.sleep(5)

    @classmethod
    def tearDownClass(cls):
        os.system("pkill tcprouter")
        os.system("pkill redis-server")
        if os.path.exists("dump.rdb"):
            os.system("rm dump.rdb")

    def setUp(self):
        super().setUp()
        self.backup = self.backup_file(path="/etc/hosts")

    def tearDown(self):
        self.wirte_file(path="/etc/hosts", content=self.backup)
        self.delete_all_tcp_records()
        super().tearDown()

    def test001_add_record(self):
        """
        TFGW-001
        *Test case for adding a record to tcprouter and try to access a website through it. *

        **Test Scenario:**
        
        #. Create a tcprouter record.
        #. Try to access the website(facebook), should success.
        """
        self.log("Create a tcprouter record.")
        hosts = "127.0.0.1       www.facebook.com\n"
        self.wirte_file(path="/etc/hosts", content=hosts)
        tcprouter.create_service(name="facebook", sni="www.facebook.com", addr="102.132.97.35:443")
        time.sleep(11)

        self.log("Try to access the website(facebook), should success.")
        r = requests.get("https://www.facebook.com")
        self.assertEquals(r.status_code, requests.codes.ok)
        self.assertIn("facebook", r.text)

    def test002_add_more_records(self):
        """
        TFGW-002
        *Test case for adding records to tcprouter and try to access websites through them. *

        **Test Scenario:**

        #. Create a tcprouter record for facebook.
        #. Create a tcprouter record for bing.
        #. Create a tcprouter record for google.
        #. Try to access facebook, bing and google through tcprouter records, should success.
        """
        self.log("Create a tcprouter record for facebook.")
        hosts = "127.0.0.1       www.facebook.com\n"
        self.wirte_file(path="/etc/hosts", content=hosts)
        tcprouter.create_service(name="facebook", sni="www.facebook.com", addr="102.132.97.35:443")

        self.log("Create a tcprouter record for bing.")
        hosts = "127.0.0.1       www.bing.com\n"
        self.wirte_file(path="/etc/hosts", content=hosts, rw="a")
        tcprouter.create_service(name="bing", sni="www.bing.com", addr="13.107.21.200:443")

        self.log("Create a tcprouter record for google.")
        hosts = "127.0.0.1       www.google.com\n"
        self.wirte_file(path="/etc/hosts", content=hosts, rw="a")
        tcprouter.create_service(name="google", sni="www.google.com", addr="172.217.19.46:443")
        time.sleep(11)

        self.log("Try to access facebook, bing and google through tcprouter records, should success.")
        req = requests.get("https://www.facebook.com")
        self.assertEquals(req.status_code, requests.codes.ok)
        self.assertIn("facebook", req.text)

        req = requests.get("https://www.bing.com")
        self.assertEquals(req.status_code, requests.codes.ok)
        self.assertIn("bing", req.text)

        req = requests.get("https://www.google.com")
        self.assertEquals(req.status_code, requests.codes.ok)
        self.assertIn("google", req.text)

    def test003_modify_records(self):
        """
        TFGW-003
        *Test case for modifying tcprouter records and try to access websites through them*

        **Test Scenario:**

        #. Create a tcprouter record for google.
        #. Try to access google, should success.
        #. Modify google tcprouter record with facebook ip.
        #. Try to access facebook through google record, should success.
        """
        self.log("Create a tcprouter record for google.")
        hosts = "127.0.0.1       www.google.com\n"
        self.wirte_file(path="/etc/hosts", content=hosts, rw="a")
        tcprouter.create_service(name="google", sni="www.google.com", addr="172.217.19.46:443")
        time.sleep(11)

        self.log("Try to access google, should success.")
        req = requests.get("https://www.google.com")
        self.assertEquals(req.status_code, requests.codes.ok)
        self.assertIn("google", req.text)

        self.log("Modify google tcprouter record with facebook ip.")
        tcprouter.create_service(name="google", sni="www.google.com", addr="102.132.97.35:443")
        time.sleep(11)

        self.log("Try to access facebook through google record, should success.")
        with self.assertRaises(requests.exceptions.SSLError) as e:
            requests.get("https://www.google.com")

        error = e.exception.args[0].args[0]
        self.assertIn("facebook.com", error)

    @unittest.skip("https://github.com/xmonader/tcprouter/issues/3")
    def test004_delete_records(self):
        """
        TFGW-004
        * Test case for deleting tcprouter records and making sure that websites can't be accessed. *

        **Test Scenario:**

        #. Create a tcprouter record for facebook.
        #. Create a tcprouter record for bing.
        #. Create a tcprouter record for google.
        #. Delete google and facebook records from tcprouter.
        #. Try to access facebook or google, should fail.
        #. Try to access bing, should success.
        """
        self.log("Create a tcprouter record for facebook.")
        hosts = "127.0.0.1       www.facebook.com\n"
        self.wirte_file(path="/etc/hosts", content=hosts)
        tcprouter.create_service(name="facebook", sni="www.facebook.com", addr="102.132.97.35:443")

        self.log("Create a tcprouter record for bing.")
        hosts = "127.0.0.1       www.bing.com\n"
        self.wirte_file(path="/etc/hosts", content=hosts, rw="a")
        tcprouter.create_service(name="bing", sni="www.bing.com", addr="13.107.21.200:443")

        self.log("Create a tcprouter record for google.")
        hosts = "127.0.0.1       www.google.com\n"
        self.wirte_file(path="/etc/hosts", content=hosts, rw="a")
        tcprouter.create_service(name="google", sni="www.google.com", addr="172.217.19.46:443")

        self.log("Delete google and facebook records from tcprouter.")
        self.delete_tcp_record("facebook")
        self.delete_tcp_record("google")
        time.sleep(11)

        self.log("Try to access facebook or google, should fail.")
        with self.assertRaises(requests.exceptions.ReadTimeout):
            requests.get("https://www.facebook.com", timeout=10)

        with self.assertRaises(requests.exceptions.ReadTimeout):
            requests.get("https://www.google.com", timeout=10)

        self.log("Try to access bing, should success.")
        r = requests.get("https://www.bing.com")
        self.assertEquals(r.status_code, requests.codes.ok)
        self.assertIn("bing", r.text)

    def test005_catch_all_record(self):
        """
        TFGW-005
        * Test case for catch_all record of tcprouter. *

        **Test Scenario:**

        #. Create a tcprouter record for bing.
        #. Create catch_all tcprouter record for facebook.
        #. Try to access facebook using catch_all record, should success.
        #. Try to access bing, should be accessed not facebook.
        """
        self.log("Create a tcprouter record for bing.")
        hosts = "127.0.0.1       www.bing.com\n"
        self.wirte_file(path="/etc/hosts", content=hosts, rw="a")
        tcprouter.create_service(name="bing", sni="www.bing.com", addr="13.107.21.200:443")

        self.log("Create catch_all tcprouter record for facebook.")
        hosts = "127.0.0.1       www.facebook.com\n"
        self.wirte_file(path="/etc/hosts", content=hosts)
        tcprouter.create_service(name="CATCH_ALL", sni="CATCH_ALL", addr="102.132.97.35:443")
        time.sleep(11)

        self.log("Try to access facebook using any subdomain, should success.")
        r = requests.get("https://www.facebook.com")
        self.assertEquals(r.status_code, requests.codes.ok)
        self.assertIn("facebook", r.text)

        self.log("Try to access bing, should be accessed not facebook.")
        r = requests.get("https://www.bing.com")
        self.assertEquals(r.status_code, requests.codes.ok)
        self.assertIn("bing", r.text)


class TFGWLocalTests(BaseTest):
    @classmethod
    def setUpClass(cls):
        self = cls()
        cls.backup = self.backup_file(path="/etc/resolv.conf")
        self.wirte_file(path="/etc/resolv.conf", content="nameserver localhost")
        os.system("tmux new-session -d -s redis 'redis-server';")
        time.sleep(2)
        coredns.create_a_record("facebook", [{"ip": "127.0.0.1"}])
        os.system("tmux new-session -d -s coredns '/usr/bin/coredns -conf corefile';")
        os.system("tmux new-session -d -s tcprouter 'tcprouter router.toml';")
        time.sleep(5)

    @classmethod
    def tearDownClass(cls):
        self = cls()
        self.wirte_file(path="/etc/resolv.conf", content=self.backup)
        os.system("pkill coredns")
        os.system("pkill tcprouter")
        os.system("pkill redis-server")
        if os.path.exists("dump.rdb"):
            os.system("rm dump.rdb")
        time.sleep(2)

    def setUp(self):
        # should has coredns record.
        pass

    def tearDown(self):
        # should remove all coredns and tcprouter records.
        self.delete_all_tcp_records()
        self.delete_all_coredns_records()

    def test001_add_record(self):
        """
        TFGW-006
        *Test case for adding a record to (coredns and tcprouter) and try to access a website through them. *

        **Test Scenario:**
        
        #. Create a coredns record.
        #. Create a tcprouter record.
        #. Try to access the website(facebook), should success.
        """
        self.log("Create a coredns record.")
        coredns.create_a_record("facebook", [{"ip": "127.0.0.1"}])

        self.log("Create a tcprouter record.")
        tcprouter.create_service(name="facebook", sni="facebook.bots.grid.tf", addr="102.132.97.35:443")
        time.sleep(11)

        self.log("Try to access the website(facebook), should success.")
        with self.assertRaises(requests.exceptions.SSLError) as e:
            requests.get("https://facebook.bots.grid.tf")

        error = e.exception.args[0].args[0]
        self.assertIn("facebook.com", error)

    def test002_add_more_records(self):
        """
        TFGW-007
        *Test case for adding records to (coredns and tcprouter) and try to access websites through them. *

        **Test Scenario:**

        #. Create (coredns and tcprouter) records for facebook.
        #. Create (coredns and tcprouter) records for bing.
        #. Create (coredns and tcprouter) records for google.
        #. Try to access facebook, bing and google through (coredns and tcprouter) records, should success.
        """
        self.log("Create (coredns and tcprouter) records for facebook.")
        coredns.create_a_record("facebook", [{"ip": "127.0.0.1"}])
        tcprouter.create_service(name="facebook", sni="facebook.bots.grid.tf", addr="102.132.97.35:443")

        self.log("Create (coredns and tcprouter) records for bing.")
        coredns.create_a_record("bing", [{"ip": "127.0.0.1"}])
        tcprouter.create_service(name="bing", sni="bing.bots.grid.tf", addr="13.107.21.200:443")

        self.log("Create (coredns and tcprouter) records for google.")
        coredns.create_a_record("google", [{"ip": "127.0.0.1"}])
        tcprouter.create_service(name="google", sni="google.bots.grid.tf", addr="172.217.19.46:443")
        time.sleep(11)

        self.log("Try to access facebook, bing and google through tcprouter records, should success.")
        with self.assertRaises(requests.exceptions.SSLError) as e:
            requests.get("https://facebook.bots.grid.tf")

        error = e.exception.args[0].args[0]
        self.assertIn("facebook.com", error)

        with self.assertRaises(requests.exceptions.SSLError) as e:
            requests.get("https://bing.bots.grid.tf")

        error = e.exception.args[0].args[0]
        self.assertIn("bing.com", error)

        with self.assertRaises(requests.exceptions.SSLError) as e:
            requests.get("https://google.bots.grid.tf")

        error = e.exception.args[0].args[0]
        self.assertIn("google.com", error)

    def test003_modify_records(self):
        """
        TFGW-008
        *Test case for modifying tcprouter records and try to access websites through them*

        **Test Scenario:**

        #. Create a tcprouter record for google.
        #. Try to access google, should success.
        #. Modify google tcprouter record with facebook ip.
        #. Try to access facebook through google record, should success.
        """
        self.log("Create a tcprouter record for google.")
        coredns.create_a_record("google", [{"ip": "127.0.0.1"}])
        tcprouter.create_service(name="google", sni="google.bots.grid.tf", addr="172.217.19.46:443")
        time.sleep(11)

        self.log("Try to access google, should success.")
        with self.assertRaises(requests.exceptions.SSLError) as e:
            requests.get("https://google.bots.grid.tf")

        error = e.exception.args[0].args[0]
        self.assertIn("google.com", error)

        self.log("Modify google tcprouter record with facebook ip.")
        tcprouter.create_service(name="google", sni="google.bots.grid.tf", addr="102.132.97.35:443")
        time.sleep(11)

        self.log("Try to access facebook through google record, should success.")
        with self.assertRaises(requests.exceptions.SSLError) as e:
            requests.get("https://google.bots.grid.tf")

        error = e.exception.args[0].args[0]
        self.assertIn("facebook.com", error)

    @unittest.skip("https://github.com/xmonader/tcprouter/issues/3")
    def test004_delete_tcprouter_records(self):
        """
        TFGW-009
        * Test case for deleting tcprouter records and making sure that websites can't be accessed. *

        **Test Scenario:**

        #. Create (coredns and tcprouter) records for facebook.
        #. Create (coredns and tcprouter) records for bing.
        #. Create (coredns and tcprouter) records for google.
        #. Delete google and facebook records from tcprouter.
        #. Try to access facebook or google, should fail.
        #. Try to access bing, should success.
        """
        self.log("Create (coredns and tcprouter) records for facebook.")
        coredns.create_a_record("facebook", [{"ip": "127.0.0.1"}])
        tcprouter.create_service(name="facebook", sni="facebook.bots.grid.tf", addr="102.132.97.35:443")

        self.log("Create (coredns and tcprouter) records for bing.")
        coredns.create_a_record("bing", [{"ip": "127.0.0.1"}])
        tcprouter.create_service(name="bing", sni="bing.bots.grid.tf", addr="13.107.21.200:443")

        self.log("Create (coredns and tcprouter) records for google.")
        coredns.create_a_record("google", [{"ip": "127.0.0.1"}])
        tcprouter.create_service(name="google", sni="google.bots.grid.tf", addr="172.217.19.46:443")
        time.sleep(11)

        self.log("Delete google and facebook records from tcprouter.")
        self.delete_tcp_record("facebook")
        self.delete_tcp_record("google")
        time.sleep(11)

        self.log("Try to access facebook or google, should fail.")
        with self.assertRaises(requests.exceptions.ReadTimeout):
            requests.get("https://facebook.bots.grid.tf", timeout=10)

        with self.assertRaises(requests.exceptions.ReadTimeout):
            requests.get("https://google.bots.grid.tf", timeout=10)

        self.log("Try to access bing, should success.")
        with self.assertRaises(requests.exceptions.SSLError) as e:
            requests.get("https://bing.bots.grid.tf")

        error = e.exception.args[0].args[0]
        self.assertIn("bing.com", error)

    def test005_delete_coredns_records(self):
        """
        TFGW-010
        * Test case for deleting coredns records and making sure that websites can't be accessed. *

        **Test Scenario:**

        #. Create (coredns and tcprouter) records for facebook.
        #. Create (coredns and tcprouter) records for bing.
        #. Create (coredns and tcprouter) records for google.
        #. Delete google and facebook records from self.
        #. Try to access facebook or google, should fail.
        #. Try to access bing, should success.
        """
        self.log("Create (coredns and tcprouter) records for facebook.")
        coredns.create_a_record("facebook", [{"ip": "127.0.0.1"}])
        tcprouter.create_service(name="facebook", sni="facebook.bots.grid.tf", addr="102.132.97.35:443")

        self.log("Create (coredns and tcprouter) records for bing.")
        coredns.create_a_record("bing", [{"ip": "127.0.0.1"}])
        tcprouter.create_service(name="bing", sni="bing.bots.grid.tf", addr="13.107.21.200:443")

        self.log("Create (coredns and tcprouter) records for google.")
        coredns.create_a_record("google", [{"ip": "127.0.0.1"}])
        tcprouter.create_service(name="google", sni="google.bots.grid.tf", addr="172.217.19.46:443")
        time.sleep(11)

        self.log("Delete google and facebook records from tcprouter.")
        self.delete_coredns_record("facebook")
        self.delete_coredns_record("google")
        time.sleep(11)

        self.log("Try to access facebook or google, should fail.")
        with self.assertRaises(requests.exceptions.ConnectionError):
            requests.get("https://facebook.bots.grid.tf", timeout=10)

        with self.assertRaises(requests.exceptions.ConnectionError):
            requests.get("https://google.bots.grid.tf", timeout=10)

        self.log("Try to access bing, should success.")
        with self.assertRaises(requests.exceptions.SSLError) as e:
            requests.get("https://bing.bots.grid.tf")

        error = e.exception.args[0].args[0]
        self.assertIn("bing.com", error)

    def test006_catch_all_record(self):
        """
        TFGW-011
        * Test case for catch_all record of tcprouter. *

        **Test Scenario:**

        #. Create (coredns and tcprouter) records for bing.
        #. Create coredns record for facebook.
        #. Create catch_all tcprouter record for facebook.
        #. Try to access facebook using any subdomain, should success.
        #. Try to access bing, should be accessed not facebook.
        """
        self.log("Create (coredns and tcprouter) records for bing.")
        coredns.create_a_record("bing", [{"ip": "127.0.0.1"}])
        tcprouter.create_service(name="bing", sni="bing.bots.grid.tf", addr="13.107.21.200:443")

        self.log("Create coredns record for facebook.")
        coredns.create_a_record("facebook", [{"ip": "127.0.0.1"}])

        self.log("Create catch_all tcprouter record for facebook.")
        tcprouter.create_service(name="CATCH_ALL", sni="CATCH_ALL", addr="102.132.97.35:443")
        time.sleep(11)

        self.log("Try to access facebook using any subdomain, should success.")
        with self.assertRaises(requests.exceptions.SSLError) as e:
            requests.get("https://facebook.bots.grid.tf")

        error = e.exception.args[0].args[0]
        self.assertIn("facebook.com", error)

        self.log("Try to access bing, should be accessed not facebook.")
        with self.assertRaises(requests.exceptions.SSLError) as e:
            requests.get("https://bing.bots.grid.tf")

        error = e.exception.args[0].args[0]
        self.assertIn("bing.com", error)
