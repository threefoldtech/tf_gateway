package tfgateway

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
	}
	for _, tt := range tests {
		t.Run(tt.domain, func(t *testing.T) {
			name, zone := splitDomain(tt.domain)
			assert.Equal(t, tt.name, name)
			assert.Equal(t, tt.zone, zone)
		})
	}
}
