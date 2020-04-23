package cache

import (
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/threefoldtech/zos/pkg/provision"
)

func TestLocalStore(t *testing.T) {
	pool, err := newRedisPool("tcp://localhost:6379")
	require.NoError(t, err)

	cache := NewRedis(pool)

	type args struct {
		r *provision.Reservation
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "main",
			args: args{
				r: &provision.Reservation{
					ID:       "r-1",
					Created:  time.Now().UTC().Add(-time.Minute).Round(time.Second),
					Duration: time.Second * 10,
					Tag:      provision.Tag{"source": "tfgateway_cache"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err = cache.Add(tt.args.r)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			actual, err := cache.Get(tt.args.r.ID)
			require.NoError(t, err)
			assert.EqualValues(t, tt.args.r, actual)

			_, err = cache.Get("foo")
			require.Error(t, err)

			expired, err := cache.GetExpired()
			require.NoError(t, err)
			assert.Equal(t, len(expired), 1)
			assert.Equal(t, tt.args.r, expired[0])

			err = cache.Remove(actual.ID)
			assert.NoError(t, err)

			_, err = cache.Get(tt.args.r.ID)
			require.Error(t, err)
		})
	}
}

func newRedisPool(address string) (*redis.Pool, error) {
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
	var opts []redis.DialOption

	if u.User != nil {
		opts = append(
			opts,
			redis.DialPassword(u.User.Username()),
		)
	}

	return &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.Dial(u.Scheme, host, opts...)
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
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
