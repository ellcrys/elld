package gossip_test

import (
	"context"
	"testing"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/node"
	"github.com/ellcrys/elld/types/core"
	net "github.com/libp2p/go-libp2p-net"

	. "github.com/ncodes/goblin"
	. "github.com/onsi/gomega"
)

func TestGossip(t *testing.T) {
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Gossip", func() {

		var lp, rp *node.Node

		g.BeforeEach(func() {
			lp = makeTestNode(getPort())
			rp = makeTestNode(getPort())
		})

		g.AfterEach(func() {
			closeNode(lp)
			closeNode(rp)
		})

		g.Describe(".checkRemotePeer", func() {

			var stream net.Stream
			var ws *core.WrappedStream
			var cc context.CancelFunc
			var err error

			g.When("protocol id is Handshake", func() {

				g.BeforeEach(func() {
					stream, cc, err = lp.Gossip().NewStream(rp, config.Versions.Handshake)
					Expect(err).To(BeNil())
					defer cc()
					defer stream.Close()
					ws = &core.WrappedStream{Stream: stream, Extra: map[string]interface{}{}}
				})

				g.It("should return nil", func() {
					err := lp.Gossip().CheckRemotePeer(ws, rp)
					Expect(err).To(BeNil())
				})
			})

			g.When("not in test mode and remote peer is unacquainted and message is not Addr", func() {

				g.BeforeEach(func() {
					lp.GetCfg().Node.Mode = config.ModeProd
				})

				g.BeforeEach(func() {
					stream, cc, err = lp.Gossip().NewStream(rp, config.Versions.GetAddr)
					Expect(err).To(BeNil())
					defer cc()
					defer stream.Close()
					ws = &core.WrappedStream{Stream: stream, Extra: map[string]interface{}{}}
				})

				g.It("should return err='unacquainted node'", func() {
					err := lp.Gossip().CheckRemotePeer(ws, rp)
					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(Equal("unacquainted node"))
				})
			})

			g.When("not in test mode and remote peer is unacquainted and message is Addr", func() {

				g.BeforeEach(func() {
					lp.GetCfg().Node.Mode = config.ModeProd
				})

				g.BeforeEach(func() {
					stream, cc, err = lp.Gossip().NewStream(rp, config.Versions.Addr)
					Expect(err).To(BeNil())
					defer cc()
					defer stream.Close()
					ws = &core.WrappedStream{Stream: stream, Extra: map[string]interface{}{}}
				})

				g.It("should return nil", func() {
					err := lp.Gossip().CheckRemotePeer(ws, rp)
					Expect(err).To(BeNil())
				})
			})
		})
	})
}
