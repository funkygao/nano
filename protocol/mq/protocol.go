package mq

import (
	"github.com/funkygao/nano"
)

type Protocol interface {
	IOLoop(nano.Endpoint)
}
