package gossip

import (
	"fmt"

	"github.com/ellcrys/elld/blockchain"
	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/util/cache"
	net "github.com/libp2p/go-libp2p-net"
)

// MakeTxHistoryKey creates an GetHistory() cache key
// for recording a received/sent transaction
func MakeTxHistoryKey(tx types.Transaction, peer core.Engine) []interface{} {
	return []interface{}{tx.GetID(), peer.StringID()}
}

// OnTx handles incoming transaction message
func (g *Gossip) OnTx(s net.Stream, rp core.Engine) error {
	defer s.Close()

	msg := &core.Transaction{}
	if err := ReadStream(s, msg); err != nil {
		s.Reset()
		return g.logErr(err, rp, "[OnTx] Failed to read")
	}

	g.log.Info("Received new transaction", "PeerID", rp.ShortID())

	historyKey := MakeTxHistoryKey(msg, rp)
	if g.engine.GetHistory().HasMulti(historyKey...) {
		return nil
	}

	// TxTypeAlloc transactions are not to
	// be relayed like standard transactions.
	if msg.Type == core.TxTypeAlloc {
		s.Reset()
		err := fmt.Errorf("Allocation transaction type is not allowed")
		g.log.Debug(err.Error())
		g.engine.GetEventEmitter().Emit(EventTransactionProcessed, err)
		return err
	}

	// Ignore the transaction if already
	// in our transaction pool
	if g.engine.GetTxPool().Has(msg) {
		return nil
	}

	// Validate the transaction and check whether
	// it already exists in the transactions pool,
	// main chain and side chains. If so, reject it
	errs := blockchain.NewTxValidator(msg, g.engine.GetTxPool(),
		g.engine.GetBlockchain()).Validate()
	if len(errs) > 0 {
		s.Reset()
		g.log.Debug("Transaction is not valid", "Err", errs[0])
		g.engine.GetEventEmitter().Emit(EventTransactionProcessed, errs[0])
		return errs[0]
	}

	// Add the transaction to the transaction
	// pool and wait for error response
	if err := g.engine.AddTransaction(msg); err != nil {
		g.log.Error("Failed to add transaction to pool", "Err", msg)
		g.engine.GetEventEmitter().Emit(EventTransactionProcessed, err)
		return err
	}

	g.engine.GetHistory().AddMulti(cache.Sec(600), historyKey...)
	g.engine.GetEventEmitter().Emit(EventTransactionProcessed)
	g.log.Info("Added new transaction to pool", "TxID", msg.GetID())

	return nil
}

// RelayTx relays transactions to peers
func (g *Gossip) RelayTx(tx types.Transaction, remotePeers []core.Engine) error {

	txID := util.String(tx.GetID()).SS()
	sent := 0
	g.log.Debug("Relaying transaction to peers", "TxID", txID, "NumPeers",
		len(remotePeers))

	for _, peer := range remotePeers {

		historyKey := MakeTxHistoryKey(tx, peer)

		if g.engine.GetHistory().HasMulti(historyKey...) {
			continue
		}

		s, c, err := g.NewStream(peer, config.TxVersion)
		if err != nil {
			g.logConnectErr(err, peer, "[RelayTx] Failed to connect")
			continue
		}
		defer c()
		defer s.Close()

		if err := WriteStream(s, tx); err != nil {
			s.Reset()
			g.log.Debug("Tx message failed. failed to write to stream",
				"Err", err, "PeerID", peer.ShortID())
			continue
		}

		g.engine.GetHistory().AddMulti(cache.Sec(600), historyKey...)

		sent++
	}

	g.log.Info("Finished relaying transaction", "TxID", txID, "NumPeersSentTo", sent)
	return nil
}
