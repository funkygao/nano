package api

import (
	"errors"
)

var (
	ErrBadDomain   = errors.New("bad domain")
	ErrBadProtocol = errors.New("bad protocol")
)
