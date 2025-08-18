package main

import (
	"sync"
)

func NewLogSubscribers() *LogSubscribers {
	return &LogSubscribers{Subscribers: &[]chan Event{}}
}

type LogSubscribers struct {
	sync.RWMutex
	Subscribers *[]chan Event
}

func (ls *LogSubscribers) AddSubscribe(ch chan Event) {
	ls.Lock()
	sub := *ls.Subscribers
	sub = append(sub, ch)
	ls.Subscribers = &sub
	ls.Unlock()
}

func (ls *LogSubscribers) BroadcastEvent(event *Event) {
	for _, ch := range *ls.Subscribers {
		ch <- *event
	}
}

func (ls *LogSubscribers) RemoveSubscriber(sub chan Event) {
	ls.Lock()
	subscribers := *ls.Subscribers
	for i, ch := range subscribers {
		if ch == sub {
			close(ch)
			subscribers = append(subscribers[:i], subscribers[i+1:]...)
			ls.Subscribers = &subscribers
		}
	}
	ls.Unlock()
}
