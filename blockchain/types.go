package blockchain

import "github.com/ellcrys/elld/wire"

// DocType represents a document type
type DocType int

const (
	// TypeBlock represents a block document type
	TypeBlock DocType = 0x1

	// TypeTx represents a transaction document type
	TypeTx DocType = 0x2
)

// Block represents a block document
type Block struct {
	DocType      DocType           `json:"docType"`
	Header       *Header           `json:"header"`
	Transactions []*Transaction    `json:"transactions"`
	Votes        []*wire.BlockVote `json:"votes"`
}

// Header represents an header document
type Header struct {
	ParentHash string `json:"parentHash"`
	Number     int64  `json:"number"`
	Root       string `json:"root"`
	TxHash     string `json:"txHash"`
	Nonce      uint64 `json:"nonce"`
	MixHash    string `json:"mixHash"`
	Difficulty string `json:"difficulty"`
	Timestamp  int64  `json:"timestamp"`
}

// Transaction represents a transaction document
type Transaction struct {
	Type         int    `json:"type"`
	Nonce        int64  `json:"nonce"`
	To           string `json:"to"`
	SenderPubKey string `json:"senderPubKey"`
	Value        string `json:"value"`
	Timestamp    int64  `json:"timestamp"`
	Fee          string `json:"fee"`
	Signature    string `json:"signature"`
	Hash         string `json:"hash"`
}

// GetNumber returns the number
func (b *Block) GetNumber() int64 {
	return b.Header.Number
}
