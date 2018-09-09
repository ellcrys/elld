package blockchain

import (
	"math/big"

	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/types/core/objects"
	"github.com/ellcrys/elld/util"
)

var GenesisBlock core.Block = &objects.Block{
	Header: &objects.Header{
		Number: 0x0000000000000001,
		Nonce: core.BlockNonce{
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
		},
		Timestamp:     1532730722,
		CreatorPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
		ParentHash: util.Hash{
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		},
		StateRoot: util.Hash{
			0x6c, 0x57, 0x3e, 0x52, 0xb4, 0x2a, 0x37, 0x75, 0x35, 0xe4, 0x75, 0xaa, 0x93, 0x4d, 0x16, 0x97,
			0x56, 0x1e, 0x47, 0x4e, 0x7d, 0x0b, 0x5a, 0xc0, 0x7c, 0xf2, 0xf7, 0x22, 0xc6, 0x66, 0x88, 0x14,
		},
		TransactionsRoot: util.Hash{
			0xc4, 0xe9, 0x9d, 0xd6, 0xa6, 0x62, 0x32, 0x4c, 0xd7, 0xea, 0x5e, 0x8b, 0x4d, 0x9d, 0x24, 0x37,
			0x52, 0xc8, 0x45, 0xaf, 0xbb, 0xfa, 0x36, 0x9a, 0x01, 0xf9, 0xd0, 0xfa, 0x6f, 0x15, 0x3a, 0xe6,
		},
		Difficulty:      new(big.Int).SetInt64(0x20000),
		TotalDifficulty: new(big.Int).SetInt64(0x20000),
		Extra:           nil,
	},
	Transactions: []*objects.Transaction{
		&objects.Transaction{
			Type:         2,
			Nonce:        0x0000000000000001,
			To:           "e6i7rxApBYUt7w94gGDKTz45A5J567JfkS",
			From:         "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad",
			SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
			Value:        "1",
			Timestamp:    1532730722,
			Fee:          "2.4",
			InvokeArgs:   (*objects.InvokeArgs)(nil),
			Sig: []uint8{
				0x20, 0x95, 0x5d, 0x23, 0xf3, 0x90, 0x3f, 0x70, 0x3a, 0x08, 0xfb, 0xf0, 0x5e, 0xb6, 0x1e, 0x1f,
				0xf2, 0xa0, 0x06, 0x61, 0x7a, 0xf6, 0xc9, 0x7d, 0xef, 0x8b, 0x2c, 0xf3, 0x0e, 0x8d, 0xf4, 0x0d,
				0x91, 0xb7, 0xdc, 0x15, 0x9c, 0x67, 0x05, 0x1b, 0x89, 0x76, 0xc8, 0x7a, 0x43, 0xa4, 0x91, 0x70,
				0x38, 0xb8, 0x25, 0x01, 0x3f, 0xaf, 0xcd, 0x8a, 0xb2, 0x2e, 0x20, 0x9c, 0xda, 0x1c, 0x3e, 0x02,
			},
			Hash: util.Hash{
				0x81, 0x2f, 0x48, 0x02, 0xed, 0xa8, 0xa2, 0xc7, 0x12, 0xcb, 0xa2, 0x0b, 0x8e, 0xeb, 0xf6, 0x47,
				0x17, 0xf3, 0xd0, 0xa0, 0xfc, 0x62, 0x63, 0xd0, 0x1b, 0xde, 0xe4, 0xa1, 0x9f, 0xd4, 0x38, 0x63,
			},
		},
	},
	Hash: util.Hash{
		0x3a, 0x03, 0xde, 0x27, 0x6f, 0xce, 0xb4, 0xae, 0x67, 0x37, 0x6c, 0xf6, 0x2d, 0xdd, 0x05, 0x67,
		0xfc, 0xb4, 0xf6, 0xc6, 0xa5, 0x01, 0xdb, 0x07, 0xf4, 0x0d, 0x2e, 0xae, 0x65, 0x70, 0x19, 0xb8,
	},
	Sig: []uint8{
		0x8c, 0x02, 0x9f, 0xb4, 0xa8, 0xb1, 0xe6, 0x6a, 0x9f, 0x6f, 0x78, 0xc5, 0x0a, 0x64, 0xfb, 0xfa,
		0x28, 0x46, 0x74, 0xb3, 0x44, 0x0f, 0xce, 0x00, 0xb7, 0x14, 0x3c, 0xdc, 0xa4, 0x9e, 0xf6, 0xe1,
		0x09, 0xe4, 0x51, 0xcc, 0x4b, 0xe1, 0x9e, 0x67, 0xd0, 0x4e, 0xcb, 0x3b, 0x53, 0x19, 0x74, 0xaf,
		0xd2, 0x2a, 0xb3, 0x06, 0x41, 0xd7, 0xeb, 0xc3, 0x13, 0x6b, 0x27, 0x30, 0xf3, 0x56, 0xa9, 0x00,
	},
	ChainReader: nil,
	Broadcaster: nil,
}
