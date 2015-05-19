package mq

import (
	"github.com/funkygao/nano"
)

type Topic struct {
	name       string
	partitions int
	broker     *broker
	channelMap map[string]*Channel
	backend    BackendQueue
}

func (this *Topic) PutMessage(m *nano.Message) {
}

func (this *Topic) GetChannel(channelName string) *Channel {
	return nil
}

func (this *Topic) Close() {

}

func (this *Topic) messagePump() {

}

func newTopic(topicName string, broker *broker) *Topic {
	t := &Topic{
		name:       topicName,
		broker:     broker,
		channelMap: make(map[string]*Channel),
	}

	go t.messagePump()

	return t
}
