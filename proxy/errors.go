package proxy

import "errors"

var (
	// ErrAuth is return when a user is not allow to do certain action on a domain
	// most usually it is because the domain is own by someone else
	ErrAuth = errors.New("unauthorized error")
)
