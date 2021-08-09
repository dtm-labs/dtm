package dtmpb

import (
	"github.com/yedf/dtm/dtmcli"
	grpc "google.golang.org/grpc"
)

var clients = map[string]DtmClient{}

// GetDtmClient 1
func GetDtmClient(grpcServer string) (cli DtmClient, rerr error) {
	if clients[grpcServer] == nil {
		conn, err := grpc.Dial(grpcServer, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithUnaryInterceptor(dtmcli.GrpcClientLog))
		if err == nil {
			clients[grpcServer] = NewDtmClient(conn)
			dtmcli.Logf("dtm client inited for %s", grpcServer)
		}
	}
	cli = clients[grpcServer]
	return
}

// MustGetDtmClient 1
func MustGetDtmClient(grpcServer string) DtmClient {
	cli, err := GetDtmClient(grpcServer)
	dtmcli.E2P(err)
	return cli
}
