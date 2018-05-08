package vm

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"sync"

	logger "github.com/ellcrys/druid/util/logger"
	"github.com/ellcrys/rpc2"
	"github.com/mholt/archiver"
	homedir "github.com/mitchellh/go-homedir"
)

var vmLog logger.Logger

func init() {
	vmLog = logger.NewLogrus()
}

//VM struct for Deploying and Invoking smart contracts
type VM struct {
	Containers map[string]*Container
	log        logger.Logger
}

//DeployConfig deploy configuration struct
type DeployConfig struct {
	ContractID string // contract id
	Archive    string // path to archive where contract is saved
}

//InvokeConfig invoke configuration struct
type InvokeConfig struct {
	ContractID string `json:"ContractID"`
	Function   string `json:"Function"`
	Data       interface{}
}

//InvokeResponseData is the structure of data expected from Invoke request
type InvokeResponseData struct {
	Status string `json:"status"`
	Code   int    `json:"code"`
	Data   struct {
		Message   string      `json:"message"`
		ReturnVal interface{} `json:"returnVal"` //Json string or a string value
	}
}

//RequestHandler for VM rpc connection
type RequestHandler struct{}

//Terminate a contract
func (t *RequestHandler) Terminate(val string, reply *string) error {

	return nil
}

//TempPath where contracts are stored
const TempPath = "/.ellcrys/tmp/"

//spawn
func (vm *VM) spawn(contractID string) *Container {

	//Create new container instance

	var err error

	container, err := NewContainer(contractID)
	if err != nil {
		vmLog.Fatal("Container initialization failed %s", err)
	}

	//container address
	addr := "127.0.0.1:" + strconv.Itoa(container.port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		vmLog.Fatal("Dial failed err: %v", err)
	}
	//Dial container
	client := rpc2.NewClient(conn)
	if err != nil {
		vmLog.Fatal("Dial failed err: %v", err)
	}

	container.service = client

	//Handle terminate request from container
	container.service.Handle("terminate", func(client *rpc2.Client, args interface{}, resp *interface{}) error {
		ID, err := vm.Terminate(contractID)
		if err != nil {
			return fmt.Errorf("Could not terminate Contract %s : %s", contractID, err)
		}
		vmLog.Info(fmt.Sprintf("Contract %s terminated successfully", ID))
		return nil
	})

	return container
}

//Deploy a new contract project
func (vm *VM) Deploy(config *DeployConfig) error {
	//Unzip archive to tmp path
	usrdir, err := homedir.Dir()
	if err != nil {
		return err
	}
	//Save contrtact at temp path with folder named after it's ID. E.g: /usr/home/.ellcrys/tmp/83545762936
	outputDir := fmt.Sprintf("%s%s%s", usrdir, TempPath, config.ContractID)

	var wg sync.WaitGroup
	wg.Add(1)
	defer wg.Wait()
	go func() {
		err = archiver.Zip.Open(config.Archive, outputDir)
		if err != nil {
			vmLog.Error(fmt.Sprintf("Could not decompress archive %s", err))

		}
		wg.Done()
	}()

	vmLog.Debug(fmt.Sprintf("Contract Deployed %s %s", config.ContractID, "√"))

	//Spawn the container
	container := vm.spawn(config.ContractID)

	//add spawned container to list of running containers
	vm.Containers[config.ContractID] = container

	return nil
}

//Invoke a smart contract
func (vm *VM) Invoke(config *InvokeConfig) error {
	var args *InvokeConfig
	args = config

	//Fetch contract container and invoke
	container := vm.Containers[config.ContractID]

	//Handle response to container

	container.service.Handle("response", func(client *rpc2.Client, args interface{}, resp *interface{}) error {

		buf, err := json.Marshal(args)
		if err != nil {
			panic(err)
		}
		var res *InvokeResponseData

		//Pass results unto res pointer
		err = json.Unmarshal(buf, &res)
		if err != nil {
			vmLog.Error(fmt.Sprintf("Could not retrieve response from contract %s : %s", config.ContractID, err))
			return fmt.Errorf("Could not retrieve response from contract %s : %s", config.ContractID, err)
		}
		//response status
		status := res.Status
		//response data
		data := res.Data

		if err != nil {
			vmLog.Error(fmt.Sprintf("Error reading response %s", err))
			return fmt.Errorf("Error reading response %s", err)
		}

		//Handle error response from function
		if status != "" && status == "error" {
			vmLog.Error(fmt.Sprintf("Code: %d => Contract %s returned an error %s", res.Code, config.ContractID, data.Message))
			return fmt.Errorf("Code: %d => Contract %s returned an error %s", res.Code, config.ContractID, data.Message)
		}

		//Handle success response from function
		if status != "" && status == "success" {
			vmLog.Info(fmt.Sprintf("Returned response from Contract: %s => %v", config.ContractID, res))
		}
		client.Close()
		return nil
	})

	go container.service.Run()

	_ = container.service.Call("invoke", args, nil)

	return nil
}

//Terminate a running contract
func (vm *VM) Terminate(contractID string) (ID string, err error) {
	defer vm.Containers[contractID].service.Close()

	//Find contract in list of running containers and terminate it
	ID, err = vm.Containers[contractID].Destroy()
	if err != nil {
		return "", err
	}
	return ID, nil
}

//NewVM create a new instance VM
//It is responsible for creating and managing contract containers
func NewVM() *VM {
	containers := make(map[string]*Container)

	return &VM{
		log:        vmLog,
		Containers: containers,
	}
}
