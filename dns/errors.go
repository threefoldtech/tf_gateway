package dns

import "errors"

var (
	// ErrAuth is return when a user is not allow to do certain action on a domain
	// most usually it is because the domain is own by someone else
	ErrAuth = errors.New("unauthorized error")
	// ErrSubdomainUsed returned if the subdomain is already reserved
	ErrSubdomainUsed = errors.New("subdomain already reserved")
)
