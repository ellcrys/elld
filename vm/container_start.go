package vm

//TerminationData ..
type TerminationData struct {
	ID string `json:"ContractID"`
}

//Start the container
// func (container *Container) Start() error {
// 	ctx := context.Background()
// 	//Start the container
// 	return container.client.ContainerStart(ctx, container.ID, types.ContainerStartOptions{})
// }

// //Connect to a container
// func (container *Container) Connect() error {
// 	//container address
// 	addr := "0.0.0.0:" + strconv.Itoa(container.port)
// 	conn, err := net.Dial("tcp", addr)
// 	if err != nil {
// 		return fmt.Errorf("dial failed err: %v", err)
// 	}
// 	//Dial container
// 	service := rpc2.NewClientWithCodec(jsonrpc.NewJSONCodec(conn))
// 	if err != nil {
// 		return fmt.Errorf("client connection failed err: %v", err)
// 	}

// 	vmLog.Info(fmt.Sprintf("VM connected to contract at %s", conn.LocalAddr().String()))

// 	container.service = service

// 	//Handle response to container

// 	container.service.Handle("response", container.responeListener)

// 	//Handle terminate request from contract
// 	container.service.Handle("terminate", container.responeListener)

// 	return nil
// }
