package mq

import (
	"github.com/funkygao/nano"
)

type Topic struct {
	name       string
	ctx        *context
	channelMap map[string]*Channel
	backend    BackendQueue
}

func (this *Topic) PutMessage(m *nano.Message) {
}

func (this *Topic) GetChannel(channelName string) *Channel {
	return nil
}

func (this *Topic) messagePump() {

}

func NewTopic(topicName string, ctx *context) *Topic {
	t := &Topic{
		name:       topicName,
		ctx:        ctx,
		channelMap: make(map[string]*Channel),
	}

	go t.messagePump()

	return t
}
