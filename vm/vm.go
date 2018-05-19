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

// // Exec a smart contract
// func (vm *VM) Exec(block *Contract) bool {
// 	done := make(chan bool)
// 	vm.log.Info(fmt.Sprintf("Executing contract %s", block.ID))
// 	container := vm.containers[block.ID]

// 	vm.log.Info(fmt.Sprintf("Connecting to server at %s", block.ID))
// 	err := container.Connect()
// 	if err != nil {
// 		vm.log.Error(fmt.Sprintf("cannot connect to container: %s", err))
// 		return false
// 	}

// 	go func() {
// 		for _, tx := range block.Transactions {
// 			vm.log.Info(fmt.Sprintf("Invoking a function %s", tx.Function))
// 			// container.OnResponse(vm.InvokeResponseListener)
// 			err = container.Invoke(&tx)
// 			if err != nil {
// 				vm.log.Error(fmt.Sprintf("invocation error: %v", err))
// 				done <- false
// 			}
// 			time.Sleep(1 * time.Second)
// 		}
// 		done <- true
// 	}()

// 	return <-done
// }

// Init prepares the ellcrys block to be processed
// func (vm *VM) Init(ellblock EllcrysBlock) bool {
// 	done := make(chan bool)
// 	go func() {
// 		for _, contract := range ellblock.GetContracts() {
// 			err := vm.deploy(&contract)
// 			if err != nil {
// 				log.Error("initialization error:", err)
// 				done <- false
// 				break
// 			}
// 		}
// 		done <- true
// 	}()

// 	return <-done
// }

// Stop the VM
// func (vm *VM) Stop() {
// 	go func() {
// 		for _, container := range vm.containers {
// 			ID, err := container.Destroy()
// 			if err != nil {
// 				log.Error("could not destroy container ID", ID, err)
// 			}
// 		}
// 	}()
// }

// Deploy a new contract project
// func (vm *VM) deploy(config *Contract) error {
// 	// Unzip archive to tmp path
// 	usrdir, err := homedir.Dir()
// 	if err != nil {
// 		return err
// 	}
// 	// Save contrtact at temp path with folder named after it's ID. E.g: /usr/home/.ellcrys/tmp/83545762936
// 	outputDir := fmt.Sprintf("%s%s%s", usrdir, MountDir, config.ID)

// 	err = archiver.Zip.Open(config.Archive, outputDir)
// 	if err != nil {
// 		return fmt.Errorf("could not decompress archive %s", err)
// 	}

// 	vm.log.Info(fmt.Sprintf("Contract Deployed %s %s", config.ID, "√"))

// 	// Spawn the container
// 	container, err := NewContainer(config.ID)
// 	if err != nil {
// 		log.Fatal("Container initialization failed %s", err)
// 	}
// 	// add spawned container to list of running containers
// 	vm.containers[config.ID] = container

// 	// start container
// 	return vm.containers[config.ID].Start()
// }
