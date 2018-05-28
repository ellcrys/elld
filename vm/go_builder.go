package vm

import (
	"bufio"
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/docker/docker/api/types"
	logger "github.com/ellcrys/druid/util/logger"
)

type goBuilder struct {
	id        string
	container *Container
	log       logger.Logger
}

// create new instance of goBuilder
func newGoBuilder(id string, container *Container, log logger.Logger) *goBuilder {

	return &goBuilder{
		id:        id,
		container: container,
		log:       log,
	}
}

// get the run script that executes a blockcode
func (gb *goBuilder) GetRunScript() []string {
	script := fmt.Sprintf("./container/bin/%s", gb.id)
	return []string{script}
}

// build a block code
func (gb *goBuilder) Build(mtx *sync.Mutex) ([]byte, error) {
	ctx := context.Background()
	archive := fmt.Sprintf("./archive/%s", gb.id)
	execoutput := fmt.Sprintf("~/bin/%s", gb.id)
	buildCmd := []string{"bash", "-c", "unzip", archive, "-d", "./container", "&&", "mkdir ./container/bin", "&&", "go", "build", "./container", "-o", execoutput}

	mtx.Lock()
	exec, err := gb.container.dockerCli.ContainerExecCreate(ctx, gb.id, types.ExecConfig{
		Cmd:          buildCmd,
		Tty:          true,
		AttachStderr: true,
		AttachStdin:  false,
		AttachStdout: true,
	})
	if err != nil {
		mtx.Unlock()
		return nil, err
	}

	execResp, _ := gb.container.dockerCli.ContainerExecAttach(ctx, exec.ID, types.ExecStartCheck{})

	err = gb.container.dockerCli.ContainerExecStart(ctx, exec.ID, types.ExecStartCheck{Detach: false})
	if err != nil {
		mtx.Unlock()
		return nil, err
	}

	outputArr := []string{}
	scanner := bufio.NewScanner(execResp.Reader)
	for scanner.Scan() {
		out := scanner.Text()
		if out != "" {
			gb.log.Debug("Blockcode Build", "output", out)
			outputArr = append(outputArr, out)
		}
	}
	mtx.Unlock()
	output := []byte(strings.Join(outputArr, "\r\n"))
	return output, nil
}
