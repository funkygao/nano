package mq

import (
	"sync"

	"github.com/funkygao/nano"
)

type Channel struct {
	topicName string
	name      string
	backend   BackendQueue
	eps       map[nano.EndpointId]nano.Endpoint

	sync.RWMutex
}

func (this *Channel) PutMessage(m *nano.Message) {

}

func (this *Channel) AddEndpoint(ep nano.Endpoint) {
	this.Lock()
	this.eps[ep.Id()] = ep
	this.Unlock()
}

func (this *Channel) RemoveEndpoint(ep nano.Endpoint) {
	delete(this.eps, ep.Id())
}

func (this *Channel) FinishMessage() {

}
