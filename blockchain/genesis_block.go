package blockchain

import (
	"math/big"

	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/wire"
)

var GenesisBlock core.Block = &wire.Block{
	Header: &wire.Header{
		Number: 0x0000000000000001,
		Nonce: core.BlockNonce{
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
		},
		Timestamp:     1533829455,
		CreatorPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
		ParentHash: util.Hash{
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		},
		StateRoot: util.Hash{
			0x9c, 0xe7, 0xf0, 0xee, 0x55, 0xcd, 0xe7, 0x7e, 0x6b, 0x51, 0x97, 0xeb, 0x3b, 0x67, 0xd5, 0xfe,
			0xf4, 0x28, 0xb2, 0x5e, 0x59, 0xed, 0xf9, 0xcd, 0xf0, 0x41, 0x94, 0x95, 0x08, 0x96, 0x05, 0x27,
		},
		TransactionsRoot: util.Hash{
			0x98, 0x82, 0x8e, 0xf7, 0x77, 0xd0, 0x75, 0xd0, 0x9f, 0xd7, 0x5c, 0x9e, 0xbb, 0xee, 0x3d, 0x24,
			0x00, 0xab, 0x59, 0xd2, 0x17, 0x94, 0xc1, 0xe2, 0xfb, 0x26, 0x57, 0x12, 0x2f, 0x65, 0x0a, 0xa3,
		},
		Difficulty: new(big.Int).SetInt64(0x1f4),
		Extra:      []uint8{},
	},
	Transactions: []*wire.Transaction{
		&wire.Transaction{
			Type:         2,
			Nonce:        123,
			To:           "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad",
			From:         "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad",
			SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
			Value:        "1000000",
			Timestamp:    1532730722,
			Fee:          "0.1",
			InvokeArgs:   (*wire.InvokeArgs)(nil),
			Sig: []uint8{
				0x1b, 0xdf, 0x9b, 0xec, 0x45, 0xe1, 0x2e, 0x23, 0x92, 0x4f, 0x61, 0x50, 0x66, 0x56, 0x0d, 0xb1,
				0x2c, 0x7c, 0xe0, 0x86, 0xc8, 0xef, 0x1d, 0x69, 0xd4, 0x5c, 0xaa, 0x31, 0x81, 0xb7, 0xc0, 0x84,
				0xc2, 0x96, 0x65, 0x39, 0xc9, 0xc9, 0x26, 0xea, 0x91, 0x9e, 0xde, 0x22, 0xca, 0x1e, 0x4e, 0xce,
				0x4e, 0xe9, 0x92, 0x1e, 0x03, 0xee, 0x4a, 0x7e, 0xb0, 0x0a, 0xca, 0xd1, 0x23, 0xc0, 0xa1, 0x0c,
			},
			Hash: util.Hash{
				0x5c, 0x28, 0x8f, 0x53, 0x95, 0xd8, 0xc7, 0xd6, 0x85, 0xf8, 0xe7, 0xa5, 0x95, 0x83, 0x87, 0x3b,
				0x56, 0x7a, 0xad, 0xf9, 0x28, 0x2f, 0x37, 0x01, 0xa6, 0xee, 0x34, 0x53, 0x4b, 0x67, 0x10, 0x27,
			},
		},
	},
	Hash: util.Hash{
		0x3d, 0x53, 0x59, 0xa6, 0x95, 0x1f, 0xdd, 0x95, 0x1a, 0xa8, 0xaf, 0x77, 0x98, 0x8d, 0xde, 0x4d,
		0xb4, 0xa9, 0x28, 0xa8, 0xfe, 0xbe, 0xd4, 0x06, 0x55, 0xda, 0xb6, 0x71, 0xd9, 0xf3, 0x1b, 0xc8,
	},
	Sig: []uint8{
		0x66, 0xce, 0x01, 0x12, 0x60, 0x39, 0x31, 0xf1, 0x77, 0x51, 0x08, 0x6f, 0x0b, 0x29, 0x69, 0x7b,
		0x4a, 0xdc, 0x49, 0x89, 0x12, 0x09, 0x80, 0x94, 0x7b, 0x4a, 0x88, 0x69, 0x42, 0x54, 0x17, 0x38,
		0x69, 0x87, 0x08, 0x46, 0x36, 0xd1, 0xe7, 0x6a, 0xf1, 0xfe, 0x2c, 0x63, 0xd5, 0xf7, 0x8c, 0x9c,
		0xa0, 0xee, 0xda, 0x12, 0xca, 0x17, 0xb8, 0x9b, 0xd2, 0xab, 0xf1, 0xa3, 0x22, 0xa3, 0x1c, 0x01,
	},
}
