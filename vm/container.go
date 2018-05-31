package vm

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/kr/pretty"

	"github.com/ellcrys/druid/util"
	logger "github.com/ellcrys/druid/util/logger"

	"github.com/cenkalti/rpc2"
	docker "github.com/fsouza/go-dockerclient"
)

var containerStopTimeout uint = 2

// Container defines the container that runs a block code.
type Container struct {
	mtx         *sync.Mutex
	id          string            // id of the container
	children    []*Container      // Child containers created by the blockcode ran in this container. Protected by mtx lock.
	client      *rpc2.Client      // Client connection to the blockcode in this container
	parent      *Container        // Parent container whose blockcode started this container. Protected by mtx loc
	dockerCli   *docker.Client    // Docker client
	buildConfig LangBuilder       // Builder used to build the blockcode
	log         logger.Logger     // Logger
	port        int               // Port to bind to the container
	img         *Image            // The image deployed by this container
	_container  *docker.Container // The underlying docker container
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

// returns the path to copy blockcodes to
func makeCopyPath(id string) string {
	return fmt.Sprintf("/go/src/github.com/ellcrys/contracts/%s", id)
}

// NewContainer creates a new container object
func NewContainer(id string, dockerCli *docker.Client, img *Image, log logger.Logger) *Container {
	c := new(Container)
	c.dockerCli = dockerCli
	c.log = log
	c.id = id
	c.img = img
	c.mtx = &sync.Mutex{}
	return c
}

func (co *Container) create() error {

	if co._container != nil {
		return fmt.Errorf("container already created")
	}

	var err error
	co._container, err = co.dockerCli.CreateContainer(docker.CreateContainerOptions{
		Name: "bcode_" + util.RandString(16),
		Config: &docker.Config{
			Image:        co.img.ID,
			AttachStderr: true,
			AttachStdout: true,
		},
	})

	return err
}

func (co *Container) stop() error {

	if co._container == nil {
		return nil
	}

	ci, err := co.dockerCli.InspectContainer(co._container.ID)
	if err != nil {
		return err
	}

	pretty.Println(ci, "<<")
	if !ci.State.Running {
		return nil
	}

	return co.dockerCli.StopContainer(co._container.ID, containerStopTimeout)
}

// starts a container
func (co *Container) start() error {

	if co._container == nil {
		return fmt.Errorf("container not initialized")
	}

	err := co.dockerCli.StartContainer(co._container.ID, &docker.HostConfig{})

	return err
}

func (co *Container) isRunning() bool {

	if co._container == nil {
		return false
	}

	ci, _ := co.dockerCli.InspectContainer(co._container.ID)
	return ci.State.Running
}

// executes a command in the container.
func (co *Container) exec(command []string) (int, error) {

	exec, err := co.dockerCli.CreateExec(docker.CreateExecOptions{
		Container:    co._container.ID,
		AttachStderr: true,
		AttachStdin:  false,
		AttachStdout: true,
		Tty:          true,
		Cmd:          command,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to create exec %s", err)
	}

	co.log.Debug(fmt.Sprintf("[Container#%s]: Executing command -> {%s}", co._container.ID[0:10], strings.Join(command, " ")))

	var outBuf = bytes.NewBuffer(nil)
	_, err = co.dockerCli.StartExecNonBlocking(exec.ID, docker.StartExecOptions{
		Detach:       false,
		ErrorStream:  outBuf,
		OutputStream: outBuf,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to start exec %s", err)
	}

	time.Sleep(100 * time.Millisecond)

	for {
		buf := make([]byte, 64)
		_, err = outBuf.Read(buf)
		if err != nil {
			break
		}

		if logStr := string(buf); strings.TrimSpace(logStr) != "" {
			co.log.Debug(fmt.Sprintf("[Container#%s]: %s", co._container.ID[0:10], strings.Join(strings.Fields(logStr), " ")))
		}
	}

	co.log.Debug(fmt.Sprintf("[Container#%s]: Finished command -> {%s}", co._container.ID[0:10], strings.Join(command, " ")))
	execI, err := co.dockerCli.InspectExec(exec.ID)
	if err != nil {
		return 0, fmt.Errorf("failed to get exec status. %s", err)
	}

	return execI.ExitCode, nil
}

// setBuildLang takes a concrete implementation of the LangBuilder
func (co *Container) setBuildLang(buildConfig LangBuilder) {
	co.buildConfig = buildConfig
}

// builds a blockcode according the build script provided
// by the language builder
func (co *Container) build() (bool, error) {
	buildCmd := co.buildConfig.GetBuildScript()
	statusCode, err := co.exec(buildCmd)
	if err != nil {
		return false, err
	}

	return statusCode == 0, nil
}

// copy block code content into container
// - creates the path in the container we intend to copy to.
// - copy TAR content into container
func (co *Container) copy(content []byte) error {

	path := makeCopyPath(co.id)
	code, err := co.exec([]string{"bash", "-c", "mkdir -p " + path})
	if err != nil {
		return err
	}

	if code != 0 {
		return fmt.Errorf("failed to create path in container")
	}

	buf := bytes.NewBuffer(content)
	err = co.dockerCli.UploadToContainer(co._container.ID, docker.UploadToContainerOptions{
		InputStream:          buf,
		Path:                 path,
		NoOverwriteDirNonDir: true,
	})
	if err != nil {
		return err
	}

	return nil
}

// adds a child container to the list of children
func (co *Container) addChild(child *Container) {
	co.mtx.Lock()
	defer co.mtx.Unlock()
	child.parent = co
	co.children = append(co.children, child)
}

// Destroy a container
func (co *Container) destroy() error {

	if co._container == nil {
		return nil
	}

	err := co.dockerCli.RemoveContainer(docker.RemoveContainerOptions{
		ID:    co._container.ID,
		Force: true,
	})

	co._container = nil

	return err
}
