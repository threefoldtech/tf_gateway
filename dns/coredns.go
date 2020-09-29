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

	if zone[len(zone)-1] != '.' {
		zone += "."
	}

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

	if zone[len(zone)-1] != '.' {
		zone += "."
	}

	b, err := json.Marshal(zr.Records)
	if err != nil {
		return err
	}

	if _, err := con.Do("HSET", zone, name, b); err != nil {
		return err
	}

	return nil
}

func (c *Mgr) setSubdomainOwner(domain, user string) error {
	log.Debug().Msgf("set managed domain owner %s %s", domain, user)
	con := c.redis.Get()
	defer con.Close()

	if _, err := con.Do("HSET", "managed_domains", domain, user); err != nil {
		return err
	}

	return nil
}

func (c *Mgr) getSubdomainOwner(domain string) (user string, err error) {
	log.Debug().Msgf("get managed domain owner %s %s", domain, user)
	con := c.redis.Get()
	defer con.Close()

	user, err = redis.String(con.Do("HGET", "managed_domains", domain))
	if err != nil {
		if errors.Is(err, redis.ErrNil) {
			return "", nil
		}
		return "", err
	}

	return user, nil
}

func (c *Mgr) deleteSubdomainOwner(domain string) error {
	log.Debug().Msgf("delete managed domain owner %s", domain)
	con := c.redis.Get()
	defer con.Close()

	_, err := con.Do("HDEL", "managed_domains", domain)
	return err
}

// AddSubdomain configures a domain A or AAA records depending on the version of
// the IP address in IPs
func (c *Mgr) AddSubdomain(user string, domain string, IPs []net.IP) error {

	log.Info().Msgf("add subdomain %s %+v", domain, IPs)

	if err := validateDomain(domain); err != nil {
		return err
	}

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

	if owner.Owner == c.identity { // this is a manged domain
		owner, err := c.getSubdomainOwner(domain)
		if err != nil {
			return err
		}
		if owner != "" && owner != user {
			return fmt.Errorf("%w cannot add subdomain %s to zone %s", ErrAuth, name, zone)
		}
	} else if owner.Owner != user { //this is a deletegatedDomain
		return fmt.Errorf("%w cannot add subdomain %s to zone %s", ErrAuth, name, zone)
	}

	zr, err := c.getZoneRecords(zone, name)
	if err != nil {
		return err
	}

	for _, ip := range IPs {
		r := recordFromIP(ip)
		zr.Add(r)
	}

	if err := c.setZoneRecords(zone, name, zr); err != nil {
		return err
	}

	if owner.Owner == c.identity {
		return c.setSubdomainOwner(domain, user)
	}
	return nil
}

// RemoveSubdomain remove a domain added with AddSubdomain
func (c *Mgr) RemoveSubdomain(user string, domain string, IPs []net.IP) error {
	if err := validateDomain(domain); err != nil {
		return err
	}

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

	if owner.Owner == c.identity { // this is a manged domain
		owner, err := c.getSubdomainOwner(domain)
		if err != nil {
			return err
		}
		if owner != "" && owner != user {
			return fmt.Errorf("%w cannot remove subdomain %s from zone %s", ErrAuth, name, zone)
		}
	} else if owner.Owner != user { //this is a deletegatedDomain
		return fmt.Errorf("%w cannot remove subdomain %s from zone %s", ErrAuth, name, zone)
	}

	zr, err := c.getZoneRecords(zone, name)
	if err != nil {
		return err
	}

	if zr.Records.IsEmpty() {
		return nil
	}

	for _, ip := range IPs {
		r := recordFromIP(ip)
		zr.Remove(r)
	}

	if err := c.setZoneRecords(zone, name, zr); err != nil {
		return err
	}

	if owner.Owner == c.identity && zr.Records.IsEmpty() {
		// if the subomain has been cleared out, we remove the owner so anyone can claim it again
		return c.deleteSubdomainOwner(domain)
	}
	return nil
}

// AddDomainDelagate configures coreDNS to manage domain
func (c *Mgr) AddDomainDelagate(user, domain string) error {
	if err := validateDomain(domain); err != nil {
		return err
	}

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
	if err := validateDomain(domain); err != nil {
		return err
	}

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
		return "", d
	}
	return ss[0], strings.Join(ss[1:], ".")
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

func validateDomain(domain string) error {
	if len(domain) == 0 {
		return fmt.Errorf("incorrect format for domain %s", domain)
	}

	if strings.Count(domain, ".") < 1 {
		return fmt.Errorf("incorrect format for domain %s", domain)
	}

	if domain[len(domain)-1] == '.' {
		return fmt.Errorf("incorrect format for domain %s", domain)
	}

	if strings.Contains(domain, "..") {
		return fmt.Errorf("incorrect format for domain %s", domain)

	}

	return nil
}
