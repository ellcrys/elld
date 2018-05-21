package vm

import (
	logger "github.com/ellcrys/druid/util/logger"
)

// VM specializes in executing transactions againts a contracts
type VM struct {
	containers             map[string]*Container
	log                    logger.Logger
	InvokeResponseListener func(interface{})
}

// MountDir where contracts are stored
const MountDir = "mountdir"

// New creates a new instance of VM
func New(log logger.Logger) *VM {
	vm := new(VM)
	vm.log = log
	vm.containers = map[string]*Container{}
	return vm
}

// Init prepares the ellcrys block to be processed
func (vm *VM) Init() bool {
	return true
}
