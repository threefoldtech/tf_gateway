package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/threefoldtech/zos/pkg/provision/mw"
)

// GatewayInfo struct
type GatewayInfo struct {
	ManagedDomains []string `json:"managed_domains"`
	Nameservers    []string `json:"nameservers"`
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
