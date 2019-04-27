package gossip

import (
	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/node/common"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/util/cache"
	net "github.com/libp2p/go-libp2p-net"
)

// OnTx handles incoming transaction message
func (g *Manager) OnTx(s net.Stream, rp core.Engine) error {

	var txID string
	tx := &core.Transaction{}
	var hk []interface{}

	msg := &core.TxInfo{}
	if err := ReadStream(s, msg); err != nil {
		s.Reset()
		return g.logErr(err, rp, "[OnTx] Failed to read TxInfo message")
	}

	// We can't accept a transaction that already
	// exists in the transaction the pool.
	if g.engine.GetTxPool().HasByHash(msg.Hash.HexStr()) {
		goto tx_not_ok
	}

	// We can't accept a transaction that already
	// exists on the main chain.
	if existingTx, _ := g.engine.GetBlockchain().
		GetTransaction(msg.Hash); existingTx != nil {
		goto tx_not_ok
	}

	// Send back TxOk message indicating readiness
	// to receive the transaction
	if err := WriteStream(s, &core.TxOk{Ok: true}); err != nil {
		s.Reset()
		return g.logErr(err, rp, "[OnTx] Failed to write TxOk message")
	}

	// At this point, we expect the peer to send the transaction
	if err := ReadStream(s, tx); err != nil {
		s.Reset()
		return g.logErr(err, rp, "[OnTx] Failed to read")
	}

	txID = util.String(tx.GetID()).SS()
	g.log.Info("Received a new transaction", "PeerID", rp.ShortID(), "TxID", txID)
	go g.engine.GetEventEmitter().Emit(core.EventTransactionReceived, tx)

	// Keep a record of us receiving this transaction,
	// so that we won't rebroadcast it to the sender
	hk = common.KeyTx(tx, rp)
	g.engine.GetHistory().AddMulti(cache.Sec(600), hk...)

tx_not_ok:
	if err := WriteStream(s, &core.TxOk{Ok: false}); err != nil {
		s.Reset()
		return g.logErr(err, rp, "[OnTx] Failed to write TxOk message")
	}

	return s.Close()
}

// BroadcastTx broadcast transactions to selected peers
func (g *Manager) BroadcastTx(tx types.Transaction, remotePeers []core.Engine) error {

	txID := util.String(tx.GetID()).SS()
	sent := 0
	g.log.Debug("Attempting to broadcast a transaction",
		"TxID", txID,
		"NumPeers", len(remotePeers))

	broadcastPeers := g.PickBroadcastersFromPeers(g.broadcasters, remotePeers, 3)
	for _, peer := range broadcastPeers.Peers() {

		// We need to remove the broadcast peer
		// if it is no longer connected
		if !peer.Connected() {
			broadcastPeers.Remove(peer)
			continue
		}

		// Check if we have an history of receiving
		// this transaction from this peer recently. If yes,
		// we will not proceed further
		hk := common.KeyTx(tx, peer)
		if g.engine.GetHistory().HasMulti(hk...) {
			continue
		}

		s, c, err := g.NewStream(peer, config.GetVersions().Tx)
		if err != nil {
			g.logConnectErr(err, peer, "[BroadcastTx] Failed to connect")
			continue
		}
		defer c()

		// Send a message describing the transaction.
		// If the peer accepts the transaction, we can send the full tx.
		txInfo := core.TxInfo{Hash: tx.GetHash()}
		if err := WriteStream(s, txInfo); err != nil {
			s.Reset()
			g.logErr(err, peer, "[BroadcastTx] Failed to write to stream")
			continue
		}

		// Read TxOk message to know whether to send the transaction
		txOk := &core.TxOk{}
		if err := ReadStream(s, txOk); err != nil {
			s.Reset()
			g.logErr(err, peer, "[BroadcastTx] Failed to read")
			continue
		}

		if !txOk.Ok {
			s.Close()
			g.log.Debug("Peer rejected our intent to broadcast a transaction",
				"PeerID", peer.ShortID(),
				"TxID", txID)
			continue
		}

		// At this point, we can send the transaction to the peer
		if err := WriteStream(s, tx); err != nil {
			s.Reset()
			g.logErr(err, peer, "[BroadcastTx] Failed to write to stream")
			continue
		}

		sent++

		s.Close()
	}

	g.log.Info("Transaction successfully broadcast",
		"TxID", txID,
		"NumPeersSentTo", sent)

	return nil
}
