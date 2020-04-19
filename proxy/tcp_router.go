package proxy

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gomodule/redigo/redis"
)

// service is the type use by the TCP router to configure proxies
// https://github.com/threefoldtech/tcprouter/blob/master/config.go#L36
type service struct {
	Addr         string `json:"addr"`
	ClientSecret string `json:"clientsecret` // will forward connection to it directly instead of hitting the Addr.
	TLSPort      int    `json:"tlsport"`
	HTTPPort     int    `json:"httpport"`

	UserID string `json:"user"`
}

type ProxyMgr struct {
	redis *redis.Pool
}

func New(pool *redis.Pool) (*ProxyMgr, error) {
	return &ProxyMgr{
		redis: pool,
	}, nil
}

func (r *ProxyMgr) key(domain string) string {
	return fmt.Sprintf("/tcprouter/services/%s", domain)
}

func (r *ProxyMgr) canUseDomain(user string, domain string) (bool, error) {
	con := r.redis.Get()
	defer con.Close()

	data, err := redis.Bytes(con.Do("GET", r.key(domain)))
	if err != nil {
		if errors.Is(err, redis.ErrNil) {
			return true, nil
		}
		return false, err
	}

	service := service{}
	if err := json.Unmarshal(data, &service); err != nil {
		return false, err
	}

	return service.UserID == user, nil
}

func (r *ProxyMgr) AddProxy(user string, domain, addr string, port, portTLS int) error {

	can, err := r.canUseDomain(user, domain)
	if err != nil {
		return err
	}
	if !can {
		return fmt.Errorf("cannot add proxy from %s: %w", domain, ErrAuth)
	}

	b, err := json.Marshal(service{
		Addr:     addr,
		HTTPPort: port,
		TLSPort:  portTLS,
		UserID:   user,
	})
	if err != nil {
		return err
	}

	con := r.redis.Get()
	defer con.Close()

	_, err = con.Do("SET", r.key(domain), b)
	return err
}
func (r *ProxyMgr) RemoveProxy(user string, domain string) error {
	can, err := r.canUseDomain(user, domain)
	if err != nil {
		return err
	}
	if !can {
		return fmt.Errorf("cannot remove proxy from %s: %w", domain, ErrAuth)
	}

	con := r.redis.Get()
	defer con.Close()
	_, err = con.Do("DELETE", r.key(domain))
	return err
}

func (r *ProxyMgr) AddReverseProxy(user string, domain, secret string) error {
	can, err := r.canUseDomain(user, domain)
	if err != nil {
		return err
	}
	if !can {
		return fmt.Errorf("cannot add reverse proxy from %s: %w", domain, ErrAuth)
	}

	b, err := json.Marshal(service{
		ClientSecret: secret,
		UserID:       user,
	})
	if err != nil {
		return err
	}

	con := r.redis.Get()
	defer con.Close()

	_, err = con.Do("SET", r.key(domain), b)
	return err
}

func (r *ProxyMgr) RemoveReverseProxy(user string, domain string) error {
	can, err := r.canUseDomain(user, domain)
	if err != nil {
		return err
	}
	if !can {
		return fmt.Errorf("cannot remove reverse proxy from %s: %w", domain, ErrAuth)
	}

	con := r.redis.Get()
	defer con.Close()
	_, err = con.Do("DELETE", r.key(domain))
	return err
}
