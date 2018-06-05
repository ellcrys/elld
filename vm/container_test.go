package vm

import (
	"fmt"

	"github.com/ellcrys/elld/blockcode"
	"github.com/ellcrys/elld/util/logger"
	docker "github.com/fsouza/go-dockerclient"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type TestBuildLang struct {
}

func (lang *TestBuildLang) GetRunScript() []string {
	return []string{"bash", "-c", "echo hello"}
}

func (lang *TestBuildLang) GetBuildScript() []string {
	return []string{"bash", "-c", "echo build"}
}

type ErrBuildLang struct {
}

func (lang *ErrBuildLang) GetRunScript() []string {
	return []string{"bash", "-c", "echo hello"}
}

func (lang *ErrBuildLang) GetBuildScript() []string {
	return []string{"basjx"}
}

var _ = Describe("Container", func() {
	containerStopTimeout = 1
	var log = logger.NewLogrus()
	log.SetToDebug()
	var co *Container
	var err error
	var dckFileURL = fmt.Sprintf(dockerFileURL, dockerFileHash)
	var cli *docker.Client
	var image *Image

	BeforeEach(func() {
		cli, err = docker.NewClient(dockerEndpoint)
		Expect(err).To(BeNil())
	})

	BeforeEach(func() {
		builder := NewImageBuilder(log, cli, dckFileURL)
		image, err = builder.Build()
		Expect(err).To(BeNil())
		Expect(image).ToNot(BeNil())
	})

	BeforeEach(func() {
		co = NewContainer(cli, image, log)
	})

	AfterEach(func() {
		err = co.destroy()
		Expect(err).To(BeNil())
	})

	Describe(".create", func() {

		It("should create the container", func() {
			err := co.create()
			Expect(err).To(BeNil())
			container, err := cli.InspectContainer(co._container.ID)
			Expect(err).To(BeNil())
			Expect(container).ToNot(BeNil())
		})
	})

	Describe(".start", func() {

		BeforeEach(func() {
			err := co.create()
			Expect(err).To(BeNil())
		})

		AfterEach(func() {
			ci, err := cli.InspectContainer(co._container.ID)
			Expect(err).To(BeNil())
			if ci.State.Running {
				err = cli.StopContainer(co._container.ID, 1)
				Expect(err).To(BeNil())
			}
		})

		It("should start a container", func() {
			err := co.start()
			Expect(err).To(BeNil())
			ci, err := cli.InspectContainer(co._container.ID)
			Expect(err).To(BeNil())
			Expect(ci.State.Running).To(BeTrue())
		})

		It("should return err = 'API error (404): No such container: <fake container id>'", func() {
			original := co._container.ID
			co._container.ID = "<fake container id>"
			err := co.start()
			Expect(err).NotTo(BeNil())
			co._container.ID = original
		})
	})

	Describe(".exec", func() {

		BeforeEach(func() {
			err := co.create()
			Expect(err).To(BeNil())
		})

		BeforeEach(func() {
			err := co.start()
			Expect(err).To(BeNil())
		})

		It("should execute a command in the container", func() {
			command := []string{"bash", "-c", "echo hello"}
			statusCode, err := co.exec(command)
			Expect(err).To(BeNil())
			Expect(statusCode).To(Equal(0))
		})
	})

	Describe(".copy", func() {
		var bc *blockcode.Blockcode

		BeforeEach(func() {
			bc, err = blockcode.FromDir("./testdata/blockcode_example")
			Expect(err).To(BeNil())
			Expect(bc.Len()).NotTo(BeZero())
		})

		BeforeEach(func() {
			err := co.create()
			Expect(err).To(BeNil())
		})

		BeforeEach(func() {
			err := co.start()
			Expect(err).To(BeNil())
		})

		It("should copy content into the container", func() {
			err = co.copy(bc.GetCode())
			Expect(err).To(BeNil())
			statusCode, err := co.exec([]string{"bash", "-c", "ls " + makeCopyPath(co.id())})
			Expect(err).To(BeNil())
			Expect(statusCode).To(Equal(0))
		})

		It("should fail to copy content into the container", func() {
			err := co.copy(bc.Bytes())
			Expect(err).NotTo(BeNil())
		})
	})

	Describe(".addChild", func() {

		BeforeEach(func() {
			err := co.create()
			Expect(err).To(BeNil())
		})

		It("should add a child container", func() {
			child := NewContainer(cli, image, log)
			err := child.create()
			Expect(err).To(BeNil())
			co.addChild(child)
			Expect(len(co.children)).NotTo(BeZero())
			Expect(co.children[0].id()).To(Equal(child.id()))
			Expect(child.parent.id()).To(Equal(co.id()))
		})
	})

	Describe(".addBuildLang", func() {
		It("should set a concrete implementation of a LangBuilder", func() {
			co.setBuildLang(new(TestBuildLang))
			Expect(co.buildConfig).ToNot(BeNil())
		})
	})

	Describe(".build", func() {

		BeforeEach(func() {
			err := co.create()
			Expect(err).To(BeNil())
		})

		BeforeEach(func() {
			err := co.start()
			Expect(err).To(BeNil())
		})

		It("should attempt to build block code", func() {
			co.setBuildLang(new(TestBuildLang))
			Expect(co.buildConfig).ToNot(BeNil())
			success, err := co.build()
			Expect(err).To(BeNil())
			Expect(success).To(BeTrue())
		})

		It("should fail to build block code", func() {
			co.setBuildLang(new(ErrBuildLang))
			success, err := co.build()
			Expect(err).To(BeNil())
			Expect(success).To(BeFalse())
		})
	})

})
