# TODO

## SDK
* Fix flaky system tests
* Ensure thread safe (currently connection isn't)
* Add long running tests
  * Use cli stream with chaos and run for hours to check no leaks or race
      conditions (use `-race`)

## CLI
* Improve benchmarks
  * See NATs and Redis bench commands

## FCM
* Refactor and ensure safe
  * Currently need sleeps to avoid races

## Service
* Add commit log retention
* Add long running tests
  * Use cli stream with chaos and run for hours to check no leaks or race
      conditions (use `-race`)

## Protocol
* Rather than split resumed messages into `DATA` messages just stream the
commit log directly
