package fcm

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ClusterInfo struct {
	ID    string     `json:"id,omitempty"`
	Nodes []NodeInfo `json:"nodes,omitempty"`
}

type NodeInfo struct {
	ID        string `json:"id,omitempty"`
	Addr      string `json:"addr,omitempty"`
	ProxyAddr string `json:"proxy_addr,omitempty"`
}

type ChaosConfig struct {
	Duration int
	Repeat   int
}

type FCM struct {
	client *Client
}

func NewFCM() *FCM {
	client := NewClient()
	return &FCM{
		client: client,
	}
}

func (f *FCM) AddCluster() (*ClusterInfo, error) {
	resp, err := f.client.Request(http.MethodPost, "/v1/clusters")
	if err != nil {
		return nil, err
	}
	defer resp.Close()

	cluster := &ClusterInfo{}
	if err := json.NewDecoder(resp).Decode(cluster); err != nil {
		return nil, err
	}
	return cluster, nil
}

func (f *FCM) RemoveCluster(id string) error {
	resp, err := f.client.Request(http.MethodDelete, "/v1/clusters/"+id)
	if err != nil {
		return err
	}
	defer resp.Close()

	return nil
}

func (f *FCM) AddChaosPartition(nodeID string, conf ChaosConfig) error {
	url := fmt.Sprintf("/v1/chaos/partition/%s?duration=%d&repeat=%d", nodeID, conf.Duration, conf.Repeat)
	resp, err := f.client.Request(http.MethodPost, url)
	if err != nil {
		return err
	}
	defer resp.Close()

	return nil
}
