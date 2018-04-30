package vm

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
)

//Container struct for managing docker containers
type Container struct {
	port        int
	targetPort  int
	execpath    string
	containerID string
}

const imgTag = "ellcrys-contract"
const registry = "localhost:5000" //ellcrys image registry

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

	//Check if docker is installed
	hasDocker := HasDocker()
	if !hasDocker {
		vmLog.Error("Please install docker")
		return nil, errors.New("Please install docker")
	}

	
	vmLog.Infof("Initializing contract execution container")
	//pull the image
	err := pullImg()
	if err != nil {
		return nil, err
	}

	//run the container
	containerID, err := runContainer(port, targetPort, execpath)
	if err != nil {
		return nil, err
	}

	return &Container{
		port:       port,
		targetPort: targetPort,
		execpath:   execpath,
		containerID : containerID
	}, nil
}

//pull container image
func pullImg() error {
	vmLog.Debugf("Pull image %s", imgTag)
	containerPullCmd := exec.Command("docker", "pull", fmt.Sprintf("%s/%s", registry, imgTag))
	var stdout, stderr []byte
	var errStdout, errStderr error
	stdoutIn, _ := containerPullCmd.StdoutPipe()
	stderrIn, _ := containerPullCmd.StderrPipe()

	containerPullCmd.Start()


	//Capture stdout
	go func() {
		stdout, errStdout = Capture(os.Stdout, stdoutIn)
	}()

	//Capture stderr
	go func() {
		stderr, errStderr = Capture(os.Stderr, stderrIn)
	}()

	//Wait for outputs from command
	err := containerCmd.Wait()
	if err != nil {
		return nil, err
	}

	if errStdout != nil || errStderr != nil {
		vmLog.Errorf("failed to pull %s", imgTag)
		return fmt.Errorf("failed to pull %s", imgTag)
	}

	outStr, errStr := string(stdout), string(stderr)
	var containerID string
	if outStr != "" {
		vmLog.Infof("%s", outStr)
	}
	if errStr != "" {
		return errors.New("failed to capture stdout or stderr")
	}	

	return nil
}


func runContainer(port int, targetPort int, execpath string) (containerID string, err error) {
	vmLog.Debugf("Create and run container with image %s", imgTag)
	containerCmd := exec.Command("docker", "run", "-d", "--volume", fmt.Sprintf("%s:%s", execpath, "/contracts"), "-p", fmt.Sprintf("%s:%s", port, targetPort), imgTag)
	var stdout, stderr []byte
	var errStdout, errStderr error
	stdoutIn, _ := containerCmd.StdoutPipe()
	stderrIn, _ := containerCmd.StderrPipe()

	//Run command
	containerCmd.Start()

	//Capture stdout
	go func() {
		stdout, errStdout = Capture(os.Stdout, stdoutIn)
	}()

	//Capture stderr
	go func() {
		stderr, errStderr = Capture(os.Stderr, stderrIn)
	}()

	//Wait for outputs from command
	err := containerCmd.Wait()
	if err != nil {
		return nil, err
	}

	if errStdout != nil || errStderr != nil {
		vmLog.Errorf("failed to capture stdout or stderr\n")
		return nil, errors.New("failed to capture stdout or stderr")
	}

	outStr, errStr := string(stdout), string(stderr)
	var containerID string
	if outStr != "" {
		containerID = outStr
	}
	if errStr != "" {
		return nil, errors.New("failed to capture stdout or stderr")
	}

	return containerID, nil
}

//Destroy this container
func (container *Container) Destroy() error {

	return nil
}
