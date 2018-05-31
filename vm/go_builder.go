package vm

import (
	"bytes"
	"context"
	"fmt"

	"github.com/kr/pretty"

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
	cmd := []string{"bash", "-c", "/bin/bcode  2>/dev/null"}
	return cmd
}

// build a block code
func (gb *goBuilder) Build() error {

	ctx := context.Background()
	archive := fmt.Sprintf("./src/contract/%s", gb.id)
	execCmd := "cd " + archive + "&& go build -x -o /bin/bcode"
	buildCmd := []string{"bash", "-c", execCmd}

	exec, err := gb.container.dockerCli.ContainerExecCreate(ctx, gb.container.id, types.ExecConfig{
		Cmd:          buildCmd,
		Tty:          true,
		AttachStderr: true,
		AttachStdin:  false,
		AttachStdout: true,
	})
	if err != nil {
		return err
	}

	execResp, err := gb.container.dockerCli.ContainerExecAttach(ctx, exec.ID, types.ExecStartCheck{})
	if err != nil {
		return fmt.Errorf("failed to attach to container exec. %s", err)
	}
	defer execResp.Close()

	for {
		execI, err := gb.container.dockerCli.ContainerExecInspect(context.Background(), exec.ID)
		if err != nil {
			return fmt.Errorf("failed to get exec status. %s", err)
		}
		if !execI.Running {
			if execI.ExitCode != 0 {
				return fmt.Errorf("build failed")
			}
			break
		}
	}

	gb.log.Debug("Starting blockcode build process")

	err = gb.container.dockerCli.ContainerExecStart(ctx, exec.ID, types.ExecStartCheck{Detach: false})
	if err != nil {
		pretty.Println(err)
		return err
	}

	var buf = bytes.NewBuffer(nil)
	buf.ReadFrom(execResp.Reader)

	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			break
		}
		if len(line) > 0 {
			gb.log.Info(fmt.Sprintf("%s", line))
		}
	}

	gb.log.Debug("Blockcode build successful")

	return nil
}
