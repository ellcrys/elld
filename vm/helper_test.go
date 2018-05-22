package vm

import (
	"reflect"
	"strings"

	"github.com/docker/docker/client"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var expectedBody = `FROM golang:1.10-stretch

# Set work directory
WORKDIR /go

# Expose RPC port
EXPOSE 9900

# Sleep forever
ENTRYPOINT [ "sleep infinity" ]

`

var _ = Describe("Helper", func() {

	Describe(".getDockerFile", func() {
		It("should fetch docker file from github using a commit hash as version", func() {

			expectedResult := strings.TrimSpace(expectedBody)
			res, err := getDockerFile()
			Expect(err).To(BeNil())

			body := strings.TrimSpace(res)

			Expect(res).NotTo(BeNil())
			Expect(body).To(Equal(expectedResult))

		})
	})

	Describe(".buildImage", func() {
		var dockerfile string
		BeforeEach(func() {
			dockerfile, _ = getDockerFile()
		})
		It("should build image from  docker file", func() {
			image, err := buildImage(dockerfile)
			Expect(err).To(BeNil())
			Expect(image).NotTo(BeNil())
			Expect(image.ID).NotTo(BeNil())
		})

		Describe(".getImage", func() {
			It("should get image if it exists", func() {
				cli, err := client.NewClientWithOpts()
				Expect(err).To(BeNil())
				image := getImage(cli)

				Expect(image).NotTo(BeNil())
				Expect(reflect.ValueOf(image.ID).Type()).To(Equal(reflect.TypeOf((string)(""))))
			})
		})

		Describe("BuildContext", func() {
			var buildCtx *BuildContext
			var err error
			BeforeEach(func() {
				dir := "test-hello"
				buildCtx, err = newBuildCtx(dir, "hellofile", "hello world")
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

})
