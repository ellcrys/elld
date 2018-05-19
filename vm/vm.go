package vm

import (
	"fmt"
	"time"

	logger "github.com/ellcrys/druid/util/logger"
	"github.com/mholt/archiver"
	homedir "github.com/mitchellh/go-homedir"
)

var vmLog logger.Logger

func init() {
	vmLog = logger.NewLogrus()
}

//VM struct for Deploying and Invoking smart contracts
type VM struct {
	Containers             map[string]*Container
	log                    logger.Logger
	InvokeResponseListener func(interface{})
}

//Contract .. a contract block
type Contract struct {
	ID           string
	Transactions []Transaction
	Archive      string
}

//EllcrysBlock .. an interface for the ellcrys block to be processed by VM
type EllcrysBlock interface {
	GetContracts() []Contract
}

//TempPath where contracts are stored
const TempPath = "/.ellcrys/tmp/"

//New creates a new instance VM
//It is responsible for creating and managing contract containers
func New() *VM {
	containers := make(map[string]*Container)
	return &VM{
		log:        vmLog,
		Containers: containers,
	}
}

//Exec  a smart contract
func (vm *VM) Exec(block *Contract) bool {
	done := make(chan bool)
	vm.log.Info(fmt.Sprintf("Executing contract %s", block.ID))
	container := vm.Containers[block.ID]

	vm.log.Info(fmt.Sprintf("Connecting to server at %s", block.ID))
	err := container.Connect()
	if err != nil {
		vm.log.Error(fmt.Sprintf("cannot connect to container: %s", err))
		return false
	}

	go func() {
		for _, tx := range block.Transactions {
			vm.log.Info(fmt.Sprintf("Invoking a function %s", tx.Function))
			//container.OnResponse(vm.InvokeResponseListener)
			err = container.Invoke(&tx)
			if err != nil {
				vm.log.Error(fmt.Sprintf("invocation error: %v", err))
				done <- false
			}
			time.Sleep(1 * time.Second)
		}
		done <- true
	}()

	return <-done
}

//Init prepares the ellcrys block to be processed
func (vm *VM) Init(ellblock EllcrysBlock) bool {
	done := make(chan bool)
	go func() {
		for _, contract := range ellblock.GetContracts() {
			err := vm.deploy(&contract)
			if err != nil {
				vmLog.Error("initialization error:", err)
				done <- false
				break
			}
		}
		done <- true
	}()

	return <-done
}

//Stop the VM
func (vm *VM) Stop() {
	go func() {
		for _, container := range vm.Containers {
			ID, err := container.Destroy()
			if err != nil {
				vmLog.Error("could not destroy container ID", ID, err)
			}
		}
	}()
}

//Deploy a new contract project
func (vm *VM) deploy(config *Contract) error {
	//Unzip archive to tmp path
	usrdir, err := homedir.Dir()
	if err != nil {
		return err
	}
	//Save contrtact at temp path with folder named after it's ID. E.g: /usr/home/.ellcrys/tmp/83545762936
	outputDir := fmt.Sprintf("%s%s%s", usrdir, TempPath, config.ID)

	err = archiver.Zip.Open(config.Archive, outputDir)
	if err != nil {
		return fmt.Errorf("could not decompress archive %s", err)
	}

	vm.log.Info(fmt.Sprintf("Contract Deployed %s %s", config.ID, "√"))

	//Spawn the container
	container, err := NewContainer(config.ID)
	if err != nil {
		vmLog.Fatal("Container initialization failed %s", err)
	}
	//add spawned container to list of running containers
	vm.Containers[config.ID] = container

	//start container
	return vm.Containers[config.ID].Start()
}
