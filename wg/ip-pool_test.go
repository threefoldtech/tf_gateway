package wg

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/threefoldtech/zos/pkg/identity"
)

func TestIPPool(t *testing.T) {
	kp, err := identity.FromSeed([]byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"))
	require.NoError(t, err)

	pool := NewIPPool(kp)

	subnet := pool.Subnet()
	assert.Equal(t, "fd5e:ca9b:d3eb:7c0::/64", subnet.String())

	ip, err := pool.Get([]byte("hello"))
	require.NoError(t, err)

	assert.Equal(t, subnet.IP[:7], ip[:7], "generated client IP must be in the same prefix")

	ip2, err := pool.Get([]byte("hello"))
	require.NoError(t, err)
	assert.Equal(t, ip, ip2, "same ID should generated the same IP")

	ip3, err := pool.Get([]byte("world"))
	require.NoError(t, err)
	assert.NotEqual(t, ip, ip3, "different ID should generated the different IP")
}
