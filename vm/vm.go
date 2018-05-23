package vm

import (
	"fmt"
	"os"

	"github.com/docker/docker/client"
	"github.com/ellcrys/druid/util"
	funk "github.com/thoas/go-funk"

	logger "github.com/ellcrys/druid/util/logger"
)

const (
	dockerFileHash = "c0879257e8136bf13b4fceb5651f751b806782a7"
	dockerFileURL  = "https://raw.githubusercontent.com/ellcrys/vm-dockerfile/%s/Dockerfile"
)

// MountDir where block codes are stored
var MountDir = "mountdir"

// BlockCode represents a block code in a blockchain
type BlockCode struct {
	Lang            string
	LangVersion     string
	Content         []byte
	PublicFunctions []string
}

// VM specializes in executing transactions against a contracts
type VM struct {
	containers             map[string]*Container
	log                    logger.Logger
	containerMountDir      string
	InvokeResponseListener func(interface{})
	dockerClient           *client.Client
}

// New creates a new instance of VM
func New(log logger.Logger, containerMountDir string) *VM {
	vm := new(VM)
	vm.log = log
	vm.containerMountDir = containerMountDir
	vm.containers = map[string]*Container{}
	return vm
}

// Init sets up the environment for execution of contracts.
// - Create and connect docker client
// - Check if container mount directory exists, otherwise create it
// - Check if docker image exists, if not, fetch and build the image
func (vm *VM) Init() error {

	var err error

	vm.dockerClient, err = client.NewClientWithOpts()
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

	imgBuilder := NewImageBuilder(vm.log, vm.dockerClient, fmt.Sprintf(dockerFileURL, dockerFileHash))
	_, err = imgBuilder.Build()
	if err != nil {
		return err
	}

	return nil
}
