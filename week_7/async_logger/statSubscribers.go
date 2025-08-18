package main

import (
	"sync"
)

func NewStatSubscribers() *StatSubscribers {
	return &StatSubscribers{Subscribers: &[]chan Stat{}}
}

type StatSubscribers struct {
	sync.RWMutex
	Subscribers *[]chan Stat
}

func (ss *StatSubscribers) AddSubscribe(ch chan Stat) {
	ss.Lock()
	sub := *ss.Subscribers
	sub = append(sub, ch)
	ss.Subscribers = &sub
	ss.Unlock()
}

func (ss *StatSubscribers) BroadcastEvent(stat *Stat) {
	ss.RLock()
	for _, ch := range *ss.Subscribers {
		ch <- *stat
	}
	ss.RUnlock()
}

func (ss *StatSubscribers) RemoveSubscriber(sub chan Stat) {
	ss.Lock()
	subscribers := *ss.Subscribers
	for i, ch := range subscribers {
		if ch == sub {
			close(ch)
			subscribers = append(subscribers[:i], subscribers[i+1:]...)
			ss.Subscribers = &subscribers
		}
	}
	ss.Unlock()
}
