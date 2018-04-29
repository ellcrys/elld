package vm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ellcrys/druid/util"
	pb "github.com/ellcrys/druid/wire"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var vmLog *zap.SugaredLogger

func init() {
	vmLog = util.NewLogger("/vm")
}

//VM struct for Deploying and Invoking smart contracts
type VM struct {
	container     *Container
	serviceClient pb.ContractServiceClient
	log           *zap.SugaredLogger
}

//DeployConfig deploy configuration struct
type DeployConfig struct {
	path string
}

//InvokeConfig invoke configuration struct
type InvokeConfig struct {
	funcName string
	data     []byte
}

//InvokeResponseData is the structure of data expected from Invoke request
type InvokeResponseData struct {
	message string
	code    int
}

//Deploy a new contract project
func (vm *VM) Deploy(config *DeployConfig) error {
	//...
	return nil
}

//Invoke a smart contract function
func (vm *VM) Invoke(config *InvokeConfig) error {
	ctx := context.Background()
	resp, err := vm.serviceClient.ContractInvoke(ctx, &pb.ContractInvokeRequest{
		Function: config.funcName,
		Data:     config.data,
	})

	if err != nil {
		vmLog.Errorf("Could not invoke function %s", config.funcName)
		return fmt.Errorf("Could not invoke function %s", config.funcName)
	}

	//response status
	status := resp.GetStatus()
	//response data
	data := resp.GetData()
	var invokeResp InvokeResponseData
	err = json.Unmarshal(data, &invokeResp)
	if err != nil {
		vmLog.Errorf("Error reading response %s", err)
		return fmt.Errorf("Error reading response %s", err)
	}

	//Handle error response from function
	if status != "" && status == "error" {
		vmLog.Errorf("Code: %d => function %s returned an error %s", invokeResp.code, invokeResp.message, config.funcName)
		return fmt.Errorf("Code: %d => function %s returned an error %s", invokeResp.code, invokeResp.message, config.funcName)
	}

	//Handle success response from function
	if status != "" && status == "success" {
		vmLog.Infof("Code: %d %s", invokeResp.code, "√")
	}

	return nil
}

//NewVM create a new instance VM
//It is responsible for creating a new container and setup a bidirectional stream
func NewVM() *VM {

	port := 4000
	targetPort := 4000
	//Create new container instance
	container, err := NewContainer(port, targetPort, "")
	if err != nil {
		vmLog.Fatalf("Container initialization failed %s", err)
	}

	//container address
	addr := fmt.Sprintf("http://127.0.0.1:%d", port)

	//Dial container
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		vmLog.Fatalf("Dial failed err: %v", err)
	}
	//Create a new grpc client
	client := pb.NewContractServiceClient(conn)

	return &VM{
		container:     container,
		serviceClient: client,
		log:           vmLog,
	}
}
