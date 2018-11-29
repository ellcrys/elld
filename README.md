![Ellcrys Network](https://storage.googleapis.com/ellcrys-docs/ellcrys-github-banner.png)

# Elld - Official Ellcrys Client
[![GoDoc](https://godoc.org/github.com/ellcrys/elld?status.svg)](https://godoc.org/github.com/ellcrys/elld)
[![CircleCI](https://circleci.com/gh/ellcrys/elld/tree/master.svg?style=svg)](https://circleci.com/gh/ellcrys/elld/tree/master)
[![Gitter chat](https://badges.gitter.im/gitterHQ/gitter.png)](https://gitter.im/ellnet)
[![Go Report Card](https://goreportcard.com/badge/github.com/ellcrys/elld)](https://goreportcard.com/report/github.com/ellcrys/elld)

Elld is the official client implemention of the Ellcrys protocol specification. It will allow users run and maintain a full network node that is capable of performing all the operations described in the [whitepaper](https://storage.googleapis.com/ellcrys-docs/Ellcrys-Whitepaper-Technical.pdf). The project is actively being developed and not ready for production use. To learn more about the Ellcry project, visit our [website](https://ellcrys.co) and [blog](https://medium.com/ellcrys).

## Tasks
- [x] **Cryptocurrency**: 
On complemetion of this task, the client will be able to join the network, mine blocks, transfer the 
native coin, achieve consensus using Bitcoin's Nakamoto consensus and provide a Javascript environment for constructing
custom behaviours and interacting with the client.
   - [x] Account-based Architecture
   - [x] Nakamoto Consensus
   - [x] RPC Client/Server
   - [x] Javacript Console

- [ ] **Hybrid PoW/PoS Consensus & Mining Protocol:**
Introduces a new consensus mechanism that will pave the way for faster network through-put and security. Additionally, a new mining protocol ([PeopleMint](https://storage.googleapis.com/ellcrys-docs/PeopleMint.pdf)) will be implemented.

- [ ] **Git Hosting:** 
Brings the ability to decentralize a git repository on the Ellcrys network. 

- [ ] **Self-Executing Functions:** 
Adds support for compiling and executing self-executing functions. Must support functions written in multiple established languages.

## Documentation
- [Documentation](https://ellcrys.gitbook.io/ellcrys/)
- [Go Documentation](https://godoc.org/github.com/ellcrys/elld)

## Requirements
Tested with [Go](http://golang.org/) 1.10.

## Contributing
We use [Dep](https://github.com/golang/dep) tool to manage project dependencies. You will need it to create deterministic builds with other developers.

## Get the Dep
Checkout the Dep [documentation](https://github.com/golang/dep#installation) for installation guide.

## Tests

Run all tests
```
make test
```

Run individual tests
```
go test ./<path to module>/...
```

## Get the source and build
```
git clone https://github.com/ellcrys/elld $GOPATH/src/github.com/ellcrys/elld
cd $GOPATH/src/github.com/ellcrys/elld
make deps
go build
```

## Contact
- Email: hello@ellcrys.co
- [Discord](https://discord.gg/QH2n2hT)
- [Gitter](https://gitter.im/ellnet)
- [Twitter](https://twitter.com/ellcryshq)
