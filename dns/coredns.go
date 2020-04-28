package dns

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/gomodule/redigo/redis"
)

// Mgr is responsible to configure CoreDNS trough its redis pluging
type Mgr struct {
	redis    *redis.Pool
	identity string
}

// New creates a DNS manager
func New(pool *redis.Pool, identity string) *Mgr {
	return &Mgr{
		redis:    pool,
		identity: identity,
	}
}

func (c *Mgr) getZoneOwner(zone string) (owner ZoneOwner, err error) {
	zone = strings.TrimSuffix(zone, ".")

	con := c.redis.Get()
	defer con.Close()

	data, err := redis.Bytes(con.Do("HGET", "zone", zone))
	if err != nil {
		if errors.Is(err, redis.ErrNil) {
			return owner, nil
		}
		return owner, fmt.Errorf("failed to read the DNS zone %s: %w", zone, err)
	}

	if err := json.Unmarshal(data, &owner); err != nil {
		return owner, err
	}
	return owner, nil
}

func (c *Mgr) setZoneOwner(zone string, owner ZoneOwner) (err error) {
	con := c.redis.Get()
	defer con.Close()

	b, err := json.Marshal(owner)
	if err != nil {
		return err
	}

	_, err = con.Do("HSET", "zone", zone, b)
	return err
}

func (c *Mgr) getZoneRecords(zone, name string) (Zone, error) {
	con := c.redis.Get()
	defer con.Close()

	zr := Zone{Records: records{}}
	data, err := redis.Bytes(con.Do("HGET", zone, name))
	if err != nil {
		if errors.Is(err, redis.ErrNil) {
			return zr, nil
		}
		return zr, fmt.Errorf("failed to read the DNS zone %s: %w", zone, err)
	}

	if err := json.Unmarshal(data, &zr.Records); err != nil {
		return zr, err
	}
	log.Debug().Msgf("get zone records %+v", zr)
	return zr, nil
}

func (c *Mgr) setZoneRecords(zone, name string, zr Zone) (err error) {
	log.Debug().Msgf("zet zone records %+v", zr)
	con := c.redis.Get()
	defer con.Close()

	b, err := json.Marshal(zr.Records)
	if err != nil {
		return err
	}

	if _, err := con.Do("HSET", zone, name, b); err != nil {
		return err
	}

	return nil
}

// AddSubdomain configures a domain A or AAA records depending on the version of
// the IP address in IPs
func (c *Mgr) AddSubdomain(user string, domain string, IPs []net.IP) error {

	log.Info().Msgf("add subdomain %s %+v", domain, IPs)

	name, zone := splitDomain(domain)

	con := c.redis.Get()
	defer con.Close()

	owner, err := c.getZoneOwner(zone)
	if err != nil {
		return fmt.Errorf("failed to read the DNS zone %s: %w", zone, err)
	}

	if owner.Owner == "" {
		return fmt.Errorf("%s is not managed by the gateway. Delegate the domain first", zone)
	}

	if owner.Owner != c.identity && owner.Owner != user {
		return fmt.Errorf("%w cannot add subdomain %s to zone %s", ErrAuth, name, zone)
	}

	zr, err := c.getZoneRecords(zone, name)
	if err != nil {
		return err
	}

	// if this is a managed domain and there is already some records, then refuse to modify it
	if owner.Owner == c.identity && len(zr.Records) > 0 {
		return fmt.Errorf("the sub-domain %s is already used by someone else: %w", domain, ErrAuth)
	}

	for _, ip := range IPs {
		r := recordFromIP(ip)
		zr.Add(r)
	}

	return c.setZoneRecords(zone, name, zr)
}

// RemoveSubdomain remove a domain added with AddSubdomain
func (c *Mgr) RemoveSubdomain(user string, domain string, IPs []net.IP) error {
	name, zone := splitDomain(domain)

	con := c.redis.Get()
	defer con.Close()

	owner, err := c.getZoneOwner(zone)
	if err != nil {
		return fmt.Errorf("failed to read the DNS zone %s: %w", zone, err)
	}

	if owner.Owner == "" {
		// domain not managed by this gateway at all, so all subdomain are already gone too.
		// this can happen when a delegated domain expires before a subdomain
		return nil
	}

	if owner.Owner != c.identity && owner.Owner != user {
		return fmt.Errorf("%w cannot remove subdomain %s from zone %s", ErrAuth, name, zone)
	}

	zr, err := c.getZoneRecords(zone, name)
	if err != nil {
		return err
	}

	if len(zr.Records) == 0 {
		return nil
	}

	for _, ip := range IPs {
		r := recordFromIP(ip)
		zr.Remove(r)
	}

	return c.setZoneRecords(zone, name, zr)
}

// AddDomainDelagate configures coreDNS to manage domain
func (c *Mgr) AddDomainDelagate(user, domain string) error {
	owner, err := c.getZoneOwner(domain)
	if err != nil {
		return err
	}

	if owner.Owner != "" && owner.Owner != user {
		return fmt.Errorf("%w cannot delegate domain %s", ErrAuth, domain)
	}

	owner.Owner = user
	return c.setZoneOwner(domain, owner)
}

// RemoveDomainDelagate remove a delagated domain added with AddDomainDelagate
func (c *Mgr) RemoveDomainDelagate(user string, domain string) error {
	owner, err := c.getZoneOwner(domain)
	if err != nil {
		return err
	}

	if owner.Owner != "" && owner.Owner != user {
		return fmt.Errorf("%w cannot remove delegated domain %s", ErrAuth, domain)
	}

	con := c.redis.Get()
	defer con.Close()

	if _, err = con.Do("HDEL", "zone", domain); err != nil {
		return err
	}
	// remove all eventual subdomain configuration for this delegated domain
	_, err = con.Do("DEL", domain)
	return err
}

func splitDomain(d string) (name, domain string) {
	ss := strings.Split(d, ".")
	if len(ss) < 3 {
		return "", d + "."
	}
	return ss[0], strings.Join(ss[1:], ".") + "."
}

func recordFromIP(ip net.IP) (r Record) {
	if ip.To4() != nil {
		r = RecordA{
			IP4: ip.String(),
			TTL: 3600,
		}
	} else {
		r = RecordAAAA{
			IP6: ip.String(),
			TTL: 3600,
		}
	}
	return r
}
