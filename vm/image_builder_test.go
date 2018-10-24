package vm

import (
	"fmt"
	"testing"

	"github.com/ellcrys/elld/util/logger"
	docker "github.com/fsouza/go-dockerclient"
	. "github.com/ncodes/goblin"
	. "github.com/onsi/gomega"
)

func TestImageBuilder(t *testing.T) {
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("ImageBuilder", func() {

		var err error
		var dockerClient *docker.Client
		var builder *ImageBuilder
		var dckFileURL = fmt.Sprintf(dockerFileURL, dockerFileHash)
		var log = logger.NewLogrus()
		log.SetToDebug()

		g.BeforeEach(func() {
			dockerClient, err = docker.NewClient(dockerEndpoint)
			Expect(err).To(BeNil())
		})

		g.BeforeEach(func() {
			builder = NewImageBuilder(log, dockerClient, dckFileURL)
		})

		g.Describe(".getDockerFile", func() {
			g.It("should fetch docker file from github using a commit hash as version", func() {
				res, err := builder.getDockerFile()
				Expect(err).To(BeNil())
				Expect(res).ToNot(BeEmpty())
			})
		})

		g.Describe(".buildImage", func() {

			g.It("should build image from docker file", func() {
				image, err := builder.Build()
				Expect(err).To(BeNil())
				Expect(image).NotTo(BeNil())
				Expect(image.ID).NotTo(BeEmpty())
				builder.getImage()
			})

			g.Describe(".getImage", func() {
				g.It("should get image if it exists", func() {
					image, err := builder.getImage()
					Expect(err).To(BeNil())
					Expect(image).NotTo(BeNil())
					Expect(image.ID).ToNot(BeEmpty())
				})
			})
		})
	})
}
