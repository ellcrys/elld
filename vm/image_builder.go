package vm

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/ellcrys/druid/util/logger"
	"github.com/franela/goreq"
	docker "github.com/fsouza/go-dockerclient"
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
	client        *docker.Client
}

// NewImageBuilder creates an instance of ImageBuilder
func NewImageBuilder(log logger.Logger, dockerClient *docker.Client, dockerFileURL string) *ImageBuilder {
	ib := new(ImageBuilder)
	ib.log = log
	ib.client = dockerClient
	ib.dockerFileURL = dockerFileURL
	return ib
}

// getDockerFile fetches Dockerfile from github.
func (ib *ImageBuilder) getDockerFile() (string, error) {

	goreq.SetConnectTimeout(10 * time.Second)
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
// - Create an in-memory tar object and add the docker file to it
// - Create the build options
// - Build image in a separate go routine
// - Read the output
// - When build completes, check if image was created
func (ib *ImageBuilder) Build() (*Image, error) {

	var err error

	image, _ := ib.getImage()
	if image != nil {
		return image, nil
	}

	dockerFileContent, err := ib.getDockerFile()
	if err != nil {
		return nil, err
	}

	inpBuf := bytes.NewBuffer(nil)
	tr := tar.NewWriter(inpBuf)
	t := time.Now()
	dockerfileSize := int64(len([]byte(dockerFileContent)))
	tr.WriteHeader(&tar.Header{Name: "Dockerfile", Size: dockerfileSize, ModTime: t, AccessTime: t, ChangeTime: t})
	tr.Write([]byte(dockerFileContent))
	tr.Close()

	r, w := io.Pipe()
	imgOpts := docker.BuildImageOptions{
		Name:           dockerFileHash,
		RmTmpContainer: true,
		Labels: map[string]string{
			"maintainer": "ellcrys",
			"version":    dockerFileHash,
		},
		InputStream:  inpBuf,
		OutputStream: w,
	}

	errCh := make(chan error, 1)
	go func() {
		defer w.Close()
		defer close(errCh)
		if err := ib.client.BuildImage(imgOpts); err != nil {
			errCh <- err
		}
	}()

	go func() {
		for {
			buf := make([]byte, 64)
			_, err := r.Read(buf)
			if err != nil {
				break
			}

			if logStr := string(buf); strings.TrimSpace(logStr) != "" {
				ib.log.Debug(fmt.Sprintf("[Building Image] -> %s", strings.Join(strings.Fields(logStr), " ")))
			}
		}
	}()

	if buildErr := <-errCh; err != nil {
		return nil, fmt.Errorf("build failed: %s", buildErr)
	}

	img, _ := ib.getImage()
	if img == nil {
		return nil, fmt.Errorf("failed to create image")
	}

	return img, nil
}

// destroyImage removes a docker image
func (ib *ImageBuilder) destroyImage() error {

	image, err := ib.getImage()
	if err != nil {
		return err
	}

	if image == nil {

		return fmt.Errorf("image not found")
	}

	err = ib.client.RemoveImageExtended(image.ID, docker.RemoveImageOptions{Force: true})
	if err != nil {
		return err
	}

	return nil
}

func (ib *ImageBuilder) getImage() (*Image, error) {

	image, err := ib.client.InspectImage(dockerFileHash)
	if err != nil {
		return nil, err
	}

	if image == nil {
		return nil, nil
	}

	if labels := image.Config.Labels; labels["version"] != dockerFileHash || labels["maintainer"] != "ellcrys" {
		return nil, fmt.Errorf("similar docker image found but not maintained by ellcrys")
	}

	return &Image{
		ID: image.ID,
	}, nil
}
