package main

import (
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"strings"
	"sync"
)

func NewAcl(methodFullName string, serviceName string, methodName string) *ACL {
	return &ACL{MethodFullName: methodFullName, Service: serviceName, Method: methodName}
}

type ACL struct {
	MethodFullName string
	Service        string
	Method         string
}

type MapAcl struct {
	sync.RWMutex
	m map[string][]*ACL
}

func ACLParser(acl string) (map[string]map[string][]*ACL, error) {
	var data interface{}
	aclEmpty := map[string]map[string][]*ACL{}

	err := json.Unmarshal([]byte(acl), &data)

	if err != nil {
		return aclEmpty, err
	}

	bizACL := map[string][]*ACL{}
	admACL := map[string][]*ACL{}
	var unpackError error

outer:
	for user, acl := range data.(map[string]interface{}) {
		v, ok := acl.([]interface{})

		if !ok {
			unpackError = fmt.Errorf("not convert %s to interface", acl)
			break
		}

		for _, aclI := range v {
			fullMethodName, ok := aclI.(string)

			if !ok {
				unpackError = fmt.Errorf("not convert %s to string", aclI)
				break outer
			}

			serviceName, methodName, err := ACLMethodParser(fullMethodName)

			if err != nil {
				unpackError = err
				break outer
			}

			acl := NewAcl(fullMethodName, serviceName, methodName)

			switch serviceName {
			case "Biz":
				if _, ok := bizACL[user]; !ok {
					aclColl := []*ACL{}
					aclColl = append(aclColl, acl)
					bizACL[user] = aclColl
				} else {
					bizACL[user] = append(bizACL[user], acl)
				}
			case "Admin":
				if _, ok := admACL[user]; !ok {
					aclColl := []*ACL{}
					aclColl = append(aclColl, acl)
					admACL[user] = aclColl
				} else {
					admACL[user] = append(admACL[user], acl)
				}
			default:
				unpackError = fmt.Errorf("invalid format acl data %s, expected format [service, method] ", data)
				break outer
			}
		}
	}

	if unpackError != nil {
		return aclEmpty, unpackError
	}

	aclEmpty["biz"] = bizACL
	aclEmpty["adm"] = admACL
	return aclEmpty, nil
}

func ACLValidations(ctx context.Context, srv interface{}, fullMethodName string) (interface{}, string, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	consumers := md.Get("consumer")
	var resp interface{}

	if len(consumers) == 0 {
		return resp, "", status.Errorf(codes.Unauthenticated, "disallowed method")
	}
	consumerName := consumers[0]

	var ok bool
	var consumerACLs []*ACL

	switch v := srv.(type) {
	case *BizServerImp:
		consumerACLs, ok = v.ACL.m[consumerName]
	case *AdminServerImp:
		consumerACLs, ok = v.ACL.m[consumerName]
	default:
		return resp, "", status.Errorf(codes.Unknown, "unknow service")
	}

	if !ok {
		return resp, "", status.Errorf(codes.Unauthenticated, "disallowed method")
	}

	authConsumer := false
	for _, acl := range consumerACLs {
		serviceName, methodName, err := ACLMethodParser(fullMethodName)

		if err != nil {
			break
		}

		if acl.Service == serviceName && (acl.Method == methodName || acl.Method == "*") {
			authConsumer = true
		}
	}

	if !authConsumer {
		return resp, "", status.Errorf(codes.Unauthenticated, "disallowed method")
	}

	return resp, consumerName, nil
}

func ACLMethodParser(fullMethodName string) (string, string, error) {
	validLenAcl := 2
	data := strings.Split(strings.TrimPrefix(fullMethodName, "/"), "/")

	if len(data) < validLenAcl {
		return "", "", fmt.Errorf("invalid format acl data %s, expected format [service, method] ", data)

	}

	serviceName := strings.Split(data[0], ".")[1]
	methodName := data[1]
	return serviceName, methodName, nil
}
