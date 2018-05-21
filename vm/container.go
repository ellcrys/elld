package vm

import (
	"github.com/cenkalti/hub"

	"github.com/cenkalti/rpc2"
	docker "github.com/docker/docker/client"
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
