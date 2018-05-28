package vm

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/ellcrys/druid/util"
	"github.com/ellcrys/druid/util/logger"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type TestBuildLang struct {
}

func (lang *TestBuildLang) GetRunScript() []string {
	return []string{"bash", "-c", "echo hello"}
}

func (lang *TestBuildLang) Build(mtx *sync.Mutex) ([]byte, error) {
	mtx.Lock()
	b := []byte("hello")
	mtx.Unlock()
	return b, nil
}

type ErrBuildLang struct {
}

func (lang *ErrBuildLang) GetRunScript() []string {
	return []string{"bash", "-c", "echo hello"}
}

func (lang *ErrBuildLang) Build(mtx *sync.Mutex) ([]byte, error) {
	return nil, fmt.Errorf("err %s", "an error")
}

var _ = Describe("Container", func() {
	containerStopTimeout = time.Millisecond * 500
	var log = logger.NewLogrusNoOp()
	var co *Container
	var transactionID string
	var err error
	var dckFileURL = fmt.Sprintf(dockerFileURL, dockerFileHash)
	var cli *client.Client
	var image *Image
	var mtx sync.Mutex

	BeforeEach(func() {
		transactionID = util.RandString(5)

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

		It("should fail to start a container", func() {
			original := co.id
			co.id = "<fake container id>"
			err := co.start()
			Expect(err).NotTo(BeNil())
			co.id = original
		})

		It("should fail to stop a container", func() {
			original := co.id
			co.id = "<fake container id>"
			err := co.stop()
			Expect(err).NotTo(BeNil())
			co.id = original
		})

		It("should fail to destroy a container", func() {
			original := co.id
			co.id = "<fake container id>"
			err := co.destroy()
			Expect(err).NotTo(BeNil())
			co.id = original
		})
	})
	Describe(".exec", func() {
		It("should execute a command in the container", func() {
			err := co.start()
			Expect(err).To(BeNil())
			command := []string{"bash", "-c", "echo hello"}
			done := make(chan error, 1)
			output := make(chan []byte)
			go co.exec(command, output, done)
			Expect(<-output).NotTo(BeEmpty())
			Expect(<-done).To(BeNil())
		})

		It("should throw error while trying to execute a command in the container", func() {
			err := co.start()
			Expect(err).To(BeNil())
			command := []string{}
			done := make(chan error, 1)
			output := make(chan []byte)
			co.exec(command, output, done)
			Expect(<-done).NotTo(BeNil())
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

			done := make(chan error, 1)
			output := make(chan []byte)
			go co.build(&mtx, output, done)
			Expect(<-output).NotTo(BeEmpty())
			Expect(<-done).To(BeNil())
		})

		It("should fail to build block code", func() {
			co.setBuildLang(new(ErrBuildLang))
			Expect(co.buildConfig).ToNot(BeNil())

			done := make(chan error, 1)
			output := make(chan []byte)
			co.build(&mtx, output, done)
			Expect(<-done).NotTo(BeNil())
		})
	})

})
