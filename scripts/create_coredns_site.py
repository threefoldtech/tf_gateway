from sys import argv
import base64
import json
import redis

r = redis.Redis()




# redis-cli> hgetall example.net.
#  1) "_ssh._tcp.host1"
#  2) "{\"srv\":[{\"ttl\":300, \"target\":\"tcp.example.com.\",\"port\":123,\"priority\":10,\"weight\":100}]}"
#  3) "*"
#  4) "{\"txt\":[{\"ttl\":300, \"text\":\"this is a wildcard\"}],\"mx\":[{\"ttl\":300, \"host\":\"host1.example.net.\",\"preference\": 10}]}"
#  5) "host1"
#  6) "{\"a\":[{\"ttl\":300, \"ip\":\"5.5.5.5\"}]}"
#  7) "sub.*"
#  8) "{\"txt\":[{\"ttl\":300, \"text\":\"this is not a wildcard\"}]}"
#  9) "_ssh._tcp.host2"
# 10) "{\"srv\":[{\"ttl\":300, \"target\":\"tcp.example.com.\",\"port\":123,\"priority\":10,\"weight\":100}]}"
# 11) "subdel"
# 12) "{\"ns\":[{\"ttl\":300, \"host\":\"ns1.subdel.example.net.\"},{\"ttl\":300, \"host\":\"ns2.subdel.example.net.\"}]}"
# 13) "@"
# 14) "{\"soa\":{\"ttl\":300, \"minttl\":100, \"mbox\":\"hostmaster.example.net.\",\"ns\":\"ns1.example.net.\",\"refresh\":44,\"retry\":55,\"expire\":66},\"ns\":[{\"ttl\":300, \"host\":\"ns1.example.net.\"},{\"ttl\":300, \"host\":\"ns2.example.net.\"}]}"
# redis-cli>

def create_bot_record(domain="", record_type="a", records=None):
    """
    for every entry you need to comply with record format


    """
    data = {}
    records = records or []
    if r.hexists("bots.grid.tf.", domain):
        data = json.loads(r.hget("bots.grid.tf.", domain))
    if record_type in data:
        records.extend(data[record_type])
    data[record_type] = records
    r.hset("bots.grid.tf.", domain, json.dumps(data))
    
    

def create_a_record(domain, records):
    for rec in records:
        assert "ip" in rec
    
    return create_bot_record(domain, "a", records)

def create_aaaa_record(domain, records):
    for rec in records:
        assert "ip" in rec
    
    return create_bot_record(domain, "aaaa", records)


def create_txt_record(domain, records):
    for rec in records:
        assert "txt" in rec
    
    return create_bot_record(domain, "txt", records)


def create_ns_record(domain, records):
    for rec in records:
        assert "host" in rec
    
    return create_bot_record(domain, "ns", records)


    
