### Stub For Go

Go.Stub enables a Blockcode to seamlessly interact with the virtual machine such that its functions can be invoked and it is able to access blockchain state and services.

#### Install
Run the following command to install
```
go get github.com/ellcrys/go.stub
```

#### Usage
The stub offers a simple to use API. First, you must create a structure that implements the `Blockcode` interface. Here is a simple example of a Blockcode. 

```go
import . "github.com/ellcrys/go.stub"

type Coin struct {
    totalSupply int
}

// OnInit is called before a function is invoked. 
// Perform any initialization here. 
func (c *Coin) OnInit() {
    
    c.totalSupply = 21000000
    
    // register an invocable function
    On("getTotalSupply", c.getTotalSupply)
}

// getTotalSupply returns the coin's total supply
func (c *Coin) getTotalSupply() (interface{}, error) {
    return c.totalSupply, nil
}
```

To attach the example to the stub so that it can communicate with the VM, we pass it to `Run`

```go
Run(new(Coin))
```

#### License
MIT
