package wcm

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Node struct {
	ID        string `json:"id,omitempty"`
	Addr      string `json:"addr,omitempty"`
	ClusterID string
	client    *Client
}

type Scenario struct {
	ID string `json:"id,omitempty"`
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

func (n *Node) AddLatency(d time.Duration) (string, error) {
	path := fmt.Sprintf("/v1/clusters/%s/nodes/%s/latency?latency=%d", n.ClusterID, n.ID, d.Milliseconds())

	resp, err := n.client.Request(http.MethodPost, path)
	if err != nil {
		return "", err
	}
	defer resp.Close()

	var scenario Scenario
	if err := json.NewDecoder(resp).Decode(&scenario); err != nil {
		return "", err
	}
	return scenario.ID, nil
}

func (n *Node) RemoveScenario(id string) error {
	path := fmt.Sprintf("/v1/clusters/%s/nodes/%s/chaos/%s", n.ClusterID, n.ID, id)

	resp, err := n.client.Request(http.MethodDelete, path)
	if err != nil {
		return err
	}
	defer resp.Close()

	return nil

}
