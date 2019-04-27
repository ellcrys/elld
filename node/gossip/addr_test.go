package gossip_test

import (
	"fmt"
	"time"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/node"
	"github.com/ellcrys/elld/node/gossip"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"
	"github.com/olebedev/emitter"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TestAddr", func() {

	var lp, rp *node.Node
	var sender, _ = crypto.NewKey(nil)
	var lpPort, rpPort int

	BeforeEach(func() {
		lpPort = getPort()
		rpPort = getPort()

		lp = makeTestNodeWith(lpPort, 399)
		Expect(lp.GetBlockchain().Up()).To(BeNil())

		rp = makeTestNodeWith(rpPort, 382)
		Expect(rp.GetBlockchain().Up()).To(BeNil())

		// On the remote node blockchain,
		// Create the sender's account
		// with some initial balance
		Expect(rp.GetBlockchain().CreateAccount(1, rp.GetBlockchain().GetBestChain(), &core.Account{
			Type:    core.AccountTypeBalance,
			Address: util.String(sender.Addr()),
			Balance: "100",
		})).To(BeNil())

		Expect(lp.Connect(rp)).To(BeNil())
	})

	AfterEach(func() {
		closeNode(lp)
		closeNode(rp)
	})

	Describe(".RelayAddresses", func() {

		It("should return err='too many items in addr message' when address is more than 10", func() {
			addrs := []*core.Address{
				{Address: ""}, {Address: ""}, {Address: ""}, {Address: ""}, {Address: ""}, {Address: ""}, {Address: ""},
				{Address: ""}, {Address: ""}, {Address: ""}, {Address: ""},
			}
			errs := lp.Gossip().RelayAddresses(addrs)
			Expect(errs).ToNot(BeEmpty())
			Expect(errs).To(ContainElement(fmt.Errorf("too many addresses in the message")))
		})

		It("should return err='no addr to relay' if non of the addresses where relayable", func() {
			addrs := []*core.Address{
				{Address: ""},
				{Address: ""},
			}
			errs := lp.Gossip().RelayAddresses(addrs)
			Expect(errs).ToNot(BeEmpty())
			Expect(errs).To(ContainElement(fmt.Errorf("no addr to relay")))
		})

		Describe("in production mode", func() {
			It("should return err when an address is not routable", func() {
				lp.GetCfg().Node.Mode = config.ModeProd
				addrs := []*core.Address{
					{Address: "/ip4/0.0.0.0/tcp/9000/ipfs/12D3KooWHHzSeKaY8xuZVzkLbKFfvNgPPeKhFBGrMbNzbm5akpqu"},
				}
				errs := lp.Gossip().RelayAddresses(addrs)
				Expect(errs).ToNot(BeEmpty())
				Expect(errs).To(ContainElement(fmt.Errorf("address {/ip4/0.0.0.0/tcp/9000/ipfs/12D3KooWHHzSeKaY8xuZVzkLbKFfvNgPPeKhFBGrMbNzbm5akpqu} is not routable")))
			})
		})

		It("should return err='no addr to relay' if address timestamp over 60 minutes", func() {
			addrs := []*core.Address{
				{Address: "/ip4/127.0.0.1/tcp/9000/ipfs/12D3KooWHHzSeKaY8xuZVzkLbKFfvNgPPeKhFBGrMbNzbm5akpqu", Timestamp: time.Now().Add(61 * time.Minute).Unix()},
			}
			errs := lp.Gossip().RelayAddresses(addrs)
			Expect(errs).ToNot(BeEmpty())
			Expect(errs).To(ContainElement(fmt.Errorf("no addr to relay")))
		})

		Describe("with 3 address to relay", func() {

			var p, p2, p3 *node.Node

			BeforeEach(func() {
				p = makeTestNode(getPort())
				p2 = makeTestNode(getPort())
				p3 = makeTestNode(getPort())
				lp.PM().AddPeer(rp)
			})

			AfterEach(func() {
				closeNode(p)
				closeNode(p2)
				closeNode(p3)
			})

			It("should successfully select 2 relay peers", func() {
				addrs := []*core.Address{
					{Address: p.GetAddress(), Timestamp: time.Now().Unix()},
					{Address: p2.GetAddress(), Timestamp: time.Now().Unix()},
					{Address: p3.GetAddress(), Timestamp: time.Now().Unix()},
				}
				Expect(p.Gossip().GetRandBroadcasters().Len()).To(Equal(0))
				p.Gossip().RelayAddresses(addrs)
				Expect(p.Gossip().GetRandBroadcasters().Len()).To(Equal(2))
			})
		})
	})

	Describe(".OnAddr", func() {

		var p, p2, p3, p4 *node.Node
		var evt emitter.Event
		var addrMsg *core.Addr

		BeforeEach(func() {
			p = makeTestNode(getPort())
			p2 = makeTestNode(getPort())
			p3 = makeTestNode(getPort())
			p4 = makeTestNode(getPort())
			addrMsg = &core.Addr{
				Addresses: []*core.Address{
					{Address: p.GetAddress(), Timestamp: time.Now().Unix()},
					{Address: p2.GetAddress(), Timestamp: time.Now().Unix()},
					{Address: p3.GetAddress(), Timestamp: time.Now().Unix()},
					{Address: p4.GetAddress(), Timestamp: time.Now().Unix()},
				},
			}
		})

		AfterEach(func() {
			closeNode(p)
			closeNode(p2)
			closeNode(p3)
			closeNode(p4)
		})

		Describe("when the number of addresses is below max address expected", func() {

			It("should select relay peers from the relayed addresses", func(done Done) {
				wait := make(chan bool)

				stream, c, err := lp.Gossip().NewStream(rp, config.GetVersions().Addr)
				Expect(err).To(BeNil())

				err = gossip.WriteStream(stream, addrMsg)
				Expect(err).To(BeNil())
				defer c()
				defer stream.Close()

				go func() {
					defer GinkgoRecover()
					evt = <-rp.GetEventEmitter().On(gossip.EventAddressesRelayed)
					Expect(evt.Args).To(BeEmpty())
					Expect(rp.Gossip().GetRandBroadcasters().Len()).To(Equal(3))
					close(wait)
				}()

				<-wait
				close(done)
			})
		})

		Context("when the number of addresses is above max address expected", func() {
			// It("should return no error", func(done Done) {
			// 	wait := make(chan bool)

			// 	rp.GetCfg().Node.MaxAddrsExpected = 1
			// 	stream, c, err := lp.Gossip().NewStream(rp, config.GetVersions().Addr)

			// 	Expect(err).To(BeNil())
			// 	defer c()
			// 	defer stream.Close()

			// 	err = gossip.WriteStream(stream, addrMsg)
			// 	Expect(err).To(BeNil())

			// 	go func() {
			// 		defer GinkgoRecover()
			// 		evt = <-rp.GetEventEmitter().On(gossip.EventAddrProcessed)
			// 		Expect(evt.Args).ToNot(BeEmpty())
			// 		Expect(evt.Args[0].(error).Error()).To(Equal("too many addresses received. Ignoring addresses"))
			// 		close(wait)
			// 	}()

			// 	<-wait
			// 	close(done)
			// }, 5)
		})

		Context("when an address has same peer ID as the local peer", func() {

			BeforeEach(func() {
				addrMsg = &core.Addr{
					Addresses: []*core.Address{
						{Address: rp.GetAddress(), Timestamp: time.Now().Unix()},
					},
				}
			})

			It("should not add the address as a peer", func(done Done) {
				wait := make(chan bool)
				stream, c, err := lp.Gossip().NewStream(rp, config.GetVersions().Addr)

				Expect(err).To(BeNil())
				defer c()
				defer stream.Close()

				err = gossip.WriteStream(stream, addrMsg)
				Expect(err).To(BeNil())

				go func() {
					defer GinkgoRecover()
					evt = <-rp.GetEventEmitter().On(gossip.EventAddrProcessed)
					Expect(evt.Args).To(BeEmpty())
					close(wait)
				}()

				<-wait
				added := rp.PM().PeerExist(addrMsg.Addresses[0].Address.StringID())
				Expect(added).To(BeFalse())
				close(done)
			}, 5)
		})
	})

})
