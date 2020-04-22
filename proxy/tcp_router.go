package proxy

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gomodule/redigo/redis"
)

// service is the type use by the TCP router to configure proxies
// https://github.com/threefoldtech/tcprouter/blob/master/config.go#L36
type service struct {
	Addr         string `json:"addr"`
	ClientSecret string `json:"clientsecret"` // will forward connection to it directly instead of hitting the Addr.
	TLSPort      int    `json:"tlsport"`
	HTTPPort     int    `json:"httpport"`

	UserID string `json:"user"`
}

// Mgr is configure a TCP router server using redis
type Mgr struct {
	redis *redis.Pool
}

// New creates a new TCP router server manager
func New(pool *redis.Pool) *Mgr {
	return &Mgr{redis: pool}
}

func (r *Mgr) key(domain string) string {
	return fmt.Sprintf("/tcprouter/service/%s", domain)
}

func (r *Mgr) canUseDomain(user string, domain string) (bool, error) {
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
	if err := valkyrieDecode(data, &service); err != nil {
		return false, err
	}

	return service.UserID == user, nil
}

// AddProxy adds a TCP proxy from domain to addr
// port is for plain text protocol, usually HTTP
// portTLS is for TCL protocol, usually HTTPS
func (r *Mgr) AddProxy(user string, domain, addr string, port, portTLS int) error {

	can, err := r.canUseDomain(user, domain)
	if err != nil {
		return err
	}
	if !can {
		return fmt.Errorf("cannot add proxy from %s: %w", domain, ErrAuth)
	}

	key := r.key(domain)
	b, err := valkyrieEncode(key, service{
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

	_, err = con.Do("SET", key, b)
	return err
}

// RemoveProxy removes a proxy added with AddProxy
func (r *Mgr) RemoveProxy(user string, domain string) error {
	can, err := r.canUseDomain(user, domain)
	if err != nil {
		return err
	}
	if !can {
		return fmt.Errorf("cannot remove proxy from %s: %w", domain, ErrAuth)
	}

	con := r.redis.Get()
	defer con.Close()
	_, err = con.Do("DEL", r.key(domain))
	return err
}

// AddReverseProxy add a reverse tunnel TCP proxy from domain to the TCP connection identityied by secret
func (r *Mgr) AddReverseProxy(user string, domain, secret string) error {
	can, err := r.canUseDomain(user, domain)
	if err != nil {
		return err
	}
	if !can {
		return fmt.Errorf("cannot add reverse proxy from %s: %w", domain, ErrAuth)
	}

	key := r.key(domain)
	b, err := valkyrieEncode(key, service{
		ClientSecret: secret,
		UserID:       user,
	})
	if err != nil {
		return err
	}

	con := r.redis.Get()
	defer con.Close()

	_, err = con.Do("SET", key, b)
	return err
}

// RemoveReverseProxy removes a reverse tunnel proxy added with AddReverseProxy
func (r *Mgr) RemoveReverseProxy(user string, domain string) error {
	can, err := r.canUseDomain(user, domain)
	if err != nil {
		return err
	}
	if !can {
		return fmt.Errorf("cannot remove reverse proxy from %s: %w", domain, ErrAuth)
	}

	con := r.redis.Get()
	defer con.Close()
	_, err = con.Do("DEL", r.key(domain))
	return err
}

type valkyrieObj struct {
	Key   string
	Value string
}

func valkyrieEncode(key string, v interface{}) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	return json.Marshal(valkyrieObj{
		Key:   key,
		Value: base64.StdEncoding.EncodeToString(b),
	})
}

func valkyrieDecode(b []byte, v interface{}) error {
	obj := valkyrieObj{}
	if err := json.Unmarshal(b, &obj); err != nil {
		return err
	}

	value, err := base64.StdEncoding.DecodeString(obj.Value)
	if err != nil {
		return err
	}

	return json.Unmarshal(value, v)
}
