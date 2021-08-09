package dtmpb

import (
	"strings"

	"github.com/yedf/dtm/dtmcli"
	grpc "google.golang.org/grpc"
)

var clients = map[string]*grpc.ClientConn{}

// GetGrpcConn 1
func GetGrpcConn(grpcServer string) (conn *grpc.ClientConn, rerr error) {
	if clients[grpcServer] == nil {
		conn, err := grpc.Dial(grpcServer, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithUnaryInterceptor(dtmcli.GrpcClientLog))
		if err == nil {
			clients[grpcServer] = conn
			dtmcli.Logf("dtm client inited for %s", grpcServer)
		}
	}
	conn = clients[grpcServer]
	return
}

// MustGetGrpcConn 1
func MustGetGrpcConn(grpcServer string) *grpc.ClientConn {
	conn, err := GetGrpcConn(grpcServer)
	dtmcli.E2P(err)
	return conn
}

// MustGetDtmClient 1
func MustGetDtmClient(grpcServer string) DtmClient {
	return NewDtmClient(MustGetGrpcConn(grpcServer))
}

// GetServerAndMethod 将grpc的url分解为server和method
func GetServerAndMethod(grpcURL string) (string, string) {
	fs := strings.Split(grpcURL, "/")
	server := fs[0]
	method := "/" + strings.Join(fs[1:], "/")
	return server, method
}
