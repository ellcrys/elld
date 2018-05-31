package vm

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"sync"

	"github.com/phayes/freeport"

	"github.com/cenkalti/rpc2"
	"github.com/cenkalti/rpc2/jsonrpc"
	"github.com/ellcrys/druid/blockcode"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"

	logger "github.com/ellcrys/druid/util/logger"
	"github.com/ellcrys/druid/wire"
)

// ContainerManager defines the module that manages containerization of Block codes and their execution
type ContainerManager struct {
	containers    map[string]*Container
	containerLock *sync.Mutex
	logger        logger.Logger
	client        *client.Client
	blockchain    Blockchain
	wg            *sync.WaitGroup
}

// ContainerTransaction defines the structure of the blockcode transaction
type ContainerTransaction struct {
	Tx       *wire.Transaction `json:"Tx"`
	Function string            `json:"Function"`
	Data     []byte            `json:"Data"`
}

type InvokeData struct {
	Function    string      `json:"Function"`
	Data        interface{} `json:"Data"`
	BlockCodeID string      `json:"BlockCodeID"`
}

type Response struct {
	Status string `json:"status"`
	Code   int    `json:"code"`
	Data   []byte `json:"data"`
}

type SampleBlockchain struct {
	blockcodes map[string]*blockcode.Blockcode
}

func NewSampleBlockchain() *SampleBlockchain {
	b := new(SampleBlockchain)
	bc, err := blockcode.FromDir("./testdata/blockcode_example")
	if err != nil {
		panic(err)
	}
	b.blockcodes = map[string]*blockcode.Blockcode{
		"some_address": bc,
	}
	return b
}

func (b *SampleBlockchain) GetBlockCode(address string) *blockcode.Blockcode {
	return b.blockcodes[address]
}

// NewContainerManager creates an instance of ContainerManager
func NewContainerManager(log logger.Logger, dockerClient *client.Client) *ContainerManager {
	cm := new(ContainerManager)
	cm.logger = log
	cm.client = dockerClient
	cm.containers = make(map[string]*Container)
	cm.containerLock = &sync.Mutex{}
	cm.wg = &sync.WaitGroup{}
	cm.blockchain = NewSampleBlockchain()
	return cm
}

// Create container that holds a blockcode
// - & add the container to containers list
func (cm *ContainerManager) create(ID string) (*Container, error) {

	imgB := NewImageBuilder(cm.logger, cm.client, fmt.Sprintf(dockerFileURL, dockerFileHash))
	image := imgB.getImage()

	port, err := freeport.GetFreePort()
	if err != nil {
		return nil, err
	}

	vol := make(map[string]struct{})
	vol["/go/src/contract/"+ID] = struct{}{}

	cb, _ := cm.client.ContainerCreate(context.Background(), &container.Config{
		Image: image.ID,
		Labels: map[string]string{
			"maintainer":    "Ellcrys",
			"image-version": dockerFileHash,
		},
		Volumes: vol,
		ExposedPorts: nat.PortSet{
			nat.Port("4000/tcp"): {},
		},
	}, &container.HostConfig{
		PortBindings: nat.PortMap{
			nat.Port("4000/tcp"): []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: strconv.Itoa(port)}},
		},
	}, &network.NetworkingConfig{}, "")

	co := new(Container)
	co.dockerCli = cm.client
	co.id = cb.ID
	co.log = cm.logger
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
func (cm *ContainerManager) Run(tx *wire.Transaction, txOutput chan []byte, done chan error) {

	bcode := cm.blockchain.GetBlockCode(tx.To)
	content := bcode.Code

	container, err := cm.create(bcode.ID())
	if err != nil {
		done <- err
		return
	}

	err = container.start()
	if err != nil {
		done <- err
		return
	}

	err = container.copy(bcode.ID(), content)
	if err != nil {
		done <- err
		return
	}

	switch bcode.Manifest.Lang {
	case blockcode.LangGo:
		cm.wg.Add(1)
		defer cm.wg.Wait()
		goBuilder := newGoBuilder(bcode.ID(), container, cm.logger)
		container.setBuildLang(goBuilder)
		err := container.build()
		if err != nil {
			done <- err
			return
		}
		cm.wg.Done()
	}

	runScript := container.buildConfig.GetRunScript()
	err = container.exec(runScript, nil)
	if err != nil {
		done <- err
		return
	}

	addr := fmt.Sprintf("0.0.0.0:%s", strconv.Itoa(container.port))
	io, _ := net.Dial("tcp", addr)

	codec := jsonrpc.NewJSONCodec(io)
	rpcCli := rpc2.NewClientWithCodec(codec)

	container.client = rpcCli
	container.client.Handle("response", func(vm *rpc2.Client, data *Response, reply *struct{}) error {
		cm.logger.Debug(fmt.Sprintf("Exec Response: %s", string(data.Data)))
		txOutput <- data.Data
		return nil
	})

	go container.client.Run()

	go func() {
		err = container.client.Call("invoke", &InvokeData{
			Function:    tx.BlockcodeParams.GetFunc(),
			BlockCodeID: bcode.ID(),
			Data:        tx.BlockcodeParams.GetData(),
		}, nil)
		if err != nil {
			done <- err
			return
		}
	}()

	done <- nil
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
