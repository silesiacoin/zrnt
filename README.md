# ZRNT [![Go Report Card](https://goreportcard.com/badge/github.com/protolambda/zrnt?no-cache)](https://goreportcard.com/report/github.com/protolambda/zrnt) [![CircleCI Build Status](https://circleci.com/gh/protolambda/zrnt.svg?style=shield)](https://circleci.com/gh/protolambda/zrnt) [![codecov](https://codecov.io/gh/protolambda/zrnt/branch/master/graph/badge.svg?no-cache)](https://codecov.io/gh/protolambda/zrnt) [![MIT Licensed](https://img.shields.io/badge/license-MIT-blue.svg)](./LICENSE)

A minimal Go implementation of the ETH 2.0 spec, by @protolambda.

The goal of this project is to have a Go version of the Python based spec,
 to enable use cases that are out of reach for unoptimized Python.

Think of:
- Realtime use in test-network monitoring, fast enough to verify full network activity.
- A much higher fuzzing speed, to explore for new bugs.
- Not affected nearly as much as pyspec when scaling to mainnet.

## Structure of `eth2` package

### `beacon`

The beacon package covers the phase0 state transition, following the Beacon-chain spec, but optimized for performance.

#### Globals

Globals are split up as following:
- `<something>Type`: a ZTYP SSZ type description, used to create typed views with.
- `<something>View`: a ZTYP SSZ view. This wraps a binary-tree backing,
 to provide typed a mutable interface to the persistent cached binary-tree datastructure.
- `<something>SSZ`: a ZSSZ SSZ definition. Used to enhance native Go datastructures with SSZ type info,
 to directly hash/encode the raw data structures, without representing them as binary tree.
- `<something>`: Other types are mostly just raw ZSSZ-compatible Go datastructures, and can be used with the corresponding `<something>SSZ`
- `(state *BeaconStateView) Process<something>`: The processing functions are all prefixed, and attached to the beacon-state view.
- `(state *BeaconStateView) StateTransition`: the main transition function
- `EpochsContext` and `(state *BeaconStateView) NewEpochsContext`: to efficiently transition, data that is used accross the epoch slot is cached in a special container.
 This includes beacon proposers, committee shuffling, pubkey cache, etc. Functions to efficiently copy and update are included.
- `GenesisFromEth1` and `KickStartState` two functions to create a `*BeaconStateView` and accompanying `*EpochsContext` with.

#### Working around lack of Generics

A common pattern is to have `As<Something` functions that work like `(view View, err error) -> (typedView *SomethingView, err error)`.
Since Go has no generics, ZTYP does not deserialize/get any typed views. If you load a new view, or get an attribute,
it will have to be wrapped with an `As<Something>` function to be fully typed. The 2nd `err` input is for convenience:
chaining getters and `As` functions together is much easier this way. If there is any error, it is proxied.

#### Importing

If you work with many of the types and functions, it may be handy to import with a `.` for brevity.

```go
import . "github.com/protolambda/zrnt/eth2/beacon"
```

### `util`

Hashing, merkleization, and other utils can be found in `eth2/util`.

BLS is a package that wraps Herumi BLS. However, it is put behind a build-flag. Use `bls_on` and `bls_off` to use it or not.

SSZ is provided by two components:
- ZRNT-SSZ (ZSSZ), optimized for speed on small raw Go datasturctures (no hash-tree-root caching, but serialization is 10x as fast as Gob):
 [`github.com/protolambda/zssz`](https://github.com/protolambda/zssz). Primarily used for inputs such as blocks and attestations.
  Not repeatedly hashed, and important to quickly decode, read/mutate, and encode.
- ZTYP, optimized for caching, representing data as binary trees. Primarily used for the `BeaconStateView`.

### Building

Re-generate dynamic parts by running `go generate ./...` (make sure the needed configuration files are in `./presets/configs/`).

By default, the `defaults.go` preset is used, which is selected if no other presets are,
i.e. `!preset_mainnet,!preset_minimal` if `mainnet` and `minimal` are the others.

For custom configuration, add a build-constraint (also known as "build tag") when building ZRNT or using it as a dependency:

```
go build -tags preset_mainnet

or

go build -tags preset_minimal

```

BLS can be turned off by adding the `bls_off` build tag (security warning: for testing use only!).

### Testing

To run all tests and generate test and coverage reports: `make test`

The specs are tested using test-vectors shared between Eth 2.0 clients,
 found here: [`ethereum/eth2.0-spec-tests`](https://github.com/ethereum/eth2.0-spec-tests).
Instructions on the usage of these test-vectors with ZRNT can be found in the [testing readme](./tests/spec/README.md).

#### Coverage reports

After running `make test`, run `make open-coverage` to open a Go-test coverage report in your browser.

## Contributing

Contributions are welcome.
If they are not small changes, please make a GH issue and/or contact me first.

## Funding

This project is based on [my work for a bountied challenge by Justin Drake](https://github.com/protolambda/beacon-challenge)
 to write the ETH 2.0 spec in 1024 lines. A ridiculous challenge, but it was fun, and proved to be useful: 
 every line of code in the spec got extra attention, and it was a fun way to get started on an executable spec.
A month later I (@protolambda) started working for the EF,
 and maintain ZRNT to keep up with changes, test the spec, and provide a performant core component for other ETH 2.0 Go projects.

## Contact

Core dev: [@protolambda on Twitter](https://twitter.com/protolambda)

## License

MIT, see license file.

