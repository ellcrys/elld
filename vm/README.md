# The Ellcrys Virtual Machine
The Ellcrys virtual machine module manages and executes Blockcodes. Blockcode is a protocol that facilitate, verify, or enforce the negotiation or performance of a digital contract on the ellcrys blockchain. 
This implements the  phenomenon of "[Smart Contracts](http://www.fon.hum.uva.nl/rob/Courses/InformationInSpeech/CDROM/Literature/LOTwinterschool2006/szabo.best.vwh.net/smart.contracts.html)", a phrase coined by [Nick Szabo](https://en.wikipedia.org/wiki/Nick_Szabo), a computer scientist and reknowned cryptographer. 

> A smart contract is a computerized transaction protocol that executes the terms of a contract
> - Nick Szabo


The vm is polyglot and thus a blockcode can be written in several supported computer programming languages such as `Go`, `Python`, `Lua` or `Typescript`

## Blockcode Execution Model
The blockcode execution concept is quite similar to the concept of Remote Method Invocation in Java, except that in our case blockcode objects are part of transactions in the ellcrys blockchain and can be hosted on the decentralized ellcrys network. For each transaction in an ellcrys blockchain, a transaction may contain a blockcode. The VM filters through, and executes a blockcode in a separate self-contained process called a [container](https://www.docker.com/what-container). Containers hold necessary components required to execute a blockcode, they run in separate instances with limited interference by the host operating system. 

### How It Works
Imagine a set of transactions in a block, where *B* represents a block and *T* represents a transaction in a block:

##### i. *B* := { T . . . n }
#### ii. *T* ∈ *B*

If *T* contains a blockcode *b*: 

#### iii. IF [ *b* ∈ *T* ] == TRUE
A new container *C* is created with the blockcode *b* loaded inside it, and it's then spawned to execute the blockcode:
#### iv. *C* := { *b* }
#### v. SPAWN(C)

As soon as a container starts, the blockcode starts a websocket server. Then the VM connects to the blockcode server and invokes the functions and payload defined in the blockcode transaction. 

![](https://bitbucket.org/Damilare_/ell/downloads/untitled_page.png)

#### Principle
1. The VM holds a reference list of all spawned containers
2. The VM must create and maintain a bi-directional communication between itself and the blockcode
3. The VM can be authorized to invoke a blockcode function
4. The VM can be authorized to terminate a blockcode execution

## Anatomy of the VM module


## Writing a blockcode (Go example)

A blockcode `struct` must implement an `OnInit()` function, which will be triggered by the VM as soon as a blockcode execution starts, and custom functions which should be executed by the VM. 

```go
package main

import (
	. "github.com/ellcrys/go.stub"
)

// BlockcodeEx defines a blockcode
type BlockcodeEx struct {
}

// OnInit initializes the blockcode
func (b *BlockcodeEx) OnInit() {
	On("add10", b.add)
}

func (b *BlockcodeEx) add() (interface{}, error) {
	return 5+5, nil
}

func main() {
	Run(new(BlockcodeEx))
}

```
In this example, we have a custom function `add()`, and the `OnInit()` method which instructs that the `add10` invocation name should be bound to the instance function `add()` of the blockcode.  

This means when the VM invokes the `add10` command `b.add()` will be executed, where `b` is an instance of `BlockcodeEx`, the return value `10` of function `add()` will then be returned to the VM client as a response.

Every blockcode package must contain a `package.json` file that describes the blockcode and it's dependencies. 

```json
 {
  "name": "example_blockcode",
  "version": "0.0.1",
  "description": "An example blockcode",
  "main": "main.go",
  "author": "",
  "license": "ISC",
  "lang": "go",
  "langVer": "1.10.2",
  "publicFuncs": ["add10"]
}
```

* ***name*** sets the name of your blockcode
* ***version*** sets the version of your blockcode
* ***description*** describes what your blockcode does
* ***main*** sets the entry point for your blockcode execution, e.g `main.go` for golang, or `index.ts` for typescript
* ***lang*** tells the VM what programming language your blockcode is written in
* ***langVer*** sets the version of environment that your blockcode should be executed in
* ***publicFuncs*** you can restrict the VM to execute only some set of functions












