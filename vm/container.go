package vm

import (
	"os/user"
)

//Container struct for managing docker containers
type Container struct {
	port        int
	targetPort  int
	execpath    string
	containerID string
}

//NewContainer creates a new docker container for executing smart contracts
func NewContainer(port int, targetPort int, mountPath string) (*Container, error) {
	var execpath string
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}

	if mountPath != "" {
		execpath = mountPath
	} else {
		execpath = usr.HomeDir + TempPath
	}

	//Todo: create container and mount path

	return &Container{
		port:       port,
		targetPort: targetPort,
		execpath:   execpath,
	}, nil
}

//Destroy this container
func (container *Container) Destroy() error {

	return nil
}
