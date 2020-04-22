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
	aaaa := RecordAAAA{
		IP6: "fdb5:7faa:7be7:3ef3::1",
		TTL: 3600,
	}

	z.Add(a)
	z.Add(aaaa)

	assert.Equal(t, 1, len(z.Records[RecordTypeA]))
	assert.Equal(t, 1, len(z.Records[RecordTypeAAAA]))
	assert.Equal(t, []Record{a}, z.Records[RecordTypeA])
	assert.Equal(t, []Record{aaaa}, z.Records[RecordTypeAAAA])

	z.Remove(aaaa)
	assert.Equal(t, 0, len(z.Records[RecordTypeAAAA]))
	assert.Equal(t, []Record{a}, z.Records[RecordTypeA])
	assert.Equal(t, []Record{}, z.Records[RecordTypeAAAA])

	// remove no existing also works
	z.Remove(aaaa)
	assert.Equal(t, 0, len(z.Records[RecordTypeAAAA]))
}
