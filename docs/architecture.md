# Architecture
The architecture of Wombat is based on Cassandra and Dynamo.

## Consistent Hashing
The nodes in each region form a hash ring. When a topic is active in a region it
is assigned to a node in that region using consistent hashing. If the topic
is active in multiple regions, it will be assigned to one node per region.

## Connection Routing
Clients should connect to a random node in the cluster, typically using a load
load balancer. That node then acts as a coordinator for the connection.

Topics are distributed around the cluster with consistent hashing. So for each
topic the client is subscribed to, the coordinator connects to the corresponding
node and forwards messages to the client.

Alternatively clients have could connected to the nodes where the topics its
subscribing to are assigned directly, though has a few downsides:
* Each node would have to be routable by the client, which is often not the
case (such as when nodes are in a private subnet behind a load balancer),
* Clients would need to create a connect per topic,
* Fan-out could be limitted as all subscribers for a topic would be connecting
to the same node.

## Gossip
Gossip is used to propagate cluster information. Each node joins the gossip and
sends its state to all other nodes. This state includes the tokens for that
node (used for consistent hashing) and routing information.
