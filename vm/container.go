package vm

import (
	"context"
	"encoding/json"

	"github.com/cenkalti/rpc2"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// Container defines the container that runs a block code.
type Container struct {
	id        string // id of the container
	children  []*Container
	client    *rpc2.Client
	parent    *Container
	dockerCli *client.Client
}

// response defines response from a blockcode execution
type response struct {
	Status string `json:"status"`
	Code   int    `json:"code"`
	Data   []byte `json:"data"`
}

// ExecRequest defines the execution request to the blockcode stub
type ExecRequest struct {
	Function string      `json:"Function"`
	Data     interface{} `json:"Data"`
}

// starts a container
func (co *Container) start() error {
	err := co.dockerCli.ContainerStart(context.Background(), co.id, types.ContainerStartOptions{})
	if err != nil {
		return err
	}
	return nil
}

// executes a block code in the container
func (co *Container) exec(execBlock *ExecRequest, output chan []byte, done chan struct{}) error {

	co.client.Handle("response", func(client *rpc2.Client, data *response, _ *struct{}) error {

		b, err := json.Marshal(&data)
		if err != nil {
			close(done)
			return err
		}
		output <- b
		close(done)
		return nil
	})

	go co.client.Run()

	err := co.client.Call("invoke", execBlock, nil)
	if err != nil {
		return err
	}

	return nil
}

// buildLang takes a concrete implementation of the LangBuilder
// - and builds a block code accordingly
func (co *Container) buildLang(buildConfig LangBuilder) error {
	err := buildConfig.Build(co.id)
	if err != nil {
		return err
	}

	return nil
}

// adds a child container to the list of children
func (co *Container) addChild(child *Container) {
	child.parent = co
	co.children = append(co.children, child)
}

// stop a started container
func (co *Container) stop() error {
	err := co.dockerCli.ContainerStop(context.Background(), co.id, nil)
	if err != nil {
		return err
	}
	return nil
}

// Destroy a container
func (co *Container) destroy() error {
	err := co.dockerCli.ContainerRemove(context.Background(), co.id, types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	})
	if err != nil {
		return err
	}

	return nil
}
