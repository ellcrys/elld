package vm

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"sync"

	"github.com/ellcrys/druid/blockcode"

	"github.com/cenkalti/rpc2"
	"github.com/cenkalti/rpc2/jsonrpc"
	"github.com/phayes/freeport"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	. "github.com/ellcrys/druid/blockcode"
	logger "github.com/ellcrys/druid/util/logger"
	"github.com/ellcrys/druid/wire"
)

// ContainerManager defines the module that manages containerization of Block codes and their execution
type ContainerManager struct {
	containers    map[string]*Container
	containerLock *sync.Mutex
	log           logger.Logger
	client        *client.Client
	wg            *sync.WaitGroup
}

// ContainerTransaction defines the structure of the blockcode transaction
type ContainerTransaction struct {
	*wire.Transaction
	Function string `json:"Function"`
	Data     []byte `json:"Data"`
}

// NewContainerManager creates an instance of ContainerManager
func NewContainerManager(logger logger.Logger, dockerClient *client.Client) *ContainerManager {

	return &ContainerManager{
		log:    logger,
		client: dockerClient,
	}
}

// Create container that holds a blockcode
// - & add the container to containers list
func (cm *ContainerManager) create(ID string) (*Container, error) {

	imgB := NewImageBuilder(cm.log, cm.client, fmt.Sprintf(dockerFileURL, dockerFileHash))
	image := imgB.getImage()

	port, err := freeport.GetFreePort()
	if err != nil {
		return nil, err
	}

	cb, err := cm.client.ContainerCreate(context.Background(), &container.Config{
		Image: image.ID,
		Labels: map[string]string{
			"maintainer":    "Ellcrys",
			"image-version": dockerFileHash,
		},
		ExposedPorts: nat.PortSet{
			nat.Port("4000/tcp"): {},
		},
	}, &container.HostConfig{
		PortBindings: nat.PortMap{
			nat.Port("4000/tcp"): []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: strconv.Itoa(port)}},
		},
	}, &network.NetworkingConfig{}, ID)
	if err != nil {
		return nil, err
	}

	co := new(Container)
	co.dockerCli = cm.client
	co.id = cb.ID
	co.log = cm.log
	co.port = port

	cm.containers[ID] = co

	return co, nil

}

// Run the blockcodes in a block chain
// - create container
// - copy block code content into container
// - build block code
// - run the block code
// - - get run script from LangBuilder instance
// - - execute blockcode with run script
// - - create container rpc client
// - - create bi-directional connection to container
// - - send transaction with container client
func (cm *ContainerManager) Run(tx *ContainerTransaction, txID string, blockcodes []Blockcode, txOutput chan []byte, done chan error) {

	container, err := cm.create(txID)
	if err != nil {
		done <- err
		return
	}

	for _, bcode := range blockcodes {
		content := bcode.Code
		err := container.copy(txID, content)
		if err != nil {
			done <- err
			return
		}

		switch bcode.Manifest.Lang {
		case blockcode.LangGo:
			go container.build(cm.containerLock, txOutput, done)
		}

		runScript := container.buildConfig.GetRunScript()
		go container.exec(runScript, txOutput, done)

		addr := fmt.Sprintf("127.0.0.1:%s", strconv.Itoa(container.port))
		io, _ := net.Dial("tcp", addr)
		codec := jsonrpc.NewJSONCodec(io)
		rpcCli := rpc2.NewClientWithCodec(codec)

		container.client = rpcCli
		container.client.Handle("response", func(vm *rpc2.Client, data []byte, reply *struct{}) error {
			txOutput <- data
			return nil
		})

		go container.client.Run()

		err = container.client.Call("invoke", tx, nil)
		if err != nil {
			done <- err
		}

	}

}

// Find looks up a container by it's ID
func (cm *ContainerManager) find(ID string) *Container {
	container := cm.containers[ID]

	if container == nil {
		return nil
	}

	return container
}

// Len returns length of running containers
func (cm *ContainerManager) len() (int, error) {
	count := 0
	for _, container := range cm.containers {
		resp, err := cm.client.ContainerInspect(context.Background(), container.id)
		if err != nil {
			return 0, err
		}

		if resp.State.Running {
			count++
		}
	}

	return count, nil
}

// Stop a container
func (cm *ContainerManager) stop(ID string) error {
	container := cm.find(ID)

	err := container.stop()
	if err != nil {
		return err
	}

	err = container.destroy()
	if err != nil {
		return err
	}

	return nil
}
