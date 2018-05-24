package vm

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/ellcrys/druid/util/logger"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type TestBuildLang struct {
}

func (lang *TestBuildLang) GetRunScript() []string {
	return []string{"bash", "-c", "echo hello"}
}

func (lang *TestBuildLang) Build(containerID string) error {
	return nil
}

var _ = Describe("Container", func() {
	containerStopTimeout = time.Millisecond * 500
	var log = logger.NewLogrusNoOp()
	var co *Container
	var transactionID = "983007356499672"
	var err error
	var dckFileURL = fmt.Sprintf(dockerFileURL, dockerFileHash)
	var cli *client.Client
	var image *Image

	BeforeEach(func() {
		cli, err = client.NewClientWithOpts()
		Expect(err).To(BeNil())

		builder := NewImageBuilder(log, cli, dckFileURL)
		image, err = builder.Build()
		Expect(err).To(BeNil())
		Expect(image).ToNot(BeNil())
	})

	BeforeEach(func() {
		co = new(Container)
		co.dockerCli = cli
		co.children = []*Container{}
		Expect(err).To(BeNil())
		container, err := cli.ContainerCreate(context.Background(), &container.Config{
			Image: image.ID,
		}, nil, nil, transactionID)
		Expect(err).To(BeNil())
		Expect(container).NotTo(BeNil())
		co.id = container.ID
		co.log = log
	})

	AfterEach(func() {
		defer cli.Close()
		err := co.stop()
		Expect(err).To(BeNil())
		err = co.destroy()
		Expect(err).To(BeNil())
	})

	Describe(".start", func() {
		It("should start a container", func() {
			err := co.start()
			Expect(err).To(BeNil())
		})
	})
	Describe(".exec", func() {
		It("should execute a command in the container", func() {
			_ = co.start()
			command := []string{"bash", "-c", "echo hello"}
			done := make(chan error)
			output := make(chan string)
			go co.exec(command, output, done)
			Expect(<-output).NotTo(BeEmpty())
			Expect(<-done).To(BeNil())
		})
	})

	Describe(".addChild", func() {
		It("should add a child container", func() {
			child := new(Container)
			child.id = "1c6de1dbaad3"
			co.addChild(child)
			Expect(len(co.children)).NotTo(BeZero())
			Expect(co.children[0].id).To(Equal(child.id))
			Expect(child.parent.id).To(Equal(co.id))
		})
	})

	Describe(".addBuildLang", func() {
		It("should set a concrete implementation of a LangBuilder", func() {
			co.setBuildLang(new(TestBuildLang))
			Expect(co.buildConfig).ToNot(BeNil())
		})
	})

	Describe(".build", func() {
		It("should attempt to build block code", func() {
			co.setBuildLang(new(TestBuildLang))
			Expect(co.buildConfig).ToNot(BeNil())
			err := co.build()
			Expect(err).To(BeNil())
		})
	})

})
