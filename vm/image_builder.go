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

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/ellcrys/druid/util/logger"
	"github.com/franela/goreq"
	funk "github.com/thoas/go-funk"
)

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

// ImageBuilder builds a struct
type ImageBuilder struct {
	log           logger.Logger
	dockerFileURL string
	client        *client.Client
}

// NewImageBuilder creates an instance of ImageBuilder
func NewImageBuilder(log logger.Logger, dockerClient *client.Client, dockerFileURL string) *ImageBuilder {
	ib := new(ImageBuilder)
	ib.log = log
	ib.client = dockerClient
	ib.dockerFileURL = dockerFileURL
	return ib
}

// getDockerFile fetches Dockerfile from github.
func (ib *ImageBuilder) getDockerFile() (string, error) {

	res, err := goreq.Request{Uri: ib.dockerFileURL}.Do()
	if err != nil {
		return "", err
	}

	switch res.StatusCode {
	case 200:
		body, err := res.Body.ToString()
		if err != nil {
			return "", err
		}
		return body, nil
	case 404:
		return "", fmt.Errorf("docker file not found")
	default:
		return "", fmt.Errorf("problem fetching docker file")
	}
}

// Build builds an image from a docker file gotten from the getDockerFile func
// - Checks if image already exists, else
// - Gets the docker file
// - it creates a build context for the docker image build command
// - get image if it exists
// - builds an image if it doesn't already exists
// - returns the Image & ID if build is successful
func (ib *ImageBuilder) Build() (*Image, error) {

	image := ib.getImage()
	if image != nil {
		return image, nil
	}

	dockerFileContent, err := ib.getDockerFile()
	if err != nil {
		return nil, err
	}

	dir := "vm-build-context"
	buildCtx, err := NewBuildContext(dir, "Dockerfile", dockerFileContent)
	if err != nil {
		return nil, err
	}

	defer buildCtx.Close()

	source, err := buildCtx.Reader()
	if err != nil {
		return nil, err
	}

	defer source.Close()

	ctx := context.Background()
	img, err := ib.client.ImageBuild(ctx, source, types.ImageBuildOptions{
		Tags: []string{dockerFileHash},
		Labels: map[string]string{
			"maintainer": "ellcrys",
			"version":    dockerFileHash,
		},
	})
	if err != nil {
		return nil, err
	}
	defer img.Body.Close()

	scanner := bufio.NewScanner(img.Body)

	var buildResp BuildResponse
	var aux Aux
	var ID string

	for scanner.Scan() {
		err := json.Unmarshal(scanner.Bytes(), &buildResp)
		if err != nil {
			return nil, err
		}
		replacer := strings.NewReplacer("\n", "")
		val := replacer.Replace(buildResp.Stream)
		if val != "" {
			ib.log.Debug("Image Build", "Output", val)
		}

		if strings.Contains(scanner.Text(), "aux") {
			json.Unmarshal(scanner.Bytes(), &aux)
			ID = strings.Split(aux.Image.ID, ":")[1]
		}
	}
	return &Image{ID}, nil
}

// destroyImage removes a docker image
func (ib *ImageBuilder) destroyImage() error {

	image := ib.getImage()
	ctx := context.Background()

	_, err := ib.client.ImageRemove(ctx, image.ID, types.ImageRemoveOptions{Force: true})
	if err != nil {
		return err
	}

	return nil
}

func (ib *ImageBuilder) getImage() *Image {
	ctx := context.Background()
	summaries, _ := ib.client.ImageList(ctx, types.ImageListOptions{})

	// check images if version already exist
	image := funk.Find(summaries, func(x types.ImageSummary) bool {
		if x.Labels["version"] == dockerFileHash && x.Labels["maintainer"] == "ellcrys" {
			return true
		}
		return false
	})

	if image == nil {
		return nil
	}

	return &Image{
		ID: image.(types.ImageSummary).ID,
	}
}

// NewBuildContext creates a build context for the docker image build
func NewBuildContext(dir string, name string, content string) (*BuildContext, error) {
	buildContext := new(BuildContext)

	tempdir, err := ioutil.TempDir("", dir)
	if err != nil {
		return nil, err
	}
	buildContext.Dir = tempdir

	err = buildContext.addFile(name, []byte(content))
	if err != nil {
		return nil, err
	}

	return buildContext, nil
}

// addFile stores the dockerfile temporarily on the system
func (b *BuildContext) addFile(file string, content []byte) error {
	fp := filepath.Join(b.Dir, filepath.FromSlash(file))
	dirpath := filepath.Dir(fp)
	if dirpath != "." {
		if err := os.MkdirAll(dirpath, 0700); err != nil {
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
