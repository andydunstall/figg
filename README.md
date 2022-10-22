# Wombat
Wombat is a lightweight pub/sub messaging service.

*This is a work in progress projects I'm building for fun, not intended to be used for production workloads.*

## Usage
The protocol is a very simple publish and subscribe. The topic to connect to
is identified by the endpoint URL, then all incoming data is considered a
publish and sent to all other subscribers on the topic.

**REST**: `POST http://{config.addr}/v1/{topic}`

REST only supports publishing, where each request is published to the given
topic

**Websocket**: `GET ws://{config.addr}/v1/{topic}/ws`

Websocket supports both publish and subscribe, where incoming messages are
published to all other subscribers
