package vm

import (
	"fmt"
	"sync"
	"time"

	"github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"

	"github.com/ellcrys/elld/blockcode"
	"github.com/ellcrys/elld/params"
	"github.com/ellcrys/elld/types/core/objects"
	logger "github.com/ellcrys/elld/util/logger"
	docker "github.com/fsouza/go-dockerclient"
)

// ContainerManager defines the module that manages containerization of Block codes and their execution
type ContainerManager struct {
	cmt        *sync.Mutex
	containers map[string]*Container // Holds all running containers. Protected by cmt
	log        logger.Logger         // Logger
	client     *docker.Client        // Docker client
	blockchain Blockchain            // Blockchain
	img        *Image                // Image ran by containers
}

// ContainerTransaction defines the structure of the blockcode transaction
type ContainerTransaction struct {
	Tx       *objects.Transaction `json:"tx"`
	Function string               `json:"function"`
	Data     []byte               `json:"data"`
}

// NewContainerManager creates an instance of ContainerManager
func NewContainerManager(dockerClient *docker.Client, img *Image, log logger.Logger) *ContainerManager {
	cm := new(ContainerManager)
	cm.log = log
	cm.client = dockerClient
	cm.containers = make(map[string]*Container)
	cm.cmt = &sync.Mutex{}
	cm.blockchain = NewSampleBlockchain()
	cm.img = img
	return cm
}

// newContainer creates a new container
func (cm *ContainerManager) newContainer() (*Container, error) {
	co := NewContainer(cm.client, cm.img, cm.log)
	return co, nil
}

// execTx executes a transaction in a container
// - Get the blockcode from the blockchain
// - Create and start a container.
// - Copy the blockcode into the container.
// - Build the blockcode within the container.
// - Run the blockcode in the container.
// - Create connection with the blockcode stub.
// - Attach an handler to process incoming messages from the blockcode.
// - Construct and pass argument to blockcode for execution
func (cm *ContainerManager) execTx(tx *objects.Transaction, output chan []byte, done chan error) {

	bcode, err := cm.blockchain.GetBlockCode(tx.To.String())
	if err != nil {
		done <- fmt.Errorf("failed to get blockcode. %s", err)
		return
	}
	if bcode == nil {
		done <- fmt.Errorf("blockcode not found")
		return
	}
	content := bcode.GetCode()
	if len(content) == 0 {
		done <- fmt.Errorf("blockcode has no content")
		return
	}

	co, err := cm.newContainer()
	if err != nil {
		done <- err
		return
	}
	err = co.create()
	if err != nil {
		done <- err
		return
	}

	err = co.start()
	if err != nil {
		co.destroy()
		done <- err
		return
	}

	err = co.copy(content)
	if err != nil {
		co.destroy()
		done <- err
		return
	}

	switch bcode.Manifest.Lang {
	case blockcode.LangGo:
		co.setBuildLang(newGoBuilder(co.id()))
	default:
		co.destroy()
		done <- fmt.Errorf("unsupported blockcode language")
		return
	}

	buildOk, err := co.build()
	if err != nil {
		co.destroy()
		done <- err
		return
	}
	if !buildOk {
		co.destroy()
		done <- fmt.Errorf("build failed")
		return
	}

	go func() {
		runScript := co.buildConfig.GetRunScript()
		_, err = co.exec(runScript)
		if err != nil {
			done <- err
			return
		}
		done <- nil
	}()

	<-time.After(500 * time.Millisecond) // wait for bcode to be ready

	c, err := gosocketio.Dial(gosocketio.GetUrl("localhost", co.port, false), transport.GetDefaultWebsocketTransport())
	if err != nil {
		co.destroy()
		done <- fmt.Errorf("failed to dial blockcode in container")
		return
	}

	c.On("msg", func(msg *BlockcodeMsg) *Result {
		return cm.handleBlockcodeMsg(msg, co)
	})

	args := Args{
		Func:    tx.InvokeArgs.Func,
		Payload: tx.InvokeArgs.Params,
		Tx: &Tx{
			ID:    tx.GetID(),
			Value: tx.Value.String(),
		},
	}

	res, err := c.Ack("invoke", args, time.Duration(params.MaxTxExecutionTime)*time.Second)
	if err != nil {
		co.destroy()
		_err := fmt.Errorf("failed to run invocation. %s", err)
		if err == gosocketio.ErrorSendTimeout {
			_err = fmt.Errorf("transaction aborted. Took too long")
		}
		done <- _err
		return
	}

	if len(co.getChildren()) == 0 {
		co.destroy()
	}

	output <- []byte(res)
	done <- nil
}

// handleBlockcodeMsg processes block messages
func (cm *ContainerManager) handleBlockcodeMsg(msg *BlockcodeMsg, co *Container) *Result {
	cm.log.Debug("container was to run a tx of its own")
	return nil
}
