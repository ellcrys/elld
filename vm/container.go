package vm

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/phayes/freeport"

	"github.com/ellcrys/elld/util"
	logger "github.com/ellcrys/elld/util/logger"

	"github.com/cenkalti/rpc2"
	docker "github.com/fsouza/go-dockerclient"
)

var containerStopTimeout uint = 2

// Container defines the container that runs a block code.
type Container struct {
	mtx         *sync.RWMutex
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
	return fmt.Sprintf("/go/src/github.com/contracts/%s", id)
}

// NewContainer creates a new container object
func NewContainer(dockerCli *docker.Client, img *Image, log logger.Logger) *Container {
	c := new(Container)
	c.dockerCli = dockerCli
	c.log = log
	c.img = img
	c.mtx = &sync.RWMutex{}
	return c
}

func (co *Container) create() error {

	var err error

	if co._container != nil {
		return fmt.Errorf("container already created")
	}

	port, err := freeport.GetFreePort()
	if err != nil {
		return fmt.Errorf("failed to get free port. %s", err)
	}

	co._container, err = co.dockerCli.CreateContainer(docker.CreateContainerOptions{
		Name: "bcode_" + util.RandString(16),
		Config: &docker.Config{
			Image:        co.img.ID,
			AttachStderr: true,
			AttachStdout: true,
		},
		HostConfig: &docker.HostConfig{
			PortBindings: map[docker.Port][]docker.PortBinding{
				"9900/tcp": {{HostIP: "127.0.0.1", HostPort: strconv.Itoa(port)}},
			},
		},
	})

	co.port = port

	return err
}

func (co *Container) stop() error {

	if co._container == nil {
		return nil
	}

	ci, err := co.dockerCli.InspectContainer(co.id())
	if err != nil {
		return err
	}

	if !ci.State.Running {
		return nil
	}

	return co.dockerCli.StopContainer(co.id(), containerStopTimeout)
}

// starts a container
func (co *Container) start() error {

	if co._container == nil {
		return fmt.Errorf("container not initialized")
	}

	err := co.dockerCli.StartContainer(co.id(), &docker.HostConfig{})

	return err
}

func (co *Container) isRunning() bool {

	if co._container == nil {
		return false
	}

	ci, _ := co.dockerCli.InspectContainer(co.id())
	return ci.State.Running
}

// executes a command in the container.
func (co *Container) exec(command []string) (int, error) {

	exec, err := co.dockerCli.CreateExec(docker.CreateExecOptions{
		Container:    co.id(),
		AttachStderr: true,
		AttachStdin:  false,
		AttachStdout: true,
		Tty:          true,
		Cmd:          command,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to create exec %s", err)
	}

	co.log.Debug(fmt.Sprintf("[Container#%s]: Executing command -> {%s}", co.id()[0:10], strings.Join(command, " ")))

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

		execI, err := co.dockerCli.InspectExec(exec.ID)
		if err != nil {
			return 0, fmt.Errorf("failed to get exec status. %s", err)
		}

		line, err := outBuf.ReadString('\n')
		if err != nil {
			if !execI.Running {
				break
			}
		}

		logStr := strings.Trim(line, "\n\t")
		if logStr != "" {
			co.log.Debug(fmt.Sprintf("[Container#%s]: %s", co.id()[0:10], logStr))
		}
	}

	co.log.Debug(fmt.Sprintf("[Container#%s]: Finished command -> {%s}", co.id()[0:10], strings.Join(command, " ")))
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

// id returns the container's ID
func (co *Container) id() string {
	if co._container == nil {
		panic("container not initialized")
	}
	return co._container.ID
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

	path := makeCopyPath(co.id())
	code, err := co.exec([]string{"bash", "-c", "mkdir -p " + path})
	if err != nil {
		return err
	}

	if code != 0 {
		return fmt.Errorf("failed to create path in container")
	}

	buf := bytes.NewBuffer(content)
	err = co.dockerCli.UploadToContainer(co.id(), docker.UploadToContainerOptions{
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

// get child containers
func (co *Container) getChildren() []*Container {
	co.mtx.RLock()
	defer co.mtx.RUnlock()
	return co.children
}

// Destroy a container
func (co *Container) destroy() error {

	if co._container == nil {
		return nil
	}

	err := co.dockerCli.RemoveContainer(docker.RemoveContainerOptions{
		ID:    co.id(),
		Force: true,
	})

	co._container = nil

	return err
}
