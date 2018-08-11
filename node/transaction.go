package node

import (
	"context"

	"github.com/ellcrys/elld/blockchain"
	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/node/histcache"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/wire"
	net "github.com/libp2p/go-libp2p-net"
)

func makeTxHistoryKey(tx *wire.Transaction, peer types.Engine) histcache.MultiKey {
	return []interface{}{tx.ID(), peer.StringID()}
}

// addTransaction adds a transaction to the transaction pool.
func (n *Node) addTransaction(tx *wire.Transaction) error {

	// Validate the transactions
	txValidator := blockchain.NewTxValidator(tx, n.GetTxPool(), n.GetBlockchain(), true)
	if errs := txValidator.Validate(); len(errs) > 0 {
		return errs[0]
	}

	switch tx.Type {
	case wire.TxTypeBalance:

		// Add the transaction to the transaction pool where
		// it will be broadcast to other peers and included in a block
		if err := n.GetTxPool().Put(tx); err != nil {
			return err
		}

		return nil

	default:
		return wire.ErrTxTypeUnknown
	}
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

	// AllocCoin transactions are meant to be added by a miner
	// to a block and not relayed like regular transactions.
	if msg.Type == wire.TxTypeAllocCoin {
		s.Reset()
		g.log.Error("cannot add <AllocCoin> transaction to pool")
		return
	}

	// Validate the transaction and check whether
	// it already exists in the transaction pool,
	// main chain and side chains. If so, reject it
	if errs := blockchain.NewTxValidator(msg, g.engine.GetTxPool(), g.engine.bchain, true).Validate(); len(errs) > 0 {
		s.Reset()
		g.log.Debug("Transaction is not valid", "Err", errs[0])
		return
	}

	// make a key for this transaction to be added to the history cache so we always know
	// when we have processed this transaction in case we see it again.
	historyKey := makeTxHistoryKey(msg, remotePeer)

	// check if we have an history about this transaction with the remote peer,
	// if no, add the transaction.
	if !g.engine.History().Has(historyKey) {

		// Add the transaction to the transaction pool and wait for error response
		if err := g.engine.addTransaction(msg); err != nil {
			g.log.Error("failed to add transaction to pool")
			return
		}

		// add transaction to the history cache using the key we created earlier
		g.engine.History().Add(historyKey)
	}

	g.log.Info("Added new transaction to pool", "TxID", msg.ID())
}

// RelayTx relays transactions to peers
func (g *Gossip) RelayTx(tx *wire.Transaction, remotePeers []types.Engine) error {

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
