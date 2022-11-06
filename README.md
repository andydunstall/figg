# Figg
Figg is a lightweight pub/sub messaging service.

*This is a work in progress projects I'm building for fun, not intended to be used for production workloads.*

## Components
* [`service/`](./service): Backend Figg service,
* [`sdk/`](./sdk): Go SDK client library,
* [`docs/`](./docs): Documentation on usage and architecture,
* [`tests/`](./tests): System tests,
* [`wcm/`](./wcm): Figg cluster manager.

## Testing
The service and SDK aims for high unit test coverage where possible which are
included in the [`service/`](./service) and [`sdk`](./sdk) packages alongside the code itself.

Though some end-to-end system tests are needed to:
* Check components are properly integrated,
* Inject chaos into a cluster to check for issues overlooked in the design.
These tests are in [`tests/`](./tests). [`WCM`](./wcm) is used to create Figg
clusters locally and inject chaos, which is used both for testing the service
and the SDK.
