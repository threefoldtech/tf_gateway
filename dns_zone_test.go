package tfgateway

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

	z.AddRecord("sub", a)
	z.AddRecord("sub", aaa)
	z.AddRecord("sub2", a)

	assert.Equal(t, 2, len(z.Records["sub"]))
	assert.Equal(t, 1, len(z.Records["sub2"]))
	assert.Equal(t, []DNSRecord{a, aaa}, z.Records["sub"])
	assert.Equal(t, []DNSRecord{a}, z.Records["sub2"])

	z.RemoveRecord("sub", aaa)
	assert.Equal(t, 1, len(z.Records["sub"]))
	assert.Equal(t, []DNSRecord{a}, z.Records["sub"])

	// remove no existing also works
	z.RemoveRecord("sub", aaa)
	assert.Equal(t, 1, len(z.Records["sub"]))
	assert.Equal(t, []DNSRecord{a}, z.Records["sub"])
}
