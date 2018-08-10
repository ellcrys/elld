package logic

import (
	evbus "github.com/asaskevich/EventBus"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/util/logger"
)

// Logic provides an interface for performing state query and transition
// operations. External packages like the RPC server can utilize API
// methods in this package to create transactions, get node information
// or configure the node.
type Logic struct {
	engine types.Engine
	log    logger.Logger
	bus    evbus.Bus
}

func sendErr(errCh chan error, err error) error {
	go func() { errCh <- err }()
	return err
}

// New creates a new Logic instance. It will register
// all public logic handles to evbus.
func New(engine types.Engine, event evbus.Bus, log logger.Logger) *Logic {

	logic := new(Logic)
	logic.engine = engine
	logic.log = log
	logic.bus = event

	// transactions events
	event.Subscribe("transaction.add", logic.TransactionAdd)

	// database events
	event.Subscribe("objects.put", logic.ObjectsPut)
	event.Subscribe("objects.get", logic.ObjectsGet)

	return logic
}
