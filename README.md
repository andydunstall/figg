# Figg
Figg is a simple pub/sub messaging service. It only runs on a single node so
theres no fault tolerance or horizonal scaling.

*This is a work in progress project I'm just building for fun, not really intended to be used in production.*

## Components
* [`service/`](./service): Backend Figg service,
* [`sdk/`](./sdk): Go SDK client library,
* [`cli/`](./cli): Figg CLI,
* [`docs/`](./docs): Documentation on usage and architecture,
* [`tests/`](./tests): System tests,
* [`fcm/`](./fcm): Figg cluster manager.

## Testing
The service and SDK aims for high unit test coverage where possible which are
included in the [`service/`](./service) and [`sdk`](./sdk) packages alongside the code itself.

Though some end-to-end system tests are needed to:
* Check components are properly integrated,
* Inject chaos into a cluster to check for issues overlooked in the design.
These tests are in [`tests/`](./tests). [`FCM`](./fcm) is used to create Figg
clusters locally and inject chaos, which is used both for testing the service
and the SDK.
