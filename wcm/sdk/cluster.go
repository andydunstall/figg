package wcm

import (
	"context"

	pb "github.com/andydunstall/wombat/wcm/sdk/pkg/rpc"
)

type Cluster struct {
	ID        string
	rpcClient pb.WCMClient
}

func NewCluster(id string, rpcClient pb.WCMClient) *Cluster {
	return &Cluster{
		ID:        id,
		rpcClient: rpcClient,
	}
}

func (c *Cluster) AddNode(ctx context.Context) (*Node, error) {
	resp, err := c.rpcClient.ClusterAddNode(ctx, &pb.ClusterInfo{
		Id: c.ID,
	})
	if err != nil {
		return nil, err
	}
	return NewNode(resp.Id, resp.Addr), nil
}

func (c *Cluster) Close(ctx context.Context) error {
	_, err := c.rpcClient.ClusterClose(ctx, &pb.ClusterInfo{
		Id: c.ID,
	})
	if err != nil {
		return err
	}
	return nil
}
