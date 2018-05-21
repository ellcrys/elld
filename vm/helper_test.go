package vm

import (
	"fmt"
	"reflect"
	"strings"

	homedir "github.com/mitchellh/go-homedir"

	"github.com/docker/docker/client"
	"github.com/franela/goreq"
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
			commitHash := "c0879257e8136bf13b4fceb5651f751b806782a7"

			expectedResult := strings.TrimSpace(expectedBody)

			res, err := getDockerFile(commitHash)
			Expect(err).To(BeNil())

			result, _ := res.Body.ToString()
			body := strings.TrimSpace(result)

			Expect(res.Body).NotTo(BeNil())
			Expect(body).To(Equal(expectedResult))

		})

		It("should fail to fetch docker file if commit hash is invalid", func() {
			commitHash := "<invalid commit hash>"
			_, err := getDockerFile(commitHash)
			Expect(err.Error()).To(Equal("Docker file not found"))
		})
	})

	Describe(".buildImage", func() {
		var commitHash string
		var imageRes *goreq.Response
		BeforeEach(func() {
			commitHash = "c0879257e8136bf13b4fceb5651f751b806782a7"
			imageRes, _ = getDockerFile(commitHash)
		})
		It("should build image from  docker file", func() {
			image, err := buildImage(imageRes)
			Expect(err).To(BeNil())
			Expect(image).NotTo(BeNil())
			Expect(image.ID).NotTo(BeNil())
		})

		Describe(".getImage", func() {
			It("should get image if it exists", func() {
				cli, _ := client.NewEnvClient()
				image := getImage(cli)

				Expect(image).NotTo(BeNil())
				Expect(reflect.ValueOf(image.ID).Type()).To(Equal(reflect.TypeOf((string)(""))))
			})
		})

		Describe("BuildContext", func() {
			var buildCtx *BuildContext
			var err error
			BeforeEach(func() {
				homeDir, _ := homedir.Dir()
				dir := fmt.Sprintf("%s/.ellcrys/test-hello", homeDir)
				buildCtx, err = newBuildCtx(dir, "hellofile", "hello world")
				Expect(err).To(BeNil())
				Expect(buildCtx).NotTo(BeNil())
			})

			AfterEach(func() {
				buildCtx.Close()
			})

			Describe(".Reader", func() {
				It("should create a stream from buildCtx", func() {
					reader, err := buildCtx.Reader()
					if err != nil {
						panic(err)
					}
					Expect(err).To(BeNil())
					Expect(reader).NotTo(BeNil())
				})
			})
		})
	})

})
