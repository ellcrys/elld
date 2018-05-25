package vm

import (
	"context"
	"fmt"
	"sync"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	logger "github.com/ellcrys/druid/util/logger"
	"github.com/ellcrys/druid/wire"
)

// ContainerManager defines the module that manages containerization of Block codes and their execution
type ContainerManager struct {
	containers    map[string]*Container
	containerLock *sync.Mutex
	log           logger.Logger
	client        *client.Client
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

	cb, err := cm.client.ContainerCreate(context.Background(), &container.Config{
		Image: image.ID,
		Labels: map[string]string{
			"maintainer":    "Ellcrys",
			"image-version": dockerFileHash,
		},
	}, &container.HostConfig{}, &network.NetworkingConfig{}, ID)
	if err != nil {
		return nil, err
	}

	co := new(Container)
	co.dockerCli = cm.client
	co.id = cb.ID
	co.log = cm.log

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
func (cm *ContainerManager) Run(tx *wire.Transaction, blockcodes []BlockCode, txOutput chan []byte, done chan error) {

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
