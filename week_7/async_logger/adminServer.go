package main

import (
	"context"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc/metadata"
)

func NewAdminServer(acl *MapAcl, sub *LogSubscribers, stat *Stat, statSub *StatSubscribers) *AdminServerImp {
	return &AdminServerImp{ACL: acl, LogSubscribers: sub, Stat: stat, StatSubscribers: statSub}
}

type AdminServerImp struct {
	sync.RWMutex
	ACL             *MapAcl
	LogSubscribers  *LogSubscribers
	StatSubscribers *StatSubscribers
	Stat            *Stat
}

func (asrv *AdminServerImp) Logging(n *Nothing, als Admin_LoggingServer) error {

	ctx := als.Context()
	md, _ := metadata.FromIncomingContext(ctx)
	consumers := md.Get("consumer")
	ch := asrv.NewLogSubscriber()

	go asrv.SendLogEvents(als.Context(), consumers[0], als, ch)
	<-als.Context().Done()

	return nil
}

func (asrv *AdminServerImp) SendLogEvents(ctx context.Context, consumer string, als Admin_LoggingServer, ch *chan Event) {
	for {
		select {
		case <-ctx.Done():
			asrv.LogSubscribers.RemoveSubscriber(*ch)
			return
		case event := <-*ch:
			if event.Consumer != consumer {
				err := als.Send(&event)
				if err != nil {
					return
				}
			}
		}
	}
}

func (asrv *AdminServerImp) SendStatEvent(ctx context.Context, ass Admin_StatisticsServer, ch *chan Stat, interval uint64) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	statEventStorage := &Stat{
		ByMethod:   map[string]uint64{},
		ByConsumer: map[string]uint64{},
	}
	for {

		select {
		case statEvent := <-*ch:
			for k, v := range statEvent.ByConsumer {
				if _, ok := statEventStorage.ByConsumer[k]; !ok {
					statEventStorage.ByConsumer[k] = v
				} else {
					statEventStorage.ByConsumer[k] += v
				}
			}
			for k, v := range statEvent.ByMethod {
				if _, ok := statEventStorage.ByMethod[k]; !ok {
					statEventStorage.ByMethod[k] = v
				} else {
					statEventStorage.ByMethod[k] += v
				}
			}
		case <-ctx.Done():
			log.Println("STOP")
			asrv.StatSubscribers.RemoveSubscriber(*ch)
			ticker.Stop()
			return
		case <-ticker.C:

			err := ass.Send(statEventStorage)
			if err != nil {
				log.Println(err, "ERROR")
				return
			}
			statEventStorage.SoftReset()
		}

	}
}

func (asrv *AdminServerImp) NewLogSubscriber() *chan Event {
	ch := make(chan Event)
	asrv.LogSubscribers.AddSubscribe(ch)
	return &ch
}

func (asrv *AdminServerImp) NewStatSubscribers() *chan Stat {
	ch := make(chan Stat)
	asrv.StatSubscribers.AddSubscribe(ch)
	return &ch
}

func (asrv *AdminServerImp) Statistics(interval *StatInterval, ass Admin_StatisticsServer) error {

	ctx := ass.Context()

	ch := asrv.NewStatSubscribers()
	i := interval.GetIntervalSeconds()

	go asrv.SendStatEvent(ctx, ass, ch, i)
	<-ass.Context().Done()

	return nil
}

func (asrv *AdminServerImp) mustEmbedUnimplementedAdminServer() {}
