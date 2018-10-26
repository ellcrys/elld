package vm

import (
	"fmt"
	"testing"

	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util/logger"
	docker "github.com/fsouza/go-dockerclient"
	. "github.com/ncodes/goblin"
	. "github.com/onsi/gomega"
)

func TestContainerManager(t *testing.T) {
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("ContainerManager", func() {
		var cm *ContainerManager
		var cli *docker.Client
		var err error
		var co *Container
		var image *Image
		var dckFileURL = fmt.Sprintf(dockerFileURL, dockerFileHash)
		var log = logger.NewLogrus()
		log.SetToDebug()

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
			cm = NewContainerManager(cli, image, log)
			Expect(cm).NotTo(BeNil())
		})

		g.BeforeEach(func() {
			cli, err = docker.NewClient(dockerEndpoint)
			Expect(err).To(BeNil())
		})

		g.BeforeEach(func() {
			co = NewContainer(cli, image, log)
		})

		g.AfterEach(func() {
			err = co.destroy()
			Expect(err).To(BeNil())
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

			g.It("should create a container", func() {
				var output = make(chan []byte)
				var errCh = make(chan error)

				tx := &core.Transaction{
					To: "blockcode_0",
					InvokeArgs: &core.InvokeArgs{
						Func: "some_func",
						Params: map[string][]byte{
							"amount": []byte("100"),
						},
					},
				}

				go cm.execTx(tx, output, errCh)
				fmt.Println(<-output)
				Expect(<-errCh).To(BeNil())
			})
		})

		// g.Describe(".find", func() {
		// 	g.BeforeEach(func() {
		// 		container, err = cm.create(containerID)
		// 		Expect(err).To(BeNil())
		// 		Expect(container).NotTo(BeNil())
		// 	})

		// 	g.AfterEach(func() {
		// 		defer cli.Close()
		// 		err := container.stop()
		// 		Expect(err).To(BeNil())
		// 		err = container.destroy()
		// 		Expect(err).To(BeNil())
		// 	})

		// 	g.It("should find a container by it's id", func() {
		// 		container := cm.find(containerID)
		// 		Expect(container).NotTo(BeNil())
		// 	})

		// 	g.It("should fail find a container", func() {
		// 		container := cm.find("")
		// 		Expect(container).To(BeNil())
		// 	})

		// })

		// g.Describe(".run", func() {

		// 	g.BeforeEach(func() {
		// 		container, err = cm.create(containerID)
		// 		Expect(err).To(BeNil())
		// 		Expect(container).NotTo(BeNil())
		// 	})

		// 	g.AfterEach(func() {
		// 		defer cli.Close()
		// 		err := container.stop()
		// 		Expect(err).To(BeNil())
		// 		err = container.destroy()
		// 		Expect(err).To(BeNil())
		// 	})

		// 	g.It("should run blockcodes", func() {
		// 		seed := int64(1)
		// 		a, _ := crypto.NewKey(&seed)

		// 		tx := &wire.Transaction{
		// 			Type:         1,
		// 			Nonce:        1,
		// 			To:           "some_address",
		// 			SenderPubKey: a.PubKey().Base58(),
		// 		}

		// 		done := make(chan error, 1)
		// 		output := make(chan []byte)

		// 		cm.Run(tx, output, done)

		// 		Expect(<-done).To(BeNil())
		// 		Expect(<-output).NotTo(BeEmpty())
		// 	})
		// })

		// g.Describe(".len", func() {
		// 	g.It("should return length of running containers", func() {

		// 	})
		// })
	})
}
