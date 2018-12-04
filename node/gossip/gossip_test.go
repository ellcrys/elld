package gossip_test

import (
	"context"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/node"
	"github.com/ellcrys/elld/types/core"
	net "github.com/libp2p/go-libp2p-net"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Gossip", func() {

	var lp, rp *node.Node

	BeforeEach(func() {
		lp = makeTestNode(getPort())
		rp = makeTestNode(getPort())
	})

	AfterEach(func() {
		closeNode(lp)
		closeNode(rp)
	})

	Describe(".checkRemotePeer", func() {

		var stream net.Stream
		var ws *core.WrappedStream
		var cc context.CancelFunc
		var err error

		When("protocol id is Handshake", func() {

			BeforeEach(func() {
				stream, cc, err = lp.Gossip().NewStream(rp, config.Versions.Handshake)
				Expect(err).To(BeNil())
				defer cc()
				defer stream.Close()
				ws = &core.WrappedStream{Stream: stream, Extra: map[string]interface{}{}}
			})

			It("should return nil", func() {
				err := lp.Gossip().CheckRemotePeer(ws, rp)
				Expect(err).To(BeNil())
			})
		})

		When("not in test mode and remote peer is unacquainted and message is not Addr", func() {

			BeforeEach(func() {
				lp.GetCfg().Node.Mode = config.ModeProd
			})

			BeforeEach(func() {
				stream, cc, err = lp.Gossip().NewStream(rp, config.Versions.GetAddr)
				Expect(err).To(BeNil())
				defer cc()
				defer stream.Close()
				ws = &core.WrappedStream{Stream: stream, Extra: map[string]interface{}{}}
			})

			It("should return err='unacquainted node'", func() {
				err := lp.Gossip().CheckRemotePeer(ws, rp)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("unacquainted node"))
			})
		})

		When("not in test mode and remote peer is unacquainted and message is Addr", func() {

			BeforeEach(func() {
				lp.GetCfg().Node.Mode = config.ModeProd
			})

			BeforeEach(func() {
				stream, cc, err = lp.Gossip().NewStream(rp, config.Versions.Addr)
				Expect(err).To(BeNil())
				defer cc()
				defer stream.Close()
				ws = &core.WrappedStream{Stream: stream, Extra: map[string]interface{}{}}
			})

			It("should return nil", func() {
				err := lp.Gossip().CheckRemotePeer(ws, rp)
				Expect(err).To(BeNil())
			})
		})
	})

})
