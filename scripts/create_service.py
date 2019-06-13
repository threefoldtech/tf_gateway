from sys import argv
import base64
import json
import redis

r = redis.Redis()

def create_service(name, sni, addr):
    service = {}
    service['Key'] = '/tcprouter/service/{}'.format(name)
    record = {"addr":addr, "sni":sni, "name":name}
    json_dumped_record_bytes = json.dumps(record).encode()
    b64_record = base64.b64encode(json_dumped_record_bytes).decode()
    service['Value'] = b64_record
    r.set(service['Key'], json.dumps(service))
    

if len(argv) > 3:
    create_service(argv[1], argv[2], argv[3])
else:
    print("usage: create_service.py SERVICE_NAME SNI ADDR")

# create_service('facebook', "www.facebook.com", "102.132.97.35:443")
# create_service('google', 'www.google.com', '172.217.19.46:443')
# create_service('bing', 'www.bing.com', '13.107.21.200:443')
            