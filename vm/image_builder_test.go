package vm

import (
	"fmt"

	"github.com/ellcrys/elld/util/logger"
	docker "github.com/fsouza/go-dockerclient"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ImageBuilder", func() {

	var err error
	var dockerClient *docker.Client
	var builder *ImageBuilder
	var dckFileURL = fmt.Sprintf(dockerFileURL, dockerFileHash)
	var log = logger.NewLogrus()
	log.SetToDebug()

	BeforeEach(func() {
		dockerClient, err = docker.NewClient(dockerEndpoint)
		Expect(err).To(BeNil())
	})

	BeforeEach(func() {
		builder = NewImageBuilder(log, dockerClient, dckFileURL)
	})

	Describe(".getDockerFile", func() {
		It("should fetch docker file from github using a commit hash as version", func() {
			res, err := builder.getDockerFile()
			Expect(err).To(BeNil())
			Expect(res).ToNot(BeEmpty())
		})
	})

	Describe(".buildImage", func() {

		It("should build image from docker file", func() {
			image, err := builder.Build()
			Expect(err).To(BeNil())
			Expect(image).NotTo(BeNil())
			Expect(image.ID).NotTo(BeEmpty())
			builder.getImage()
		})

		Describe(".getImage", func() {
			It("should get image if it exists", func() {
				image, err := builder.getImage()
				Expect(err).To(BeNil())
				Expect(image).NotTo(BeNil())
				Expect(image.ID).ToNot(BeEmpty())
			})
		})
	})

})
