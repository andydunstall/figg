# Client Protocol

## Connection
Clients connect to Figg over TCP. Since Figg currently only supports a single
node, the address of that node is passed to the client.

### Ping/Pong

### Reconnect
If the client detects the connection has dropped it will automatically
reconnect, with exponential backoff to avoid overloading the server.

A user callback can be provided to be notified about connection state events,
such as disconnected and connected
