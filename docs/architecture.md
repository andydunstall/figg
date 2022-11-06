# Architecture

Starting building with only a single node and will build up from there. The plan
is to eventually support multiple nodes and regions.

## Message Flow
The [Client Protocol](client_protocol.md) docs describe the protocol from the
SDK to a figg node. This section describes how that message is processed once
its been received by a Figg node.

### Commit Log
Each topic has a commit log on disk containing the messages on that topic. This
is similar to how Kafka works as described in the
[paper](https://notes.stephenholiday.com/Kafka.pdf).

When a message arrives it is appended to the commit log for that topic. At
this point it is considered accepted so an acknowledgement is sent to the
client.

Note the purpose of the commit log is retention not fault tolerance. The plan
is to have a configurable number of replicas we send the message too before it
is accepted. So it doesn't wait for the message to be flushed before
acknowledging.

### Subscribers
Once the message has been accepted, it can be sent to the subscribers.

The subscribers on the node (not the SDK subscriber) have a goroutine and an
offset of the next message they need. They just keep iterating though the topic
sending the message to the connection, then once up to date block waiting for
the next message. So the topic just signals these subscribers to wake up and
check for new messages.

The pull based model means the topic itself doesn't have to worry about blocking
or sending messages.

## Next
* Add a cluster where the topics preference list is selected with consistent
hashing
  * Use Dynamo/Cassandra style of replication, with a configurable replication
strategy (both number of nodes to replicate too, and number of replicas that
must respond syncrously before acknowledging)
