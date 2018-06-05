package vm

import (
	"fmt"

	"github.com/ellcrys/elld/util/logger"
	"github.com/ellcrys/elld/wire"
	docker "github.com/fsouza/go-dockerclient"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ContainerManager", func() {

	var cm *ContainerManager
	var cli *docker.Client
	var err error
	var co *Container
	var image *Image
	var dckFileURL = fmt.Sprintf(dockerFileURL, dockerFileHash)
	var log = logger.NewLogrus()
	log.SetToDebug()

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
		cm = NewContainerManager(cli, image, log)
		Expect(cm).NotTo(BeNil())
	})

	BeforeEach(func() {
		cli, err = docker.NewClient(dockerEndpoint)
		Expect(err).To(BeNil())
	})

	BeforeEach(func() {
		co = NewContainer(cli, image, log)
	})

	AfterEach(func() {
		err = co.destroy()
		Expect(err).To(BeNil())
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

		It("should create a container", func() {
			var output = make(chan []byte)
			var errCh = make(chan error)

			tx := &wire.Transaction{
				To: "blockcode_0",
				InvokeArgs: &wire.InvokeArgs{
					Func: "some_func",
					Params: map[string]string{
						"amount": "100",
					},
				},
			}

			go cm.execTx(tx, output, errCh)
			fmt.Println(<-output)
			Expect(<-errCh).To(BeNil())
		})
	})

	// Describe(".find", func() {
	// 	BeforeEach(func() {
	// 		container, err = cm.create(containerID)
	// 		Expect(err).To(BeNil())
	// 		Expect(container).NotTo(BeNil())
	// 	})

	// 	AfterEach(func() {
	// 		defer cli.Close()
	// 		err := container.stop()
	// 		Expect(err).To(BeNil())
	// 		err = container.destroy()
	// 		Expect(err).To(BeNil())
	// 	})

	// 	It("should find a container by it's id", func() {
	// 		container := cm.find(containerID)
	// 		Expect(container).NotTo(BeNil())
	// 	})

	// 	It("should fail find a container", func() {
	// 		container := cm.find("")
	// 		Expect(container).To(BeNil())
	// 	})

	// })

	// Describe(".run", func() {

	// 	BeforeEach(func() {
	// 		container, err = cm.create(containerID)
	// 		Expect(err).To(BeNil())
	// 		Expect(container).NotTo(BeNil())
	// 	})

	// 	AfterEach(func() {
	// 		defer cli.Close()
	// 		err := container.stop()
	// 		Expect(err).To(BeNil())
	// 		err = container.destroy()
	// 		Expect(err).To(BeNil())
	// 	})

	// 	It("should run blockcodes", func() {
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

	// Describe(".len", func() {
	// 	It("should return length of running containers", func() {

	// 	})
	// })
})
