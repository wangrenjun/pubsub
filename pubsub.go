package pubsub

import (
	"sync"
	"time"
)

type TopicMessage struct {
	Topic   string
	Message interface{}
}

type PubSub struct {
	subsmap map[string]map[chan<- TopicMessage]struct{}
	lock    sync.RWMutex
}

func New() *PubSub {
	return &PubSub{
		subsmap: make(map[string]map[chan<- TopicMessage]struct{}),
	}
}

func (self *PubSub) subChannels(topic string) []chan<- TopicMessage {
	var channels []chan<- TopicMessage
	self.lock.RLock()
	defer self.lock.RUnlock()
	if subs, found := self.subsmap[topic]; found {
		channels = make([]chan<- TopicMessage, 0, len(subs))
		for ch := range subs {
			channels = append(channels, ch)
		}
	}
	return channels
}

func (self *PubSub) Pub(msg interface{}, topics ...string) (num int) {
	for _, topic := range topics {
		channels := self.subChannels(topic)
		for _, ch := range channels {
			ch <- TopicMessage{Topic: topic, Message: msg}
			num++
		}
	}
	return
}

func (self *PubSub) TryPub(msg interface{}, topics ...string) (num int) {
	self.lock.RLock()
	defer self.lock.RUnlock()
	for _, topic := range topics {
		if subs, found := self.subsmap[topic]; found {
			for ch := range subs {
				select {
				case ch <- TopicMessage{Topic: topic, Message: msg}:
					num++
				default:
				}
			}
		}
	}
	return
}

func timedSend(ch chan<- TopicMessage, topicMsg TopicMessage, wait time.Duration, wg *sync.WaitGroup, num *int) {
	defer wg.Done()
	select {
	case ch <- topicMsg:
		*num++
	case <-time.After(wait):
	}
}

func (self *PubSub) TimedPub(msg interface{}, wait time.Duration, topics ...string) (num int) {
	var wg sync.WaitGroup
	self.lock.RLock()
	defer self.lock.RUnlock()
	for _, topic := range topics {
		if subs, found := self.subsmap[topic]; found {
			for ch := range subs {
				wg.Add(1)
				go timedSend(ch, TopicMessage{Topic: topic, Message: msg}, wait, &wg, &num)
			}
		}
	}
	wg.Wait()
	return
}

func (self *PubSub) Sub(ch chan<- TopicMessage, topics ...string) {
	self.lock.Lock()
	defer self.lock.Unlock()
	for _, topic := range topics {
		if _, found := self.subsmap[topic]; !found {
			self.subsmap[topic] = make(map[chan<- TopicMessage]struct{})
		}
		if _, found := self.subsmap[topic][ch]; !found {
			self.subsmap[topic][ch] = struct{}{}
		}
	}
}

func (self *PubSub) UnSub(ch chan<- TopicMessage, topics ...string) {
	self.lock.Lock()
	defer self.lock.Unlock()
	for _, topic := range topics {
		if _, found := self.subsmap[topic]; found {
			if _, found := self.subsmap[topic][ch]; found {
				delete(self.subsmap[topic], ch)
			}
			if len(self.subsmap[topic]) == 0 {
				delete(self.subsmap, topic)
			}
		}
	}
}

func (self *PubSub) RemoveTopics(topics ...string) {
	self.lock.Lock()
	defer self.lock.Unlock()
	for _, topic := range topics {
		delete(self.subsmap, topic)
	}
}

func (self *PubSub) Topics() []string {
	self.lock.RLock()
	defer self.lock.RUnlock()
	topics := make([]string, 0, len(self.subsmap))
	for topic := range self.subsmap {
		topics = append(topics, topic)
	}
	return topics
}

func (self *PubSub) NumTopic() int {
	self.lock.RLock()
	defer self.lock.RUnlock()
	return len(self.subsmap)
}

func (self *PubSub) NumSub(topic string) int {
	self.lock.RLock()
	defer self.lock.RUnlock()
	return len(self.subsmap[topic])
}
