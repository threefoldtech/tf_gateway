package dns

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestZone(t *testing.T) {
	z := &Zone{}

	a := RecordA{
		IP4: "10.0.0.1",
		TTL: 3600,
	}
	aaa := RecordAAA{
		IP6: "fdb5:7faa:7be7:3ef3::1",
		TTL: 3600,
	}

	z.Add("sub", a)
	z.Add("sub", aaa)
	z.Add("sub2", a)

	assert.Equal(t, 2, len(z.Records["sub"]))
	assert.Equal(t, 1, len(z.Records["sub2"]))
	assert.Equal(t, []Record{a, aaa}, z.Records["sub"])
	assert.Equal(t, []Record{a}, z.Records["sub2"])

	z.Remove("sub", aaa)
	assert.Equal(t, 1, len(z.Records["sub"]))
	assert.Equal(t, []Record{a}, z.Records["sub"])

	// remove no existing also works
	z.Remove("sub", aaa)
	assert.Equal(t, 1, len(z.Records["sub"]))
	assert.Equal(t, []Record{a}, z.Records["sub"])
}
