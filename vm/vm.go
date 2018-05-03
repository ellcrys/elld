package vm

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/ybbus/jsonrpc"

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
	containers map[string]*Container
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
	Function   string `json:"FuncName"`
	Data       interface{}
}

//InvokeResponseData is the structure of data expected from Invoke request
type InvokeResponseData struct {
	Status string `json:"status"`
	Code   int    `json:"code"`
	Data   struct {
		Message   string `json:"message"`
		ReturnVal string `json:"returnVal"`
	}
}

//TempPath where contracts are stored
const TempPath = "/.ellcrys/tmp/"

//spawn
func spawn(contractID string) (*Container, jsonrpc.RPCClient) {

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

	return container, conn
}

//Deploy a new contract project
func (vm *VM) Deploy(config *DeployConfig) error {
	//verify if archive is valid
	// vmLog.Infof("Verifying archive")
	// signer := NewSigner()
	// err := signer.Verify(config.archive)
	// if err != nil {
	// 	vmLog.Errorf("Verification Failed: Invalid archive %s", err)
	// 	return fmt.Errorf("Verification Failed: Invalid archive %s", err)
	// }

	// vmLog.Infof("Contract verification passed %s %s", config.contractID, "√")

	//Unzip archive to tmp path
	usrdir, err := homedir.Dir()
	if err != nil {
		return err
	}
	//Save contrtact at temp path with folder named after it's ID. E.g: /usr/home/.ellcrys/tmp/83545762936
	outputDir := fmt.Sprintf("%s%s%s", usrdir, TempPath, config.ContractID)

	err = archiver.Zip.Open(config.Archive, outputDir)
	if err != nil {
		vmLog.Error("Could not decompress archive %s", err)
		return fmt.Errorf("Could not decompress archive %s", err)
	}

	vmLog.Info("Contract Deployed %s %s", config.ContractID, "√")

	return nil
}

//Invoke a smart contract
func (vm *VM) Invoke(config *InvokeConfig) error {
	container, rpc := spawn(config.ContractID)

	//add spawned container to list of running containers
	vm.containers[config.ContractID] = container

	var args [1]*InvokeConfig
	args[0] = config

	resp, err := rpc.Call("invoke", args)

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
		log.Fatal(err)
	}
	//response status
	status := res.Status
	//response data
	data := res.Data

	if err != nil {
		vmLog.Error("Error reading response %s", err)
		return fmt.Errorf("Error reading response %s", err)
	}

	//Handle error response from function
	if status != "" && status == "error" {
		vmLog.Error("Code: %d => function %s returned an error %s", res.Code, data.Message, config.Function)
		return fmt.Errorf("Code: %d => function %s returned an error %s", res.Code, data.Message, config.Function)
	}

	//Handle success response from function
	if status != "" && status == "success" {
		vmLog.Info("Code: %d %s", res.Code, "√")
	}

	return nil
}

//NewVM create a new instance VM
//It is responsible for creating and managing contract containers
func NewVM() *VM {
	containers := make(map[string]*Container)
	return &VM{
		log:        vmLog,
		containers: containers,
	}
}
