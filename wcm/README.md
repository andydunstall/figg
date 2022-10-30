# WCM (Wombat Cluster Manager)

A tool for launching Wombat clusters on localhost, used for local development
and system testing.

This is inspired by [CCM (Cassandra Cluster Manager)](https://github.com/riptano/ccm)
which is used by Cassandra.

WCM consists of two parts. A server written in Go that manages the cluster and
proxies cluster traffic, and a client issuing commands to the manager over gRPC.
This means its easy to write new clients, such as a CLI and Go library, and
also lets you control a remote cluster (such as in a Docker container) locally.
