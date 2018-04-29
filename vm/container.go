package vm

import (
	"os/exec"
	"os/user"
)

//Container struct for managing docker containers
type Container struct {
	port       int
	targetPort int
	execpath   string
}

//NewContainer creates a new docker container for executing smart contracts
func NewContainer(port int, targetPort int, mountPath string) (*Container, error) {
	var execpath string
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}

	if mountPath != "" {
		//do mount path
		execpath = mountPath
	} else {
		execpath = usr.HomeDir + "/.ellcrys/tmp/"
	}

	return &Container{
		port:       port,
		targetPort: targetPort,
		execpath:   execpath,
	}, nil
}

//Run commands in a container
func (container *Container) Run(command *exec.Cmd) error {
	command.Args = append([]string{"container args"}, command.Args...)
	return nil
}

//Destroy this container
func (container *Container) Destroy() error {

	return nil
}
