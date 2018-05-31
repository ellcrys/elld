package vm

import (
	"context"
	"fmt"
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

func (lang *TestBuildLang) Build() error {
	return nil
}

type ErrBuildLang struct {
}

func (lang *ErrBuildLang) GetRunScript() []string {
	return []string{"bash", "-c", "echo hello"}
}

func (lang *ErrBuildLang) Build() error {
	return fmt.Errorf("err %s", "an error")
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
			Volumes: map[string]struct{}{
				"/archive": struct{}{},
			},
		}, nil, nil, transactionID)
		Expect(err).To(BeNil())
		Expect(container).NotTo(BeNil())
		co.id = container.ID
		co.log = log
		co.log.SetToDebug()
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

			err = co.exec(command, nil)
			Expect(err).To(BeNil())
		})

		It("should throw error while trying to execute a command in the container", func() {
			err := co.start()
			Expect(err).To(BeNil())
			command := []string{}

			err = co.exec(command, nil)
			Expect(err).NotTo(BeNil())
		})
	})

	// Describe(".copy", func() {
	// 	var bc *blockcode.Blockcode
	// 	BeforeEach(func() {
	// 		bc, err = blockcode.FromDir("../blockcode/testdata/blockcode_example")
	// 		Expect(err).To(BeNil())
	// 		Expect(bc).NotTo(BeNil())
	// 	})
	// 	It("should copy content into the container", func() {
	// 		err := co.start()
	// 		Expect(err).To(BeNil())
	// 		err = co.copy("736729", bc.Bytes())
	// 		Expect(err).To(BeNil())
	// 	})

	// 	It("should fail to copy content into the container", func() {
	// 		err := co.copy("", bc.Bytes())
	// 		Expect(err).NotTo(BeNil())
	// 	})
	// })

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

			// done := make(chan error, 1)
			// output := make(chan []byte)
			err := co.build()
			Expect(err).To(BeNil())
		})

		It("should fail to build block code", func() {
			co.setBuildLang(new(ErrBuildLang))
			Expect(co.buildConfig).ToNot(BeNil())

			// done := make(chan error, 1)
			// output := make(chan []byte)
			err := co.build()
			Expect(err).NotTo(BeNil())
		})
	})

})
