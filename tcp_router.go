package tfgateway

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

type TCPRouter struct {
	redis *redis.Pool
}

func NewTCPRouter(pool *redis.Pool) (*TCPRouter, error) {
	return &TCPRouter{
		redis: pool,
	}, nil
}

func (r *TCPRouter) key(domain string) string {
	return fmt.Sprintf("/tcprouter/services/%s", domain)
}

func (r *TCPRouter) canUseDomain(user string, domain string) (bool, error) {
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

func (r *TCPRouter) AddProxy(user string, p Proxy) error {

	can, err := r.canUseDomain(user, p.Domain)
	if err != nil {
		return err
	}
	if !can {
		return fmt.Errorf("cannot add proxy from %s: %w", p.Domain, ErrAuth)
	}

	b, err := json.Marshal(service{
		Addr:     p.Addr,
		HTTPPort: int(p.Port),
		TLSPort:  int(p.PortTLS),
		UserID:   user,
	})
	if err != nil {
		return err
	}

	con := r.redis.Get()
	defer con.Close()

	_, err = con.Do("SET", r.key(p.Domain), b)
	return err
}
func (r *TCPRouter) RemoveProxy(user string, p Proxy) error {
	can, err := r.canUseDomain(user, p.Domain)
	if err != nil {
		return err
	}
	if !can {
		return fmt.Errorf("cannot remove proxy from %s: %w", p.Domain, ErrAuth)
	}

	con := r.redis.Get()
	defer con.Close()
	_, err = con.Do("DELETE", r.key(p.Domain))
	return err
}

func (r *TCPRouter) AddReverseProxy(user string, p ReverseProxy) error {
	can, err := r.canUseDomain(user, p.Domain)
	if err != nil {
		return err
	}
	if !can {
		return fmt.Errorf("cannot add reverse proxy from %s: %w", p.Domain, ErrAuth)
	}

	b, err := json.Marshal(service{
		ClientSecret: p.Secret,
		UserID:       user,
	})
	if err != nil {
		return err
	}

	con := r.redis.Get()
	defer con.Close()

	_, err = con.Do("SET", r.key(p.Domain), b)
	return err
}

func (r *TCPRouter) RemoveReverseProxy(user string, p ReverseProxy) error {
	can, err := r.canUseDomain(user, p.Domain)
	if err != nil {
		return err
	}
	if !can {
		return fmt.Errorf("cannot remove reverse proxy from %s: %w", p.Domain, ErrAuth)
	}

	con := r.redis.Get()
	defer con.Close()
	_, err = con.Do("DELETE", r.key(p.Domain))
	return err
}
