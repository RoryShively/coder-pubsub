package main

import (
	"errors"
	"sync"
)

// Global vars for datastore singleton.
var (
	dsOnce sync.Once
	ds     *Datastore
)

// MessageEvent is a data stucture used to pass new message events to
// subscribers.
type MessageEvent struct {
	topic   string
	message []byte
}

// Subscription contains channels used to recieve new messages and close the
// subscription when needed.
type Subscription struct {
	msgs   chan MessageEvent
	done   chan bool
	closed bool
}

// NewSubscription creates a new Subscription and starts a goroutine to close
// the subscription when the done channel closes.
func NewSubscription() *Subscription {
	sub := &Subscription{
		msgs:   make(chan MessageEvent),
		done:   make(chan bool),
		closed: false,
	}
	go func() {
		select {
		case _ = <-sub.done:
			sub.closed = true
		}
	}()
	return sub
}

// Utility type defining topic data.
type TopicMessages [][]byte

// Datastore stores all topics messages and stores all outstanding subscriptions.
// Topics can be safely accessed by and public methods but care should be taken
// using private methods as they aren't wrapped in the mutex.
type Datastore struct {
	topics      map[string]TopicMessages
	subscribers []*Subscription
	mu          sync.Mutex
}

// Create new Datastore with preloaded data.
func NewDataStore() *Datastore {
	topics := make(map[string]TopicMessages)
	topics["test_topic"] = append(topics["test_topic"], []byte("preloaded data 1"))
	topics["test_topic"] = append(topics["test_topic"], []byte("preloaded data 2"))
	topics["test_topic"] = append(topics["test_topic"], []byte("preloaded data 3"))
	topics["test_topic"] = append(topics["test_topic"], []byte("preloaded data 4"))
	topics["test_topic"] = append(topics["test_topic"], []byte("preloaded data 5"))

	return &Datastore{
		topics:      topics,
		subscribers: []*Subscription{},
	}
}

// Singleton access for ds (global Datastore).
func GetDatastore() *Datastore {
	dsOnce.Do(func() {
		ds = NewDataStore()
	})
	return ds
}

// RetrieveMessages returns all historic data on a topic.
func (v *Datastore) RetrieveMessages(topic string) (TopicMessages, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	// Check if topic exists
	if msgs, ok := v.topics[topic]; ok {
		return msgs, nil
	} else {
		return nil, errors.New("topic does not exist")
	}
}

// RetrieveWithSubscription returns all historic data along with a subscription
// to receive realtime data.
func (v *Datastore) RetrieveWithSubscription(topic string) (TopicMessages, *Subscription, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	// Check if topic exists
	if msgs, ok := v.topics[topic]; ok {
		// Create subscription channel to notify of new messages
		sub := NewSubscription()
		v.subscribers = append(v.subscribers, sub)
		return msgs, sub, nil
	} else {
		return nil, nil, errors.New("topic does not exist")
	}
}

// Add new message to datastore topic and notify subscribers.
func (v *Datastore) StoreMessage(topic string, msg []byte) {
	v.mu.Lock()
	defer v.mu.Unlock()

	// Check if topic exists
	if _, ok := v.topics[topic]; ok {
		// Append msg to TopicMessages
		v.topics[topic] = append(v.topics[topic], msg)
	} else {
		// Create new topic in Datastore
		v.topics[topic] = TopicMessages{msg}
	}
	evt := MessageEvent{
		topic:   topic,
		message: msg,
	}
	v.notifySubscribers(evt)
}

// Notify all subscribers of new data through their channels.
// Removes closed channels from sub list in the process.
func (v *Datastore) notifySubscribers(evt MessageEvent) {
	openSubs := []*Subscription{}
	for _, c := range v.subscribers {
		if c.closed {
			continue
		}
		openSubs = append(openSubs, c)
		c.msgs <- evt
	}
	v.subscribers = openSubs
}
