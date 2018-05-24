package vm

import (
	"bufio"
	"context"
	"time"

	logger "github.com/ellcrys/druid/util/logger"

	"github.com/cenkalti/rpc2"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

var containerStopTimeout = time.Second * 2

// Container defines the container that runs a block code.
type Container struct {
	id          string // id of the container
	children    []*Container
	client      *rpc2.Client
	parent      *Container
	dockerCli   *client.Client
	buildConfig LangBuilder
	log         logger.Logger
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

// executes a command in the container. It will block and channel std output to output.
// If an error occurs, done will be sent an error, otherwise, nil.
func (co *Container) exec(command []string, output chan string, done chan error) {

	ctx := context.Background()
	exec, err := co.dockerCli.ContainerExecCreate(ctx, co.id, types.ExecConfig{
		Cmd:          command,
		Tty:          true,
		AttachStderr: true,
		AttachStdin:  false,
		AttachStdout: true,
	})
	if err != nil {
		done <- err
		return
	}

	execResp, _ := co.dockerCli.ContainerExecAttach(ctx, exec.ID, types.ExecStartCheck{})

	defer execResp.Close()

	_ = co.dockerCli.ContainerExecStart(ctx, exec.ID, types.ExecStartCheck{Detach: false})

	scanner := bufio.NewScanner(execResp.Reader)
	for scanner.Scan() {
		out := scanner.Text()
		if out != "" {
			output <- out
		}
	}

	done <- nil

	return
}

// buildLang takes a concrete implementation of the LangBuilder

func (co *Container) setBuildLang(buildConfig LangBuilder) {
	co.buildConfig = buildConfig
}

// builds a block code
func (co *Container) build() error {
	err := co.buildConfig.Build(co.id)
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

	err := co.dockerCli.ContainerStop(context.Background(), co.id, &containerStopTimeout)
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
