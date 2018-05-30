package vm

import (
	"bytes"
	"context"
	"fmt"
	"time"

	logger "github.com/ellcrys/druid/util/logger"

	"github.com/cenkalti/rpc2"
	"github.com/ellcrys/docker/api/types"
	"github.com/ellcrys/docker/client"
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
	port        int
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

// executes a command in the container.
func (co *Container) exec(command []string) error {

	ctx := context.Background()
	exec, err := co.dockerCli.ContainerExecCreate(ctx, co.id, types.ExecConfig{
		Cmd:          command,
		Tty:          true,
		AttachStderr: true,
		AttachStdin:  false,
		AttachStdout: true,
	})
	if err != nil {
		return fmt.Errorf("failed to create exec %s", err)
	}

	execResp, err := co.dockerCli.ContainerExecAttach(ctx, exec.ID, types.ExecStartCheck{})
	if err != nil {
		return fmt.Errorf("failed to attach to container exec. %s", err)
	}

	defer execResp.Close()

	co.log.Debug("Starting blockcode execution process")
	err = co.dockerCli.ContainerExecStart(ctx, exec.ID, types.ExecStartCheck{Detach: false})
	if err != nil {
		return fmt.Errorf("failed to start exec %s", err)
	}

	for {
		line, _, _ := execResp.Reader.ReadLine()
		if len(line) > 0 {
			co.log.Debug(fmt.Sprintf("Blockcode Execution ==> %s", string(line)))
		}
		break
	}

	co.log.Debug("Blockcode execution successful")
	return nil
}

// setBuildLang takes a concrete implementation of the LangBuilder
func (co *Container) setBuildLang(buildConfig LangBuilder) {
	co.buildConfig = buildConfig
}

// builds a block code
func (co *Container) build() error {
	err := co.buildConfig.Build()
	if err != nil {
		return err
	}
	return nil
}

// copy block code content into container
// - creates new instance of BuildContext
// - build context creates temporary dir to store block code content
// - build context creates a TAR reader stream for docker cli to copy content into container
// - docker cli copies TAR stream into container
func (co *Container) copy(id string, content []byte) error {

	buf := bytes.NewBuffer(content)

	err := co.dockerCli.CopyToContainer(context.Background(), co.id, "/go/src/contract/"+id, buf, types.CopyToContainerOptions{
		AllowOverwriteDirWithFile: false,
	})

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
