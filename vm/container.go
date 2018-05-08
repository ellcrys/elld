package vm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/ellcrys/rpc2"
	docker "github.com/fsouza/go-dockerclient"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/phayes/freeport"
)

//Container struct for managing docker containers
type Container struct {
	port     int
	execpath string
	ID       string
	client   *docker.Client
	service  *rpc2.Client
}

//ContractManifest defines the project metadata
type ContractManifest struct {
	Name     string `json:"name"`
	Language string `json:"language"`
	Version  string `json:"version"`
}

const imgTag = "ellcrys-contract"
const registry = "localhost:5000" //ellcrys image registry

//NewContainer creates a new docker container for executing smart contracts
func NewContainer(contractID string) (*Container, error) {
	var execpath string
	usrdir, err := homedir.Dir()
	if err != nil {
		return nil, err
	}

	//assign executable path
	execpath = fmt.Sprintf("%s/%s", usrdir+TempPath, contractID)

	//Check if docker is installed
	// hasDocker := HasDocker()
	// if !hasDocker {
	// 	vmLog.Error("Please install docker")
	// 	return nil, errors.New("Please install docker")
	// }

	vmLog.Debug("Initializing contract execution container")

	ctx := context.Background()
	//new docker client
	//endpoint := "unix:///var/run/docker.sock"
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return nil, err
	}

	//Setup container config
	config := docker.Config{
		Image: "ellcrys-contract",
	}

	//find available port on OS
	availablePort, err := freeport.GetFreePort()
	if err != nil {
		return nil, err
	}

	//bind port to container config
	ports := map[docker.Port][]docker.PortBinding{}
	ports["4000/tcp"] = []docker.PortBinding{{
		HostIP:   "0.0.0.0",
		HostPort: strconv.Itoa(availablePort),
	}}

	//mount executable path to container contracts path
	contractsDir := fmt.Sprintf("/contracts/%s", contractID)
	mounts := []docker.HostMount{{
		Target: contractsDir,
		Source: execpath,
		Type:   "bind",
	}}

	config.Mounts = []docker.Mount{{
		Name:        "project-path",
		Source:      execpath,
		Destination: contractsDir,
		RW:          true,
		Mode:        "",
		Driver:      "bind",
	}}

	//Read manifest of contract
	manifest, err := ioutil.ReadFile(fmt.Sprintf("%s/manifest.json", execpath))

	if err != nil {
		return nil, fmt.Errorf("Cannot Read Manifest :%s", err)
	}

	var contractManifest *ContractManifest

	json.Unmarshal(manifest, &contractManifest)

	if contractManifest.Language == "" {
		return nil, fmt.Errorf("Language undefined :%s", err)
	}

	switch contractManifest.Language {
	case "ts", "typescript":
		config.Cmd = []string{"npm", "start", "--prefix", "." + contractsDir}
	case "go", "golang":
		config.Cmd = []string{"go", "run", "." + contractsDir + "/main.go"}
	}

	//Create the container
	container, err := client.CreateContainer(docker.CreateContainerOptions{
		Config: &config,
		HostConfig: &docker.HostConfig{
			PortBindings: ports,
			Mounts:       mounts,
		},
		Context: ctx,
	})

	if err != nil {
		log.Fatal(err)
	}

	//Start the container
	err = client.StartContainer(container.ID, &docker.HostConfig{
		PortBindings: ports,
		Mounts:       mounts,
	})
	if err != nil {
		log.Fatal(err)
	}

	// show log
	var stdoutBuffer bytes.Buffer

	opts := docker.LogsOptions{
		Container:    container.ID,
		OutputStream: &stdoutBuffer,
		Follow:       true,
		Stdout:       true,
	}

	exit := make(chan bool)
	go func() {
		client.Logs(opts)
		close(exit)
	}()
	stdoutCh := readerToChan(&stdoutBuffer, exit)

	ret := make(chan *Container)

	// print log
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
	loop:
		for value := range <-stdoutCh {
			select {
			case <-stdoutCh:
				if value == 2 {
					wg.Done()
					ret <- &Container{
						port:     availablePort,
						execpath: execpath,
						ID:       container.ID,
						client:   client,
					}
					vmLog.Debug(fmt.Sprintf("vm:Container => %s", "done"))
					break loop
				}
			}
		}

	}()
	wg.Wait()

	return <-ret, nil
}

//Destroy this container
func (container *Container) Destroy() (string, error) {

	err := container.client.StopContainer(container.ID, 1000)
	if err != nil {
		return "", err
	}
	container.client.RemoveContainer(docker.RemoveContainerOptions{
		ID: container.ID,
	})
	if err != nil {
		return "", err
	}

	return container.ID, nil
}

func readerToChan(reader *bytes.Buffer, exit <-chan bool) <-chan string {
	c := make(chan string)
	go func() {

		for {
			select {
			case <-exit:
				close(c)
				return
			default:
				line, err := reader.ReadString('\n')

				if err != nil && err != io.EOF {
					close(c)
					return
				}

				line = strings.TrimSpace(line)
				if line != "" {
					c <- line
				}
			}
		}
	}()

	return c
}
