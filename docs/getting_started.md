# Getting Started
Figg is a pub/sub messaging service.

Message streams are split into user defined topics, where users publish to
topics with `client.publish(topic, message)` and subscribe with
`client.subscribe(topic, subscriber)`.
