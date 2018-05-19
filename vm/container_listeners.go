package vm

//get responses from the container
// func (container *Container) responeListener(client *rpc2.Client, res *InvokeResponseData, reply *struct{}) error {
// 	fmt.Printf("%v\n", res)

// 	//response status
// 	status := res.Status

// 	//Handle error response from function
// 	if status != "" && status == "error" {
// 		vmLog.Info(fmt.Sprintf("Code: %d => Contract %s returned an error", res.Code, container.contractID))
// 		return fmt.Errorf("Code: %d => Contract %s returned an error", res.Code, container.contractID)
// 	}

// 	//Handle success response from function
// 	if status != "" && status == "success" {
// 		vmLog.Info(fmt.Sprintf("Returned response from Contract: %s => %v", container.contractID, res))
// 	}
// 	return nil
// }

// //get command to terminate the contract
// func (container *Container) terminateListener(client *rpc2.Client, data TerminationData, reply *struct{}) error {
// 	defer container.service.Close()

// 	//Find contract in list of running containers and terminate it
// 	ID, err := container.Destroy()

// 	if err != nil {
// 		return fmt.Errorf("could not terminate Contract %s : %s", data.ID, err)
// 	}
// 	vmLog.Info(fmt.Sprintf("contract %s terminated successfully √", ID))
// 	return nil
// }
