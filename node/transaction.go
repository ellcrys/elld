package node

import (
	"context"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/node/histcache"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/wire"
	net "github.com/libp2p/go-libp2p-net"
)

func makeTxHistoryKey(tx *wire.Transaction, peer *Node) histcache.MultiKey {
	return []interface{}{tx.ID(), peer.StringID()}
}

// OnTx handles incoming transaction message
func (g *Gossip) OnTx(s net.Stream) {

	defer s.Close()

	remotePeer := NewRemoteNode(util.FullRemoteAddressFromStream(s), g.engine)
	remotePeerIDShort := remotePeer.ShortID()

	g.log.Info("Received new transaction", "PeerID", remotePeerIDShort)

	// read the message
	msg := &wire.Transaction{}
	if err := readStream(s, msg); err != nil {
		s.Reset()
		g.log.Error("failed to read tx message", "Err", err, "PeerID", remotePeerIDShort)
		return
	}

	// make a key for this transaction to be added to the history cache so we always know
	// when we have processed this transaction in case we see it again.
	historyKey := makeTxHistoryKey(msg, remotePeer)

	// check if we have an history about this transaction with the remote peer,
	// if no, add the transaction.
	if !g.engine.History().Has(historyKey) {

		// Add the transaction to the transaction pool and wait for error response
		var errCh = make(chan error)
		g.engine.logicEvt.Publish("transaction.add", msg, errCh)
		if err := <-errCh; err != nil {
			s.Reset()
			g.log.Error("failed to add transaction to pool", "Err", err)
			return
		}

		// add transaction to the history cache using the key we created earlier
		g.engine.History().Add(historyKey)
	}

	g.log.Info("Added new transaction to pool", "TxID", msg.ID())
}

// RelayTx relays transactions to peers
func (g *Gossip) RelayTx(tx *wire.Transaction, remotePeers []*Node) error {

	txID := tx.ID()
	sent := 0

	g.log.Debug("Relaying transaction to peers", "TxID", txID, "NumPeers", len(remotePeers))

	for _, peer := range remotePeers {

		historyKey := makeTxHistoryKey(tx, peer)

		// check if we have an history of transaction with this remote peer,
		// if yes, do not relay
		if g.engine.History().Has(historyKey) {
			continue
		}

		// create a stream to the remote peer
		s, err := g.newStream(context.Background(), peer, config.TxVersion)
		if err != nil {
			g.log.Debug("Tx message failed. failed to connect to peer", "Err", err, "PeerID", peer.ShortID())
			continue
		}
		defer s.Close()

		// write to the stream
		if err := writeStream(s, tx); err != nil {
			s.Reset()
			g.log.Debug("Tx message failed. failed to write to stream", "Err", err, "PeerID", peer.ShortID())
			continue
		}

		// add new history
		g.engine.History().Add(historyKey)

		sent++
	}

	g.log.Info("Finished relaying transaction", "TxID", txID, "NumPeersSentTo", sent)

	return nil
}
