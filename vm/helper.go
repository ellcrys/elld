package vm

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/ellcrys/druid/configdir"
	homedir "github.com/mitchellh/go-homedir"

	"github.com/thoas/go-funk"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/franela/goreq"
)

const gitURL = "https://raw.githubusercontent.com/ellcrys/vm-dockerfile"

// BuildContext for building a docker image
type BuildContext struct {
	Dir string
}

// BuildResponse that respresents a stream from docker build progression
type BuildResponse struct {
	Stream string `json:"stream"`
}

// Aux represents the aux structure in a docker image build response
type Aux struct {
	Image Image `json:"aux"`
}

// Image defines the structure of the final output of the image build
type Image struct {
	ID string `json:"ID"`
}

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

// getDockerFile fetches Dockerfile from github.
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

// buildImage builds an image from a docker file gotten from the getDockerFile func
// - it creates a build context for the docker image build command
// - returns the Image & ID if build is successful
func buildImage(dockerFile *goreq.Response) (*Image, error) {
	ctx := context.Background()

	body, err := dockerFile.Body.ToString()

	buildCtx, err := newBuildCtx(body)
	if err != nil {
		return nil, err
	}

	defer buildCtx.Close()
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}

	source, err := buildCtx.Reader()
	if err != nil {
		return nil, err
	}

	img, err := cli.ImageBuild(ctx, source, types.ImageBuildOptions{
		Tags: []string{"ellcrys-vm"},
	})
	if err != nil {
		return nil, err
	}
	defer img.Body.Close()

	scanner := bufio.NewScanner(img.Body)

	var buildResp BuildResponse
	var aux Aux
	var ID string
	fmt.Print("Building.")
	for scanner.Scan() {
		err := json.Unmarshal(scanner.Bytes(), &buildResp)
		if err != nil {
			return nil, err
		}

		fmt.Print(".")

		if strings.Contains(scanner.Text(), "aux") {
			json.Unmarshal(scanner.Bytes(), &aux)
			ID = strings.Split(aux.Image.ID, ":")[1]
		}
	}
	fmt.Print(".100%\n")
	return &Image{ID}, nil
}

// addFile stores the dockerfile temporarily on the system
func (b *BuildContext) addFile(file string, content []byte) error {
	fp := filepath.Join(b.Dir, filepath.FromSlash(file))
	dirpath := filepath.Dir(fp)
	if dirpath != "." {
		if err := os.MkdirAll(dirpath, 0755); err != nil {
			return err
		}
	}
	return ioutil.WriteFile(fp, content, 0644)
}

// Close deletes the context
func (b *BuildContext) Close() error {
	return os.RemoveAll(b.Dir)
}

// Reader outputs a tar stream of the docker file
func (b *BuildContext) Reader() (io.ReadCloser, error) {
	reader, err := archive.TarWithOptions(b.Dir, &archive.TarOptions{})
	if err != nil {
		return nil, err
	}

	return reader, nil
}

// newBuildCtx creates a build context for the docker image build
func newBuildCtx(content string) (*BuildContext, error) {
	buildContext := new(BuildContext)
	homeDir, _ := homedir.Dir()

	dir := fmt.Sprintf("%s/.ellcrys/vm-build-context", homeDir)

	err := os.Mkdir(dir, 0700)
	if err != nil {
		return nil, err
	}

	cfdir, err := configdir.NewConfigDir(dir)
	if err != nil {
		return nil, err
	}

	buildContext.Dir = cfdir.Path()

	err = buildContext.addFile("Dockerfile", []byte(content))
	if err != nil {
		return nil, err
	}

	return buildContext, nil
}
