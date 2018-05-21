package vm

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const expectedBody = `FROM golang:1.10-stretch

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
			if err != nil {
				panic(err)
			}
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

})
