package vm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/cenkalti/hub"

	"github.com/cenkalti/rpc2"
	"github.com/docker/docker/api/types"
	dockerContainer "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	docker "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/phayes/freeport"
)

//Container struct for managing docker containers
type Container struct {
	port       int
	execpath   string
	ID         string
	client     *docker.Client
	service    *rpc2.Client
	contractID string
	eventHub   *hub.Hub
}

//ContractManifest defines the project metadata
type ContractManifest struct {
	Name     string `json:"name"`
	Language string `json:"language"`
	Version  string `json:"version"`
	Port     int    `json:"port"`
}

const imgTag = "ellcrys-contract"
const registry = "localhost:5000" //ellcrys image registry

//NewContainer creates a new docker container for executing smart contracts
func NewContainer(contractID string) (*Container, error) {

	ctx := context.Background()
	var execpath string
	usrdir, err := homedir.Dir()
	if err != nil {
		return nil, err
	}

	//assign executable path
	execpath = fmt.Sprintf("%s%s", usrdir+TempPath, contractID)

	//Check if docker is installed
	hasDocker := HasDocker()
	if !hasDocker {
		return nil, errors.New("docker is not installed")
	}

	vmLog.Info("Initializing contract execution container")

	//ctx := context.Background()
	//new docker client
	client, err := docker.NewEnvClient()
	if err != nil {
		return nil, err
	}

	//mount executable path to container contracts path
	contractsDir := fmt.Sprintf("/contracts/%s", contractID)

	//Read manifest of contract
	manifest, err := ioutil.ReadFile(fmt.Sprintf("%s/manifest.json", execpath))

	if err != nil {
		return nil, fmt.Errorf("Cannot Read Manifest :%s", err)
	}

	var contractManifest *ContractManifest

	json.Unmarshal(manifest, &contractManifest)

	//find available port on OS
	availablePort, err := freeport.GetFreePort()
	if err != nil {
		return nil, err
	}

	//get port from manifest
	port := contractManifest.Port

	nPort, err := nat.NewPort("tcp", strconv.Itoa(port))
	if err != nil {
		return nil, err
	}

	//Setup container config
	config := dockerContainer.Config{
		Image: "ellcrys-contract:latest",
		ExposedPorts: nat.PortSet{
			nPort: struct{}{},
		},
	}

	if contractManifest.Language == "" {
		return nil, fmt.Errorf("Language undefined :%s", err)
	}

	switch contractManifest.Language {
	case "ts", "typescript":
		config.Cmd = []string{"npm", "start", "--prefix", "." + contractsDir}
	case "go", "golang":
		config.Cmd = []string{"go", "run", "." + contractsDir + "/main.go"}
	}

	hostConfig := dockerContainer.HostConfig{
		Binds: []string{
			"/var/run/docker.sock:/var/run/docker.sock",
		},
		PortBindings: nat.PortMap{
			nPort: []nat.PortBinding{{
				HostIP:   "0.0.0.0",
				HostPort: strconv.Itoa(availablePort),
			}},
		},
		PublishAllPorts: true,
		Mounts: []mount.Mount{{
			Type:     "bind",
			Source:   execpath,
			Target:   contractsDir,
			ReadOnly: false,
		}},
	}

	//Create the container
	container, err := client.ContainerCreate(ctx, &config, &hostConfig, &network.NetworkingConfig{}, "")
	if err != nil {
		return nil, err
	}

	return &Container{
		port:       availablePort,
		execpath:   execpath,
		ID:         container.ID,
		client:     client,
		contractID: contractID,
	}, nil
}

//Destroy this container
func (container *Container) Destroy() (string, error) {
	ctx := context.Background()
	err := container.client.ContainerStop(ctx, container.ID, nil)
	if err != nil {
		return "", err
	}

	err = container.client.ContainerRemove(ctx, container.ID, types.ContainerRemoveOptions{})
	if err != nil {
		return "", err
	}

	return container.ID, nil
}
