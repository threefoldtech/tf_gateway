package redis

import (
	"fmt"
	"net/url"
	"time"

	redigo "github.com/gomodule/redigo/redis"
)

// NewPool creates a redis pool by connecting to address
// address format must contains the scheme to use
// tcp://host:port or unix:///path/to/socket
func NewPool(address string) (*redigo.Pool, error) {
	u, err := url.Parse(address)
	if err != nil {
		return nil, err
	}
	var host string
	switch u.Scheme {
	case "tcp":
		host = u.Host
	case "unix":
		host = u.Path
	default:
		return nil, fmt.Errorf("unknown scheme '%s' expecting tcp or unix", u.Scheme)
	}
	var opts []redigo.DialOption

	if u.User != nil {
		opts = append(
			opts,
			redigo.DialPassword(u.User.Username()),
		)
	}

	return &redigo.Pool{
		Dial: func() (redigo.Conn, error) {
			return redigo.Dial(u.Scheme, host, opts...)
		},
		TestOnBorrow: func(c redigo.Conn, t time.Time) error {
			if time.Since(t) > 10*time.Second {
				//only check connection if more than 10 second of inactivity
				_, err := c.Do("PING")
				return err
			}

			return nil
		},
		MaxActive:   3,
		MaxIdle:     3,
		IdleTimeout: 1 * time.Minute,
		Wait:        true,
	}, nil
}
