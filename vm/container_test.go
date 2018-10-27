package vm

import (
	"fmt"
	"testing"

	"github.com/ellcrys/elld/blockcode"
	"github.com/ellcrys/elld/util/logger"
	docker "github.com/fsouza/go-dockerclient"
	. "github.com/ncodes/goblin"
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

func TestContainer(t *testing.T) {
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Container", func() {
		containerStopTimeout = 1
		var log = logger.NewLogrus()
		log.SetToDebug()
		var co *Container
		var err error
		var dckFileURL = fmt.Sprintf(dockerFileURL, dockerFileHash)
		var cli *docker.Client
		var image *Image

		g.BeforeEach(func() {
			cli, err = docker.NewClient(dockerEndpoint)
			Expect(err).To(BeNil())
		})

		g.BeforeEach(func() {
			builder := NewImageBuilder(log, cli, dckFileURL)
			image, err = builder.Build()
			Expect(err).To(BeNil())
			Expect(image).ToNot(BeNil())
		})

		g.BeforeEach(func() {
			co = NewContainer(cli, image, log)
		})

		g.AfterEach(func() {
			err = co.destroy()
			Expect(err).To(BeNil())
		})

		g.Describe(".create", func() {

			g.It("should create the container", func() {
				err := co.create()
				Expect(err).To(BeNil())
				container, err := cli.InspectContainer(co._container.ID)
				Expect(err).To(BeNil())
				Expect(container).ToNot(BeNil())
			})
		})

		g.Describe(".start", func() {

			g.BeforeEach(func() {
				err := co.create()
				Expect(err).To(BeNil())
			})

			g.AfterEach(func() {
				ci, err := cli.InspectContainer(co._container.ID)
				Expect(err).To(BeNil())
				if ci.State.Running {
					err = cli.StopContainer(co._container.ID, 1)
					Expect(err).To(BeNil())
				}
			})

			g.It("should start a container", func() {
				err := co.start()
				Expect(err).To(BeNil())
				ci, err := cli.InspectContainer(co._container.ID)
				Expect(err).To(BeNil())
				Expect(ci.State.Running).To(BeTrue())
			})

			g.It("should return err = 'API error (404): No such container: <fake container id>'", func() {
				original := co._container.ID
				co._container.ID = "<fake container id>"
				err := co.start()
				Expect(err).NotTo(BeNil())
				co._container.ID = original
			})
		})

		g.Describe(".exec", func() {

			g.BeforeEach(func() {
				err := co.create()
				Expect(err).To(BeNil())
			})

			g.BeforeEach(func() {
				err := co.start()
				Expect(err).To(BeNil())
			})

			g.It("should execute a command in the container", func() {
				command := []string{"bash", "-c", "echo hello"}
				statusCode, err := co.exec(command)
				Expect(err).To(BeNil())
				Expect(statusCode).To(Equal(0))
			})
		})

		g.Describe(".copy", func() {
			var bc *blockcode.Blockcode

			g.BeforeEach(func() {
				bc, err = blockcode.FromDir("./testdata/blockcode_example")
				Expect(err).To(BeNil())
				Expect(bc.Size()).NotTo(BeZero())
			})

			g.BeforeEach(func() {
				err := co.create()
				Expect(err).To(BeNil())
			})

			g.BeforeEach(func() {
				err := co.start()
				Expect(err).To(BeNil())
			})

			g.It("should copy content into the container", func() {
				err = co.copy(bc.GetCode())
				Expect(err).To(BeNil())
				statusCode, err := co.exec([]string{"bash", "-c", "ls " + makeCopyPath(co.id())})
				Expect(err).To(BeNil())
				Expect(statusCode).To(Equal(0))
			})

			g.It("should fail to copy content into the container", func() {
				err := co.copy(bc.Bytes())
				Expect(err).NotTo(BeNil())
			})
		})

		g.Describe(".addChild", func() {

			g.BeforeEach(func() {
				err := co.create()
				Expect(err).To(BeNil())
			})

			g.It("should add a child container", func() {
				child := NewContainer(cli, image, log)
				err := child.create()
				Expect(err).To(BeNil())
				co.addChild(child)
				Expect(len(co.children)).NotTo(BeZero())
				Expect(co.children[0].id()).To(Equal(child.id()))
				Expect(child.parent.id()).To(Equal(co.id()))
			})
		})

		g.Describe(".addBuildLang", func() {
			g.It("should set a concrete implementation of a LangBuilder", func() {
				co.setBuildLang(new(TestBuildLang))
				Expect(co.buildConfig).ToNot(BeNil())
			})
		})

		g.Describe(".build", func() {

			g.BeforeEach(func() {
				err := co.create()
				Expect(err).To(BeNil())
			})

			g.BeforeEach(func() {
				err := co.start()
				Expect(err).To(BeNil())
			})

			g.It("should attempt to build block code", func() {
				co.setBuildLang(new(TestBuildLang))
				Expect(co.buildConfig).ToNot(BeNil())
				success, err := co.build()
				Expect(err).To(BeNil())
				Expect(success).To(BeTrue())
			})

			g.It("should fail to build block code", func() {
				co.setBuildLang(new(ErrBuildLang))
				success, err := co.build()
				Expect(err).To(BeNil())
				Expect(success).To(BeFalse())
			})
		})

	})
}
