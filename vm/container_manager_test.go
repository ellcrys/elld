package vm

import (
	"github.com/ellcrys/druid/crypto"
	"github.com/ellcrys/druid/wire"

	"github.com/docker/docker/client"
	"github.com/ellcrys/druid/util"
	"github.com/ellcrys/druid/util/logger"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ContainerManager", func() {
	var cm *ContainerManager
	var containerID = util.RandString(10)
	var cmlogger = logger.NewLogrus()
	cmlogger.SetToDebug()
	var cli *client.Client
	var err error
	var container *Container

	BeforeEach(func() {
		cli, err = client.NewClientWithOpts()
		Expect(err).To(BeNil())
	})

	BeforeEach(func() {
		cm = NewContainerManager(cmlogger, cli)
		Expect(cm).NotTo(BeNil())
		Expect(cm.logger).NotTo(BeNil())
		Expect(cm.client).NotTo(BeNil())
	})

	Describe(".create", func() {
		AfterEach(func() {
			defer cli.Close()
			err := container.stop()
			Expect(err).To(BeNil())
			err = container.destroy()
			Expect(err).To(BeNil())
		})

		It("should create a container", func() {
			container, err = cm.create(containerID)
			Expect(err).To(BeNil())
			Expect(container).NotTo(BeNil())
			Expect(container.id).NotTo(BeEmpty())
			Expect(container.port).NotTo(BeZero())
		})
	})

	Describe(".find", func() {
		BeforeEach(func() {
			container, err = cm.create(containerID)
			Expect(err).To(BeNil())
			Expect(container).NotTo(BeNil())
		})

		AfterEach(func() {
			defer cli.Close()
			err := container.stop()
			Expect(err).To(BeNil())
			err = container.destroy()
			Expect(err).To(BeNil())
		})

		It("should find a container by it's id", func() {
			container := cm.find(containerID)
			Expect(container).NotTo(BeNil())
		})

		It("should fail find a container", func() {
			container := cm.find("")
			Expect(container).To(BeNil())
		})

	})

	Describe(".run", func() {

		BeforeEach(func() {
			container, err = cm.create(containerID)
			Expect(err).To(BeNil())
			Expect(container).NotTo(BeNil())
		})

		AfterEach(func() {
			defer cli.Close()
			err := container.stop()
			Expect(err).To(BeNil())
			err = container.destroy()
			Expect(err).To(BeNil())
		})

		It("should run blockcodes", func() {
			seed := int64(1)
			a, _ := crypto.NewKey(&seed)

			tx := &wire.Transaction{
				Type:         1,
				Nonce:        1,
				To:           "some_address",
				SenderPubKey: a.PubKey().Base58(),
				BlockcodeParams: &wire.BlockcodeParams{
					Func: "DoSomething",
					Data: []byte("hello world"),
				},
			}

			done := make(chan error, 1)
			output := make(chan []byte)

			cm.Run(tx, output, done)

			Expect(<-done).To(BeNil())
			Expect(<-output).NotTo(BeEmpty())
		})
	})

	// Describe(".len", func() {
	// 	It("should return length of running containers", func() {

	// 	})
	// })
})
