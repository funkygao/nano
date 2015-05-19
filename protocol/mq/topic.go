package mq

import (
	"github.com/funkygao/nano"
)

type Topic struct {
	name       string
	partitions int
	mq         *mq
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

func NewTopic(topicName string, mq *mq) *Topic {
	t := &Topic{
		name:       topicName,
		mq:         mq,
		channelMap: make(map[string]*Channel),
	}

	go t.messagePump()

	return t
}
