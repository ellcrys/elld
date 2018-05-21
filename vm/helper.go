package vm

import (
	"context"
	"fmt"

	"github.com/thoas/go-funk"

	"github.com/docker/docker/client"
	"github.com/franela/goreq"
)

const gitURL = "https://raw.githubusercontent.com/ellcrys/vm-dockerfile"

// dockerAlive checks whether docker server is alive
func dockerAlive() error {

	cli, err := client.NewClientWithOpts()
	if err != nil {
		return err
	}

	_, err = cli.Info(context.Background())
	if err != nil {
		if funk.Contains(err.Error(), "Cannot connect to the Docker") {
			return err
		}
		panic(err)
	}

	return cli.Close()
}

//getDockerFile fetches Dockerfile from github.
func getDockerFile(version string) (*goreq.Response, error) {
	commitURI := fmt.Sprintf("%s/%s/Dockerfile", gitURL, version)

	res, err := goreq.Request{
		Uri: commitURI,
	}.Do()
	if err != nil {
		return nil, err
	}

	if res.Status == "404 Not Found" {
		return nil, fmt.Errorf("%s", "Docker file not found")
	}

	return res, nil
}
