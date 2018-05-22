package vm

import (
	"fmt"

	"github.com/docker/docker/client"
	"github.com/ellcrys/druid/util/logger"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ImageBuilder", func() {

	var err error
	var dockerClient *client.Client
	var log = logger.NewLogrusNoOp()
	var builder *ImageBuilder
	var dckFileURL = fmt.Sprintf(dockerFileURL, dockerFileHash)

	BeforeEach(func() {
		dockerClient, err = client.NewClientWithOpts()
		Expect(err).To(BeNil())
	})

	BeforeEach(func() {
		builder = NewImageBuilder(log, dockerClient, dckFileURL)
	})

	AfterSuite(func() {
		err := builder.destroyImage()
		Expect(err).To(BeNil())
	})

	Describe(".getDockerFile", func() {
		It("should fetch docker file from github using a commit hash as version", func() {
			res, err := builder.getDockerFile()
			Expect(err).To(BeNil())
			Expect(res).ToNot(BeEmpty())
		})
	})

	Describe(".buildImage", func() {

		BeforeEach(func() {
			dockerfile, err := builder.getDockerFile()
			Expect(err).To(BeNil())
			Expect(dockerfile).ToNot(BeEmpty())
		})

		It("should build image from  docker file", func() {
			image, err := builder.Build()
			Expect(err).To(BeNil())
			Expect(image).NotTo(BeNil())
			Expect(image.ID).NotTo(BeNil())
		})

		Describe(".getImage", func() {
			It("should get image if it exists", func() {
				image := builder.getImage()
				Expect(image).NotTo(BeNil())
				Expect(image.ID).ToNot(BeEmpty())
			})
		})
	})

	Describe("BuildContext", func() {

		var buildCtx *BuildContext
		var err error

		BeforeEach(func() {
			dir := "test-hello"
			buildCtx, err = NewBuildContext(dir, "hellofile", "hello world")
			Expect(err).To(BeNil())
			Expect(buildCtx).NotTo(BeNil())
		})

		Describe(".addFile", func() {
			It("should create a file and content", func() {
				err := buildCtx.addFile("hellofile", []byte("hello world"))
				Expect(err).To(BeNil())
			})
		})

		Describe(".Reader", func() {
			It("should create a stream from buildCtx", func() {
				reader, err := buildCtx.Reader()
				Expect(err).To(BeNil())
				Expect(reader).NotTo(BeNil())
			})
		})
	})
})
