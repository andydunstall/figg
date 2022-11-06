package fcm

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Cluster struct {
	ID     string `json:"id,omitempty"`
	client *Client
}

func NewCluster() (*Cluster, error) {
	client := NewClient()

	resp, err := client.Request(http.MethodPost, "/v1/clusters")
	if err != nil {
		return nil, err
	}
	defer resp.Close()

	var cluster Cluster
	if err := json.NewDecoder(resp).Decode(&cluster); err != nil {
		return nil, err
	}

	cluster.client = client
	return &cluster, nil
}

func (c *Cluster) AddNode() (*Node, error) {
	path := fmt.Sprintf("/v1/clusters/%s/nodes", c.ID)

	resp, err := c.client.Request(http.MethodPost, path)
	if err != nil {
		return nil, err
	}
	defer resp.Close()

	var node Node
	if err := json.NewDecoder(resp).Decode(&node); err != nil {
		return nil, err
	}
	node.ClusterID = c.ID
	node.client = c.client
	return &node, nil
}

func (c *Cluster) Shutdown() error {
	path := fmt.Sprintf("/v1/clusters/%s", c.ID)

	resp, err := c.client.Request(http.MethodDelete, path)
	if err != nil {
		return err
	}
	defer resp.Close()

	return nil
}
