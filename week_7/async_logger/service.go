package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"log"
	"net"
	"strings"
)

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные

func StartMyMicroservice(ctx context.Context, listenAddr string, ACLData string) error {
	acl, err := ACLParser(ACLData)

	if err != nil {
		return err
	}

	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Println(err)
		return err
	}

	server := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor),
		grpc.StreamInterceptor(streemAuthInterceptor),
	)
	go stopServer(ctx, server)
	sub := NewLogSubscribers()
	stat := NewStat()
	statSub := NewStatSubscribers()
	RegisterBizServer(server, NewBizServer(&MapAcl{m: acl["biz"]}, sub, stat, statSub))
	RegisterAdminServer(server, NewAdminServer(&MapAcl{m: acl["adm"]}, sub, stat, statSub))

	fmt.Println("start server on the port :" + strings.Split(listenAddr, ":")[1])

	go func() {
		server.Serve(lis)
	}()

	return nil
}

func stopServer(ctx context.Context, srv *grpc.Server) {
	<-ctx.Done()
	srv.Stop()
}

func authInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {

	resp, consumer, err := ACLValidations(ctx, info.Server, info.FullMethod)

	if err != nil {
		return resp, err
	}

	host := GetHostAndPort(ctx)

	go AddStat(consumer, info.Server, info.FullMethod)
	go Logger(consumer, host, info.Server, info.FullMethod)

	reply, err := handler(ctx, req)

	return reply, err
}

func streemAuthInterceptor(
	srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {

	ctx := ss.Context()

	_, consumer, err := ACLValidations(ctx, srv, info.FullMethod)
	if err != nil {
		return err
	}

	host := GetHostAndPort(ctx)

	go AddStat(consumer, srv, info.FullMethod)
	go Logger(consumer, host, srv, info.FullMethod)

	err = handler(srv, ss)

	return err
}

func Logger(consumerName string, host string, srv interface{}, fullMethodName string) error {
	switch v := srv.(type) {
	case *BizServerImp:
		v.LogSubscribers.RLock()
		v.LogSubscribers.BroadcastEvent(&Event{Consumer: consumerName, Method: fullMethodName, Host: host})
		v.LogSubscribers.RUnlock()
	case *AdminServerImp:
		v.LogSubscribers.RLock()
		v.LogSubscribers.BroadcastEvent(&Event{Consumer: consumerName, Method: fullMethodName, Host: host})
		v.LogSubscribers.RUnlock()
	default:
		return status.Errorf(codes.Unknown, "unknow service")
	}
	return nil
}

func AddStat(consumerName string, srv interface{}, fullMethodName string) error {
	switch v := srv.(type) {
	case *BizServerImp:
		v.Lock()
		v.StatSubscribers.BroadcastEvent(&Stat{ByMethod: map[string]uint64{fullMethodName: 1}, ByConsumer: map[string]uint64{consumerName: 1}})
		v.Unlock()
	case *AdminServerImp:
		v.Lock()
		v.StatSubscribers.BroadcastEvent(&Stat{ByMethod: map[string]uint64{fullMethodName: 1}, ByConsumer: map[string]uint64{consumerName: 1}})
		v.Unlock()
	default:
		return status.Errorf(codes.Unknown, "unknow service")
	}
	return nil
}

func GetHostAndPort(ctx context.Context) string {
	var host string
	var port int
	p, ok := peer.FromContext(ctx)

	if ok && p.Addr != nil {
		if tcpAddr, ok := p.Addr.(*net.TCPAddr); ok {
			host = tcpAddr.IP.String()
			port = tcpAddr.Port
		}
	}

	return fmt.Sprintf("%s:%d", host, port)
}
