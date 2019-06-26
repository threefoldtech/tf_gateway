import unittest
from ..base_test import BaseTest
import redis, requests


class TFGWTests(BaseTest):
    def setUp(self):
        # should has coredns record.
        pass

    def tearDown(self):
        # should remove all coredns and tcprouter records.
        pass

    def test001_add_record(self):
        """
        TFGW-001
        *Test case for adding a record to (coredns and tcprouter) and try to access a website through them. *

        **Test Scenario:**

        #. Create a coredns record.
        #. Try to access the website(facebook), should fail.
        #. Create a tcprouter record.
        #. Try to access the website(facebook), should success.
        """
        self.log("Create a coredns record.")

        self.log("Try to access the website(facebook), should fail.")
        r = requests.get("https://facebook.wgtest.grid.tf")
        self.assertNotEquals(r.status_code, requests.codes.ok)

        self.log("Create a tcprouter record.")

        self.log("Try to access the website(facebook), should success.")
        r = requests.get("https://facebook.wgtest.grid.tf")
        self.assertEquals(r.status_code, requests.codes.ok)
        self.assertIn("facebook", r.text)

    def test002_add_more_records(self):
        """
        TFGW-002
        *Test case for adding records to tcprouter and try to access websites through them. *

        **Test Scenario:**

        #. Create a coredns record.
        #. Create a tcprouter record for facebook.
        #. Create a tcprouter record for bing.
        #. Create a tcprouter record for google.
        #. Try to access facebook, bing and google through tcprouter records, should success.
        """
        self.log("Create a coredns record.")
        self.log("Create a tcprouter record for facebook.")
        self.log("Create a tcprouter record for bing.")
        self.log("Create a tcprouter record for google.")

        self.log("Try to access facebook, bing and google through tcprouter records, should success.")
        r = requests.get("https://facebook.wgtest.grid.tf")
        self.assertEquals(r.status_code, requests.codes.ok)
        self.assertIn("facebook", r.text)

        r = requests.get("https://bing.wgtest.grid.tf")
        self.assertEquals(r.status_code, requests.codes.ok)
        self.assertIn("bing", r.text)

        r = requests.get("https://google.wgtest.grid.tf")
        self.assertEquals(r.status_code, requests.codes.ok)
        self.assertIn("google", r.text)

    def test003_modify_records(self):
        """
        TFGW-003
        *Test case for modifying tcprouter records and try to access websites through them*

        **Test Scenario:**

        #. Create a coredns record.
        #. Create a tcprouter record for facebook.
        #. Create a tcprouter record for bing.
        #. Create a tcprouter record for google.
        #. Modify google tcprouter record with facebook ip.
        #. Try to access facebook through google record, should success.
        """
        self.log("Create a coredns record.")
        self.log("Create a tcprouter record for facebook.")
        self.log("Create a tcprouter record for bing.")
        self.log("Create a tcprouter record for google.")
        self.log("Modify google tcprouter record with facebook ip.")

        self.log("Try to access facebook through google record, should success.")
        r = requests.get("https://facebook.wgtest.grid.tf")
        self.assertEquals(r.status_code, requests.codes.ok)
        self.assertIn("facebook", r.text)

        r = requests.get("https://bing.wgtest.grid.tf")
        self.assertEquals(r.status_code, requests.codes.ok)
        self.assertIn("bing", r.text)

        r = requests.get("https://google.wgtest.grid.tf")
        self.assertEquals(r.status_code, requests.codes.ok)
        self.assertIn("facebook", r.text)

    def test004_delete_records(self):
        """
        TFGW-004
        * Test case for deleting tcprouter records and make sure websites can't be accessed. *

        **Test Scenario:**

        #. Create a coredns record.
        #. Create a tcprouter record for facebook.
        #. Create a tcprouter record for bing.
        #. Create a tcprouter record for google.
        #. Delete google and facebook records from tcprouter.
        #. Try to access facebook or google, should fail.
        #. Try to access bing, should success.
        """
        self.log("Create a coredns record.")
        self.log("Create a tcprouter record for facebook.")
        self.log("Create a tcprouter record for bing.")
        self.log("Create a tcprouter record for google.")
        self.log("Delete google and facebook records from tcprouter.")

        self.log("Try to access facebook or google, should fail.")
        r = requests.get("https://facebook.wgtest.grid.tf")
        self.assertNotEquals(r.status_code, requests.codes.ok)

        r = requests.get("https://google.wgtest.grid.tf")
        self.assertNotEquals(r.status_code, requests.codes.ok)

        self.log("Try to access bing, should success.")
        r = requests.get("https://bing.wgtest.grid.tf")
        self.assertEquals(r.status_code, requests.codes.ok)
        self.assertIn("bing", r.text)

    def test005_catch_all_record(self):
        """
        TFGW-005
        * Test case for catch_all record of tcprouter. *

        **Test Scenario:**

        #. Create a coredns record.
        #. Create a tcprouter record for bing.
        #. Create catch_all tcprouter record for facebook.
        #. Try to access facebook using any subdomain, should success.
        #. Try to access bing, should be accessed not facebook.
        """
        self.log("Create a coredns record.")
        self.log("Create a tcprouter record for bing.")
        self.log("Create catch_all tcprouter record for facebook.")
        self.log("Try to access facebook using any subdomain, should success.")
        word = self.random_string()
        r = requests.get("https://{}.wgtest.grid.tf".format(word))
        self.assertEquals(r.status_code, requests.codes.ok)
        self.assertIn("facebook", r.text)

        self.log("Try to access bing, should be accessed not facebook.")
        r = requests.get("https://bing.wgtest.grid.tf")
        self.assertEquals(r.status_code, requests.codes.ok)
        self.assertIn("bing", r.text)
