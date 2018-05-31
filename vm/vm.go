package vm

import (
	"fmt"
	"os"

	"github.com/ellcrys/druid/util"
	logger "github.com/ellcrys/druid/util/logger"
	docker "github.com/fsouza/go-dockerclient"
	funk "github.com/thoas/go-funk"
)

const (
	dockerFileHash         = "2a7262215a616106b644a489e6e1da1d52834853"
	dockerFileURL          = "https://raw.githubusercontent.com/ellcrys/vm-dockerfile/%s/Dockerfile"
	targetDockerAPIVersion = "1.37"
	dockerEndpoint         = "unix:///var/run/docker.sock"
)

// MountDir where block codes are stored
var MountDir = "mountdir"

// VM specializes in executing transactions against a contracts
type VM struct {
	log                    logger.Logger
	containerMountDir      string
	InvokeResponseListener func(interface{})
	dockerCli              *docker.Client
}

// New creates a new instance of VM
func New(log logger.Logger, containerMountDir string) *VM {
	vm := new(VM)
	vm.log = log
	vm.containerMountDir = containerMountDir
	return vm
}

// Init sets up the environment for execution of contracts.
// - Create and connect docker client
// - Check if container mount directory exists, otherwise create it
// - Check if docker image exists, if not, fetch and build the image
func (vm *VM) Init() error {

	var err error

	vm.dockerCli, err = docker.NewClient(dockerEndpoint)
	if err != nil {
		if funk.Contains(err.Error(), "Cannot connect to the Docker") {
			return err
		}
		return err
	}

	if !util.IsPathOk(vm.containerMountDir) {
		if err := os.MkdirAll(vm.containerMountDir, 0700); err != nil {
			return fmt.Errorf("failed to create container mount directory. %s", err)
		}
	}

	imgBuilder := NewImageBuilder(vm.log, vm.dockerCli, fmt.Sprintf(dockerFileURL, dockerFileHash))
	_, err = imgBuilder.Build()
	if err != nil {
		return err
	}

	return nil
}
