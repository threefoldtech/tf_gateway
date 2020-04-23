package dns

import (
	"encoding/json"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_splitDomain(t *testing.T) {
	tests := []struct {
		domain string
		name   string
		zone   string
	}{
		{
			domain: "domain.com",
			name:   "",
			zone:   "domain.com.",
		},
		{
			domain: "a.domain.com",
			name:   "a",
			zone:   "domain.com.",
		},
		{
			domain: "a.b.c.domain.com",
			name:   "a",
			zone:   "b.c.domain.com.",
		},
		{
			domain: "bleh.grid.deboeck.xyz",
			name:   "bleh",
			zone:   "grid.deboeck.xyz.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.domain, func(t *testing.T) {
			name, zone := splitDomain(tt.domain)
			assert.Equal(t, tt.name, name)
			assert.Equal(t, tt.zone, zone)
		})
	}
}

func TestRecordFromIP(t *testing.T) {
	tests := []struct {
		ip     net.IP
		record Record
	}{
		{ip: net.ParseIP("185.15.201.80"),
			record: RecordA{
				IP4: "185.15.201.80",
				TTL: 3600,
			},
		},
		{
			ip: net.ParseIP("2a02:2788:864:1314:9eb6:d0ff:fe97:764b"),
			record: RecordAAAA{
				IP6: "2a02:2788:864:1314:9eb6:d0ff:fe97:764b",
				TTL: 3600,
			},
		},
	}

	for _, tt := range tests {
		r := recordFromIP(tt.ip)
		assert.Equal(t, tt.record, r)
	}
}

func TestLoadRecords(t *testing.T) {
	z := Zone{}
	z.Add(RecordA{
		IP4: "142.93.229.35",
		TTL: 3600,
	})

	b, err := json.Marshal(z.Records)
	require.NoError(t, err)

	z2 := Zone{Records: records{}}
	err = json.Unmarshal(b, &z2.Records)
	require.NoError(t, err)
	require.Equal(t, 1, len(z.Records))
	assert.Equal(t, z.Records, z2.Records)
}
