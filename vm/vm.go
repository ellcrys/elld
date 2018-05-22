package vm

import (
	"fmt"
	"os"

	"github.com/ellcrys/druid/util"

	logger "github.com/ellcrys/druid/util/logger"
)

// MountDir where contracts are stored
var MountDir = "mountdir"

// VM specializes in executing transactions against a contracts
type VM struct {
	containers             map[string]*Container
	log                    logger.Logger
	containerMountDir      string
	InvokeResponseListener func(interface{})
}

// New creates a new instance of VM
func New(log logger.Logger, containerMountDir string) *VM {
	vm := new(VM)
	vm.log = log
	vm.containerMountDir = containerMountDir
	vm.containers = map[string]*Container{}
	return vm
}

// Init sets up the environment for execution of contracts.
// - Check if docker daemon is accessible
// - Check if container mount directory exists, otherwise create it
// - Check if docker image exists, if not, fetch and build the image
func (vm *VM) Init() error {

	if err := dockerAlive(); err != nil {
		return fmt.Errorf("docker not running. %s", err)
	}

	if !util.IsPathOk(vm.containerMountDir) {
		if err := os.MkdirAll(vm.containerMountDir, 0700); err != nil {
			return fmt.Errorf("failed to create container mount directory. %s", err)
		}
	}

	return nil
}
