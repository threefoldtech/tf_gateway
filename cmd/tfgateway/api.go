package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/threefoldtech/zos/pkg/provision/mw"
)

var (
	startTime = time.Now()
)

// Uptime struct calculates the seconds since the object was created
type Uptime struct{}

// Duration returns the uptime duration
func (u *Uptime) Duration() time.Duration {
	return time.Since(startTime)
}

//MarshalJSON implementation
func (u Uptime) MarshalJSON() ([]byte, error) {
	seconds := u.Duration() / time.Second
	return json.Marshal(int64(seconds))
}

// GatewayInfo struct
type GatewayInfo struct {
	ManagedDomains []string `json:"managed_domains"`
	Nameservers    []string `json:"nameservers"`
	Uptime         Uptime   `json:"uptime"`
}

// InfoAPI struct
type InfoAPI struct {
	info GatewayInfo
}

// NewInfoAPI sets up the info api
func NewInfoAPI(router *mux.Router, info GatewayInfo) error {

	api := InfoAPI{info}
	return api.setup(router)
}

func (i *InfoAPI) setup(router *mux.Router) error {
	router.Path("/info").Methods(http.MethodGet).HandlerFunc(mw.AsHandlerFunc(i.getInfo)).Name("gateway-info")
	return nil
}

func (i *InfoAPI) getInfo(r *http.Request) (interface{}, mw.Response) {
	return i.info, nil
}
