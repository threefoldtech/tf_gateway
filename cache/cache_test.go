package cache

import (
	"fmt"
	"testing"
	"time"

	"github.com/threefoldtech/tfgateway/redis"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/threefoldtech/zos/pkg/provision"
)

func TestLocalStore(t *testing.T) {
	s, err := miniredis.Run()
	require.NoError(t, err)
	defer s.Close()

	pool, err := redis.NewPool(fmt.Sprintf("tcp://%s", s.Addr()))
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
