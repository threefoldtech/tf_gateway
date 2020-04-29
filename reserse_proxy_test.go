package tfgateway

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReverseProxyValidate(t *testing.T) {
	for _, tt := range []struct {
		ReverseProxy ReverseProxy
		WantError    bool
		User         string
	}{
		{
			ReverseProxy: ReverseProxy{
				Domain: "hello.world",
				Secret: "user1:asdasdasd",
			},
			User:      "user1",
			WantError: false,
		},
		{
			ReverseProxy: ReverseProxy{
				Domain: "",
				Secret: "",
			},
			User:      "user1",
			WantError: true,
		},
		{
			ReverseProxy: ReverseProxy{
				Domain: "hello.world",
				Secret: "",
			},
			WantError: true,
		},
		{
			ReverseProxy: ReverseProxy{
				Domain: "hello.world",
				Secret: "notprefix:asdasdasd",
			},
			User:      "user1",
			WantError: true,
		},
	} {
		t.Run(fmt.Sprintf("%+v", tt.ReverseProxy), func(t *testing.T) {
			err := tt.ReverseProxy.validate(tt.User)
			if tt.WantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
