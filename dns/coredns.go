package dns

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"github.com/gomodule/redigo/redis"
)

// Mgr is responsible to configure CoreDNS trough its redis pluging
type Mgr struct {
	redis  *redis.Pool
	prefix string
}

// New creates a DNS manager
func New(pool *redis.Pool, prefix string) *Mgr {
	return &Mgr{
		redis:  pool,
		prefix: prefix,
	}
}

func (c *Mgr) zone(zone string) string {
	return fmt.Sprintf("%s%s", c.prefix, zone)
}

// AddSubdomain configures a domain A or AAA records depending on the version of
// the IP address in IPs
func (c *Mgr) AddSubdomain(user string, domain string, IPs []net.IP) error {

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
		var r Record
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
		z.Add(name, r)
	}

	b, err := json.Marshal(z)
	if err != nil {
		return err
	}

	_, err = con.Do("HSET", c.zone(zone), name, b)
	return err
}

// RemoveSubdomain remove a domain added with AddSubdomain
func (c *Mgr) RemoveSubdomain(user string, domain string, IPs []net.IP) error {
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
		var r Record
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
		z.Remove(name, r)
	}

	b, err := json.Marshal(z)
	if err != nil {
		return err
	}

	_, err = con.Do("HSET", c.zone(zone), name, b)
	return err
}

// AddDomainDelagate configures coreDNS to manage domain
func (c *Mgr) AddDomainDelagate(user, domain string) error {
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

// RemoveDomainDelagate remove a delagated domain added with AddDomainDelagate
func (c *Mgr) RemoveDomainDelagate(user string, domain string) error {
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
