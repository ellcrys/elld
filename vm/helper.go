package vm

import (
	"fmt"

	"github.com/franela/goreq"
)

const gitURL = "https://raw.githubusercontent.com/ellcrys/vm-dockerfile"

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
