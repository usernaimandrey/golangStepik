package main

import (
	"context"
	"sync"
)

func NewBizServer(acl *MapAcl, sub *LogSubscribers, stat *Stat, statSub *StatSubscribers) *BizServerImp {
	return &BizServerImp{ACL: acl, LogSubscribers: sub, Stat: stat, StatSubscribers: statSub}
}

type BizServerImp struct {
	sync.RWMutex
	ACL             *MapAcl
	LogSubscribers  *LogSubscribers
	StatSubscribers *StatSubscribers
	Stat            *Stat
}

func (bsrv *BizServerImp) Check(ctx context.Context, n *Nothing) (*Nothing, error) {
	return &Nothing{}, nil
}

func (bsrv *BizServerImp) Add(ctx context.Context, n *Nothing) (*Nothing, error) {
	return &Nothing{}, nil
}

func (bsrv *BizServerImp) Test(ctx context.Context, n *Nothing) (*Nothing, error) {
	return &Nothing{}, nil
}

func (bsrv *BizServerImp) mustEmbedUnimplementedBizServer() {}
