package vm

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/ybbus/jsonrpc"

	"net/http"
	netrpc "net/rpc"

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
func spawn(contractID string) *Container {

	//Create new container instance

	var err error

	container, err := NewContainer(contractID)
	if err != nil {
		vmLog.Fatal("Container initialization failed %s", err)
	}

	//container address
	addr := fmt.Sprintf("http://127.0.0.1:%d", container.port)

	//Dial container
	conn := jsonrpc.NewClient(addr)
	if err != nil {
		vmLog.Fatal("Dial failed err: %v", err)
	}

	container.service = conn

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

	err = archiver.Zip.Open(config.Archive, outputDir)
	if err != nil {
		vmLog.Error(fmt.Sprintf("Could not decompress archive %s", err))
		return fmt.Errorf("Could not decompress archive %s", err)
	}

	vmLog.Info(fmt.Sprintf("Contract Deployed %s %s", config.ContractID, "√"))

	//Spawn the container
	container := spawn(config.ContractID)

	//add spawned container to list of running containers
	vm.Containers[config.ContractID] = container

	return nil
}

//Invoke a smart contract
func (vm *VM) Invoke(config *InvokeConfig) error {
	var args [1]*InvokeConfig
	args[0] = config

	//Fetch contract container and invoke
	container := vm.Containers[config.ContractID]
	resp, err := container.service.Call("invoke", args)

	if err != nil {
		vmLog.Error(fmt.Sprintf("Could not invoke function %s : %s", config.Function, err))
		return fmt.Errorf("Could not invoke function %s : %s", config.Function, err)
	}

	var res *InvokeResponseData

	//Get RPC results
	details, _ := json.Marshal(resp.Result)

	//Pass results unto res pointer
	err = json.Unmarshal(details, &res)
	if err != nil {
		vmLog.Error(fmt.Sprintf("Could not retrieve response from %s : %s", config.Function, err))
		return fmt.Errorf("Could not retrieve response from %s : %s", config.Function, err)
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
		vmLog.Error(fmt.Sprintf("Code: %d => function %s returned an error %s", res.Code, data.Message, config.Function))
		return fmt.Errorf("Code: %d => function %s returned an error %s", res.Code, data.Message, config.Function)
	}

	//Handle success response from function
	if status != "" && status == "success" {
		vmLog.Info(fmt.Sprintf("Code: %d function %s invoked at Contract:%s %s", res.Code, config.Function, config.ContractID, "√"))
		vmLog.Info(fmt.Sprintf("Returned response from Contract: %s => %v", config.ContractID, res))
	}

	return nil
}

//Terminate a running contract
func (vm *VM) Terminate(contractID string) (ID string, err error) {
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

	//Register request handler
	netrpc.Register(new(RequestHandler))

	netrpc.HandleHTTP()

	l, e := net.Listen("tcp", ":4000")
	if e != nil {
		vmLog.Fatal("listen error: %s", e)
	}

	go startServer(l)

	return &VM{
		log:        vmLog,
		Containers: containers,
	}
}

func startServer(l net.Listener) {
	vmLog.Info("VM Server listening at 4000")
	err := http.Serve(l, nil)

	if err != nil {
		vmLog.Fatal(fmt.Sprintf("Error serving: %s", err))
	}

}
