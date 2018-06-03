package stub

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/googollee/go-socket.io"
)

var (
	defaultStub *stub
	room        = "_"
	server      *socketio.Server
)

func _init() {
	defaultStub = new(stub)
	defaultStub.mtx = &sync.Mutex{}
	defaultStub.funcs = make(map[string]invocableFunc)
	defaultStub.wait = make(chan bool)
}

func init() {
	_init()
}

// invocableFunc represents a function that can be invoked externally
type invocableFunc func() (interface{}, error)

// Blockcode defines the structure of a blockcode
type Blockcode interface {
	// OnInit is always called everytime a function is executed.
	// This is where the blockcode performs initialization like,
	// registering invocable functions using `On` and setting default state
	OnInit()
}

// stub provides functionalities for blockcode execution
// and provides the blockcode with the ability to mutate,
// respond to function invocation and call other contracts.
type stub struct {
	mtx       *sync.Mutex
	wait      chan bool
	blockcode Blockcode                // Blockcode to initialize
	funcs     map[string]invocableFunc // Functions on the blockcode that can be invoked externally. Protected by mxt
}

// Run runs a blockcode
func Run(b Blockcode) {

	if b == nil {
		panic(fmt.Errorf("blockcode not initialized"))
	}

	defaultStub.blockcode = b

	go func() {
		if err := serve(); err != nil {
			panic(err)
		}
	}()

	<-defaultStub.wait
}

// On registers a function
func On(name string, f invocableFunc) {

	if f == nil {
		return
	}

	defaultStub.mtx.Lock()
	defer defaultStub.mtx.Unlock()
	defaultStub.funcs[name] = f
}

func getFunc(name string) invocableFunc {
	defaultStub.mtx.Lock()
	defer defaultStub.mtx.Unlock()
	return defaultStub.funcs[name]
}

func reset() {
	_init()
}

func serve() error {

	var err error
	service := newService(defaultStub)
	server, err = socketio.NewServer(nil)
	if err != nil {
		return fmt.Errorf("failed to create server")
	}

	server.On("connection", func(so socketio.Socket) {
		so.Join(room)
		log.Println("Connection established")
	})

	server.On("invoke", func(args Args) *Result {
		log.Println("Received new invoke request")
		return service.Invoke(args)
	})

	server.On("echo", func(args Args) *Args {
		return &args
	})

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/", server)
		log.Println("Stub server now running at :9900")
		http.ListenAndServe(":9900", mux)
	}()

	return nil
}

func stopService() {
	if defaultStub == nil {
		return
	}
	close(defaultStub.wait)
}
