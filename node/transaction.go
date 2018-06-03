package node

import (
	"bufio"
	"context"

	"github.com/ellcrys/druid/node/histcache"
	"github.com/ellcrys/druid/util"
	"github.com/ellcrys/druid/wire"
	net "github.com/libp2p/go-libp2p-net"
	pc "github.com/multiformats/go-multicodec/protobuf"
)

func makeTxHistoryKey(tx *wire.Transaction, peer *Node) histcache.MultiKey {
	return []interface{}{tx.ID(), peer.StringID()}
}

// OnTx handles incoming transaction message
func (pt *Inception) OnTx(s net.Stream) {
	defer s.Close()
	remotePeer := NewRemoteNode(util.FullRemoteAddressFromStream(s), pt.LocalPeer())
	remotePeerIDShort := remotePeer.ShortID()

	pt.log.Info("Received new transaction", "PeerID", remotePeerIDShort)

	msg := &wire.Transaction{}
	if err := pc.Multicodec(nil).Decoder(bufio.NewReader(s)).Decode(msg); err != nil {
		s.Reset()
		pt.log.Error("failed to read tx message", "Err", err, "PeerID", remotePeerIDShort)
		return
	}

	historyKey := makeTxHistoryKey(msg, remotePeer)

	// check if we have a history about this transaction with the remote peer,
	// if no, add the transaction.
	if !pt.LocalPeer().History().Has(historyKey) {

		if err := pt.LocalPeer().ActionAddTx(msg); err != nil {
			s.Reset()
			pt.log.Error("failed to add transaction to pool", "Err", err)
			return
		}

		// add new history
		pt.LocalPeer().History().Add(historyKey)
	}

	pt.log.Info("Added new transaction to pool", "TxID", msg.ID())
}

// RelayTx relays transactions to peers
func (pt *Inception) RelayTx(tx *wire.Transaction, remotePeers []*Node) error {

	txID := tx.ID()
	pt.log.Debug("Relaying transaction to peers", "TxID", txID, "NumPeers", len(remotePeers))
	sent := 0
	for _, peer := range remotePeers {

		historyKey := makeTxHistoryKey(tx, peer)

		// check if we have a history of transaction with this remote peer,
		// if yes, do not relay
		if pt.LocalPeer().History().Has(historyKey) {
			continue
		}

		s, err := pt.LocalPeer().addToPeerStore(peer).newStream(context.Background(), peer.ID(), util.TxVersion)
		if err != nil {
			pt.log.Debug("Tx message failed. failed to connect to peer", "Err", err, "PeerID", peer.ShortID())
			continue
		}

		w := bufio.NewWriter(s)
		if err := pc.Multicodec(nil).Encoder(w).Encode(tx); err != nil {
			s.Reset()
			pt.log.Debug("Tx message failed. failed to write to stream", "Err", err, "PeerID", peer.ShortID())
			continue
		}

		// add new history
		pt.LocalPeer().History().Add(historyKey)

		w.Flush()
		s.Close()
		sent++
	}

	pt.log.Info("Finished relaying transaction", "TxID", txID, "NumPeersSentTo", sent)

	return nil
}
