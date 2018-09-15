package blockchain

import (
	"math/big"

	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/types/core/objects"
	"github.com/ellcrys/elld/util"
)

// GenesisBlock is the first block of
// the block chain from which all other future
// blocks are rooted from. It contains the initial
// state and takes the block number 1.
var GenesisBlock core.Block = &objects.Block{
	Header: &objects.Header{
		Number: 0x0000000000000001,
		Nonce: core.BlockNonce{
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
		},
		Timestamp:     1537008210,
		CreatorPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
		ParentHash: util.Hash{
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		},
		StateRoot: util.Hash{
			0xff, 0xe9, 0xf7, 0x29, 0x8b, 0x00, 0xa5, 0x4a, 0x31, 0x0a, 0xf7, 0x92, 0xc1, 0x02, 0x75, 0x3f,
			0xac, 0x37, 0x3a, 0xcf, 0x3d, 0x4c, 0x04, 0xbe, 0x75, 0xf2, 0xa6, 0xc4, 0xa0, 0xdb, 0x07, 0x7b,
		},
		TransactionsRoot: util.Hash{
			0xa1, 0x06, 0x14, 0xc4, 0x78, 0xc6, 0x74, 0xe7, 0xd4, 0x3c, 0x88, 0x8f, 0x49, 0x41, 0x58, 0x70,
			0x35, 0xb7, 0x1a, 0xcf, 0xa2, 0xa1, 0x0d, 0x9c, 0x17, 0xb1, 0x63, 0xb0, 0xa7, 0x75, 0x83, 0x0b,
		},
		Difficulty:      new(big.Int).SetInt64(0x20000),
		TotalDifficulty: new(big.Int).SetInt64(0x20000),
		Extra:           nil,
	},
	Transactions: []*objects.Transaction{
		&objects.Transaction{
			Type:         2,
			Nonce:        0x0000000000000001,
			To:           "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad",
			From:         "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad",
			SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
			Value:        "1",
			Timestamp:    1532730722,
			Fee:          "5",
			InvokeArgs:   (*objects.InvokeArgs)(nil),
			Sig: []uint8{
				0x81, 0x53, 0x33, 0x24, 0xfb, 0x01, 0xd3, 0xdb, 0xd8, 0xc8, 0x15, 0x1b, 0x85, 0xf2, 0xbf, 0x9f,
				0x9b, 0x64, 0xc6, 0xdf, 0x67, 0x47, 0x54, 0x80, 0x88, 0xee, 0xed, 0x40, 0xab, 0xe3, 0x2d, 0x30,
				0xe6, 0xe6, 0x8b, 0x3d, 0x0f, 0x33, 0xda, 0xa9, 0xf5, 0x56, 0xf0, 0xe4, 0xcd, 0x14, 0x77, 0xf9,
				0x41, 0x25, 0x34, 0x72, 0xa0, 0xd6, 0x7c, 0x3d, 0x3b, 0xdf, 0x2a, 0x7b, 0xae, 0x47, 0x70, 0x04,
			},
			Hash: util.Hash{
				0x49, 0x79, 0x47, 0x66, 0x8c, 0x11, 0x88, 0x04, 0x95, 0xc8, 0x09, 0xc1, 0x12, 0x7a, 0xaa, 0xf8,
				0xa8, 0x75, 0x0f, 0x85, 0xe3, 0xc2, 0xf3, 0x12, 0x3d, 0x2e, 0xfe, 0xf6, 0x53, 0x42, 0x61, 0xa9,
			},
		},
	},
	Hash: util.Hash{
		0x19, 0xc0, 0xf2, 0xf3, 0x27, 0xdb, 0x01, 0xa5, 0xd6, 0xca, 0x4d, 0xaa, 0xc9, 0x72, 0xb3, 0x4b,
		0x44, 0xe0, 0x6f, 0x6e, 0x5b, 0x3f, 0xf2, 0xdc, 0x04, 0x87, 0x99, 0x5c, 0x21, 0x6f, 0xc6, 0x49,
	},
	Sig: []uint8{
		0x26, 0x89, 0x08, 0x46, 0x5e, 0x98, 0x5b, 0x3b, 0x98, 0x99, 0x63, 0x51, 0xb0, 0x74, 0x51, 0xd6,
		0x61, 0x59, 0x79, 0x06, 0x51, 0xb7, 0xf9, 0xae, 0x23, 0xa8, 0xd4, 0x20, 0x03, 0x7f, 0x37, 0xea,
		0x8d, 0x9e, 0xbf, 0xd3, 0xdd, 0xe7, 0x29, 0xd3, 0x12, 0xac, 0x21, 0xd9, 0x97, 0x6b, 0xdf, 0x98,
		0x91, 0x5b, 0xec, 0x25, 0xdc, 0x33, 0x7b, 0xb0, 0x82, 0xd7, 0x44, 0xf6, 0x77, 0xf6, 0xed, 0x0a,
	},
	ChainReader: nil,
	Broadcaster: nil,
}
