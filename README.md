# Wombat
Wombat is a lightweight pub/sub messaging service.

*This is a work in progress projects I'm building for fun, not intended to be used for production workloads.*

## Components
* [`service/`](./service): Backend Wombat service,
* [`sdk/`](./sdk): Go SDK client library,
* [`docs/`](./docs): Documentation on usage and architecture,
* [`wcm/`](./wcm): Wombat cluster manager.

## Testing
The service and SDK aims for high unit test coverage where possible.

Though some end-to-end system tests are needed to:
* Check components are properly integrated,
* Inject chaos into a cluster to check for issues overlooked in the design.
[`WCM`](./wcm) is used to create Wombat clusters locally and inject chaos, which
is used both for testing the service and the SDK.
