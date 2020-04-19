package tfgateway

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gomodule/redigo/redis"
)

type CoreDNS struct {
	redis  *redis.Pool
	prefix string
}

func NewCoreDNS(pool *redis.Pool, prefix string) (*CoreDNS, error) {
	return &CoreDNS{
		redis:  pool,
		prefix: prefix,
	}, nil
}

func (c *CoreDNS) zone(zone string) string {
	return fmt.Sprintf("%s%s", c.prefix, zone)
}

func (c *CoreDNS) AddSubdomain(user string, s Subdomain) error {

	name, zone := splitDomain(s.Domain)

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

	for _, ip := range s.IPs {
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

func (c *CoreDNS) RemoveSubdomain(user string, s Subdomain) error {
	name, zone := splitDomain(s.Domain)

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

	for _, ip := range s.IPs {
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

func (c *CoreDNS) AddDomainDelagate(user string, d Delegate) error {
	con := c.redis.Get()
	defer con.Close()

	data, err := redis.Bytes(con.Do("GET", c.zone(d.Domain)))
	if err != nil {
		return err
	}

	z := Zone{}
	if err := json.Unmarshal(data, &z); err != nil {
		return err
	}

	if z.Owner != "" && z.Owner != user {
		return fmt.Errorf("%w cannot delegate domain %s", ErrAuth, d.Domain)
	}

	z.Owner = user
	b, err := json.Marshal(z)
	if err != nil {
		return err
	}

	_, err = con.Do("HSET", c.zone(d.Domain), b)
	return err
}

func (c *CoreDNS) RemoveDomainDelagate(user string, d Delegate) error {
	con := c.redis.Get()
	defer con.Close()

	data, err := redis.Bytes(con.Do("GET", c.zone(d.Domain)))
	if err != nil {
		return err
	}

	z := Zone{}
	if err := json.Unmarshal(data, &z); err != nil {
		return err
	}

	if z.Owner != "" && z.Owner != user {
		return fmt.Errorf("%w cannot remove delegated domain %s", ErrAuth, d.Domain)
	}

	_, err = con.Do("HDEL", c.zone(d.Domain))
	return err
}

func splitDomain(d string) (name, domain string) {
	ss := strings.Split(d, ".")
	if len(ss) < 3 {
		return "", d + "."
	}
	return ss[0], strings.Join(ss[1:], ".") + "."
}
