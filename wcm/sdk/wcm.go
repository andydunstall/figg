package wcm

import (
	"context"

	pb "github.com/andydunstall/wombat/wcm/sdk/pkg/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type WCM struct {
	conn      *grpc.ClientConn
	rpcClient pb.WCMClient
}

func Connect() (*WCM, error) {
	conn, err := grpc.Dial(
		"127.0.0.1:7229", grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	return &WCM{
		conn:      conn,
		rpcClient: pb.NewWCMClient(conn),
	}, nil
}

func (w *WCM) CreateCluster(ctx context.Context) (*Cluster, error) {
	resp, err := w.rpcClient.CreateCluster(ctx, &pb.EmptyMessage{})
	if err != nil {
		return nil, err
	}
	return NewCluster(resp.Id, w.rpcClient), nil
}

func (w *WCM) Close() error {
	return w.conn.Close()
}
