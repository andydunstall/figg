package wcm

import (
	"fmt"
	"net/http"
)

type Node struct {
	ID        string `json:"id,omitempty"`
	Addr      string `json:"addr,omitempty"`
	ClusterID string
	client    *Client
}

func (n *Node) Enable() error {
	path := fmt.Sprintf("/v1/clusters/%s/nodes/%s/enable", n.ClusterID, n.ID)

	resp, err := n.client.Request(http.MethodPost, path)
	if err != nil {
		return err
	}
	defer resp.Close()

	return nil
}

func (n *Node) Disable() error {
	path := fmt.Sprintf("/v1/clusters/%s/nodes/%s/disable", n.ClusterID, n.ID)

	resp, err := n.client.Request(http.MethodPost, path)
	if err != nil {
		return err
	}
	defer resp.Close()

	return nil
}
