package dns

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"github.com/gomodule/redigo/redis"
)

type DNSMgr struct {
	redis  *redis.Pool
	prefix string
}

func New(pool *redis.Pool, prefix string) *DNSMgr {
	return &DNSMgr{
		redis:  pool,
		prefix: prefix,
	}
}

func (c *DNSMgr) zone(zone string) string {
	return fmt.Sprintf("%s%s", c.prefix, zone)
}

func (c *DNSMgr) AddSubdomain(user string, domain string, IPs []net.IP) error {

	name, zone := splitDomain(domain)

	con := c.redis.Get()
	defer con.Close()

	data, err := redis.Bytes(con.Do("GET", c.zone(zone)))
	if err != nil {
		return err
	}

	z := Zone{}
	if err := json.Unmarshal(data, &z); err != nil {
		return err
	}

	if z.Owner != "" && z.Owner != user {
		return fmt.Errorf("%w cannot add subdomain %s to zone %s", ErrAuth, name, zone)
	}

	for _, ip := range IPs {
		var r DNSRecord
		if ip.To16() == nil {
			r = RecordA{
				IP4: ip.String(),
				TTL: 3600,
			}
		} else {
			r = RecordAAA{
				IP6: ip.String(),
				TTL: 3600,
			}
		}
		z.AddRecord(name, r)
	}

	b, err := json.Marshal(z)
	if err != nil {
		return err
	}

	_, err = con.Do("HSET", c.zone(zone), name, b)
	return err
}

func (c *DNSMgr) RemoveSubdomain(user string, domain string, IPs []net.IP) error {
	name, zone := splitDomain(domain)

	con := c.redis.Get()
	defer con.Close()

	data, err := redis.Bytes(con.Do("GET", c.zone(zone)))
	if err != nil {
		return err
	}

	z := Zone{}
	if err := json.Unmarshal(data, &z); err != nil {
		return err
	}

	if z.Owner != "" && z.Owner != user {
		return fmt.Errorf("%w cannot remove subdomain %s from zone %s", ErrAuth, name, zone)
	}

	for _, ip := range IPs {
		var r DNSRecord
		if ip.To16() == nil {
			r = RecordA{
				IP4: ip.String(),
				TTL: 3600,
			}
		} else {
			r = RecordAAA{
				IP6: ip.String(),
				TTL: 3600,
			}
		}
		z.RemoveRecord(name, r)
	}

	b, err := json.Marshal(z)
	if err != nil {
		return err
	}

	_, err = con.Do("HSET", c.zone(zone), name, b)
	return err
}

func (c *DNSMgr) AddDomainDelagate(user, domain string) error {
	con := c.redis.Get()
	defer con.Close()

	data, err := redis.Bytes(con.Do("GET", c.zone(domain)))
	if err != nil {
		return err
	}

	z := Zone{}
	if err := json.Unmarshal(data, &z); err != nil {
		return err
	}

	if z.Owner != "" && z.Owner != user {
		return fmt.Errorf("%w cannot delegate domain %s", ErrAuth, domain)
	}

	z.Owner = user
	b, err := json.Marshal(z)
	if err != nil {
		return err
	}

	_, err = con.Do("HSET", c.zone(domain), b)
	return err
}

func (c *DNSMgr) RemoveDomainDelagate(user string, domain string) error {
	con := c.redis.Get()
	defer con.Close()

	data, err := redis.Bytes(con.Do("GET", c.zone(domain)))
	if err != nil {
		return err
	}

	z := Zone{}
	if err := json.Unmarshal(data, &z); err != nil {
		return err
	}

	if z.Owner != "" && z.Owner != user {
		return fmt.Errorf("%w cannot remove delegated domain %s", ErrAuth, domain)
	}

	_, err = con.Do("HDEL", c.zone(domain))
	return err
}

func splitDomain(d string) (name, domain string) {
	ss := strings.Split(d, ".")
	if len(ss) < 3 {
		return "", d + "."
	}
	return ss[0], strings.Join(ss[1:], ".") + "."
}
