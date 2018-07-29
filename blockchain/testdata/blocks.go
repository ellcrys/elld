package testdata

import "github.com/ellcrys/elld/wire"

var GenesisBlock = &wire.Block{
	Header: &wire.Header{ParentHash: "", CreatorPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC", Number: 0x1, StateRoot: "0x03a82f13db319a687882a5e52d809b4cdd4271e7975ba6882d458d9215644fe8", TransactionsRoot: "0xe15bb15726de09feecb364c9b66865a0a95ca9213c46a805df719eabdcce1db7", Nonce: 0x5def4, MixHash: "0x1cdf0e214bcdb7af36885316506f7388f262f7b710a28a00d21706550cdd72c2", Difficulty: "102994", Timestamp: 1532816862},
	Transactions: []*wire.Transaction{
		&wire.Transaction{
			Type:         1,
			Nonce:        123,
			To:           "e6i7rxApBYUt7w94gGDKTz45A5J567JfkS",
			From:         "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad",
			SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
			Value:        "1",
			Timestamp:    1532730722,
			Fee:          "0.1",
			InvokeArgs:   (*wire.InvokeArgs)(nil),
			Sig:          "0x5512603b23fc924cffaa907970c43e1943af00df5c1216755bf012f87b013c070eccf8d2d5fff34b3570f9d74ec159526621fb93adcd836e6c67d0cec0519e03",
			Hash:         "0x1c58e122238d0a715a3ead4747baf03a81b4aa4dc1da933fba376fe277d8f0ca",
		},
	},
	Hash: "0x4d2a36390fd9423c141770f14c96ece0d03cd46137f8d8aceb0bf024992c8df9",
	Sig:  "0x1b780e791e52b8dbbac94883187e5ee4189c566c45c0a8e8fd9c0eeca62fceca39ff13af020b995603046249f4c534998198965121389040529ec8cb26f90601",
}

var Block2 = &wire.Block{
	Header: &wire.Header{ParentHash: "0x4d2a36390fd9423c141770f14c96ece0d03cd46137f8d8aceb0bf024992c8df9", CreatorPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC", Number: 0x2, StateRoot: "0x2fbf11beeb8d78b2d81456a730e8a57b0503c1182f88c135e43d5d4ac02258bd", TransactionsRoot: "0xe15bb15726de09feecb364c9b66865a0a95ca9213c46a805df719eabdcce1db7", Nonce: 0x5def4, MixHash: "0x1cdf0e214bcdb7af36885316506f7388f262f7b710a28a00d21706550cdd72c2", Difficulty: "102994", Timestamp: 1532816936},
	Transactions: []*wire.Transaction{
		&wire.Transaction{
			Type:         1,
			Nonce:        123,
			To:           "e6i7rxApBYUt7w94gGDKTz45A5J567JfkS",
			From:         "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad",
			SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
			Value:        "1",
			Timestamp:    1532730722,
			Fee:          "0.1",
			InvokeArgs:   (*wire.InvokeArgs)(nil),
			Sig:          "0x5512603b23fc924cffaa907970c43e1943af00df5c1216755bf012f87b013c070eccf8d2d5fff34b3570f9d74ec159526621fb93adcd836e6c67d0cec0519e03",
			Hash:         "0x1c58e122238d0a715a3ead4747baf03a81b4aa4dc1da933fba376fe277d8f0ca",
		},
	},
	Hash: "0x62f69b213154438fbdb38e30c22fb6f40672db2ca257504361bc0f08bf0c8fe7",
	Sig:  "0x6d6fdd5539ff085d2d4ddd75aeb1b93eba6ee4bf53d454db75df372ffaa3bcd32053fe38aa9db56a6cb855560d4475118213bf728cd3a0a8d2e3a0f82e960608",
}

var ProcessTransaction = []*wire.Block{
	&wire.Block{
		Transactions: []*wire.Transaction{
			&wire.Transaction{
				Type:         1,
				Nonce:        123,
				To:           "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad",
				From:         "e6i7rxApBYUt7w94gGDKTz45A5J567JfkS",
				SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
				Value:        "1",
				Timestamp:    1532730722,
				Fee:          "0.1",
				InvokeArgs:   (*wire.InvokeArgs)(nil),
				Sig:          "0x5512603b23fc924cffaa907970c43e1943af00df5c1216755bf012f87b013c070eccf8d2d5fff34b3570f9d74ec159526621fb93adcd836e6c67d0cec0519e03",
				Hash:         "0x1c58e122238d0a715a3ead4747baf03a81b4aa4dc1da933fba376fe277d8f0ca",
			},
		},
	},
	&wire.Block{
		Header: &wire.Header{ParentHash: "0x4d2a36390fd9423c141770f14c96ece0d03cd46137f8d8aceb0bf024992c8df9", CreatorPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC", Number: 0x2, StateRoot: "0x2fbf11beeb8d78b2d81456a730e8a57b0503c1182f88c135e43d5d4ac02258bd", TransactionsRoot: "0xe15bb15726de09feecb364c9b66865a0a95ca9213c46a805df719eabdcce1db7", Nonce: 0x5def8, MixHash: "0x1cdf0e214bcdb7af36885316506f7388f262f7b710a28a00d21706550cdd72c2", Difficulty: "102994", Timestamp: 1532817643},
		Transactions: []*wire.Transaction{
			&wire.Transaction{
				Type:         1,
				Nonce:        123,
				To:           "e6i7rxApBYUt7w94gGDKTz45A5J567JfkS",
				From:         "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad",
				SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
				Value:        "100000",
				Timestamp:    1532730722,
				Fee:          "0.1",
				InvokeArgs:   (*wire.InvokeArgs)(nil),
				Sig:          "0x5512603b23fc924cffaa907970c43e1943af00df5c1216755bf012f87b013c070eccf8d2d5fff34b3570f9d74ec159526621fb93adcd836e6c67d0cec0519e03",
				Hash:         "0x1c58e122238d0a715a3ead4747baf03a81b4aa4dc1da933fba376fe277d8f0ca",
			},
		},
		Hash: "0x187d4bb9123c6cc6a2d19979e5aba60ad844707423b36bde5e140d6a07807729",
		Sig:  "0x237428cbfe1691b379a3b32aad980ae5ade1393794cc74501bb9f43d6192146ac32cb3bf211e1a40a454c743eabcfbb4987a2f639bf61e957019d5ccbf614309",
	},
	&wire.Block{
		Transactions: []*wire.Transaction{
			&wire.Transaction{
				Type:         1,
				Nonce:        123,
				To:           "e6i7rxApBYUt7w94gGDKTz45A5J567JfkS",
				From:         "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad",
				SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
				Value:        "100_333",
				Timestamp:    1532730722,
				Fee:          "0.1",
				InvokeArgs:   (*wire.InvokeArgs)(nil),
				Sig:          "0x5512603b23fc924cffaa907970c43e1943af00df5c1216755bf012f87b013c070eccf8d2d5fff34b3570f9d74ec159526621fb93adcd836e6c67d0cec0519e03",
				Hash:         "0x1c58e122238d0a715a3ead4747baf03a81b4aa4dc1da933fba376fe277d8f0ca",
			},
		},
	},
	&wire.Block{
		Transactions: []*wire.Transaction{
			&wire.Transaction{
				Type:         1,
				Nonce:        123,
				To:           "e6i7rxApBYUt7w94gGDKTz45A5J567JfkS",
				From:         "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad",
				SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
				Value:        "1",
				Timestamp:    1532730722,
				Fee:          "0.1",
				InvokeArgs:   (*wire.InvokeArgs)(nil),
				Sig:          "0x5512603b23fc924cffaa907970c43e1943af00df5c1216755bf012f87b013c070eccf8d2d5fff34b3570f9d74ec159526621fb93adcd836e6c67d0cec0519e03",
				Hash:         "0x1c58e122238d0a715a3ead4747baf03a81b4aa4dc1da933fba376fe277d8f0ca",
			},
		},
	},
	&wire.Block{
		Transactions: []*wire.Transaction{
			&wire.Transaction{
				Type:         1,
				Nonce:        123,
				To:           "eQ9TnvMUUsB8ztZchSe3o7f5XfifEmZvJR",
				From:         "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad",
				SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
				Value:        "1",
				Timestamp:    1532730722,
				Fee:          "0.1",
				InvokeArgs:   (*wire.InvokeArgs)(nil),
				Sig:          "0xa3e0edce436fe4ba2dc5df3ab9a41d5be1d399da8b5a4e1a54ea19b7b528eb4fb279c26d2a57151e4461e7a6032333a37f23bc3ed1e3cd794786a1fc94fc4508",
				Hash:         "0xf767f0b1a94b03c731752f9dffa50eefed1a1edcc153cf25b2b6fc7c02f50fa6",
			},
		},
	},
	&wire.Block{
		Transactions: []*wire.Transaction{
			&wire.Transaction{
				Type:         1,
				Nonce:        123,
				To:           "e6i7rxApBYUt7w94gGDKTz45A5J567JfkS",
				From:         "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad",
				SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
				Value:        "1",
				Timestamp:    1532730722,
				Fee:          "0.1",
				InvokeArgs:   (*wire.InvokeArgs)(nil),
				Sig:          "0x5512603b23fc924cffaa907970c43e1943af00df5c1216755bf012f87b013c070eccf8d2d5fff34b3570f9d74ec159526621fb93adcd836e6c67d0cec0519e03",
				Hash:         "0x1c58e122238d0a715a3ead4747baf03a81b4aa4dc1da933fba376fe277d8f0ca",
			},
			&wire.Transaction{
				Type:         1,
				Nonce:        1234,
				To:           "e6i7rxApBYUt7w94gGDKTz45A5J567JfkS",
				From:         "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad",
				SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
				Value:        "1",
				Timestamp:    1532730723,
				Fee:          "0.1",
				InvokeArgs:   (*wire.InvokeArgs)(nil),
				Sig:          "0x1bceba944de603a980f3814c9e1e5fd2a7db79cfe2c8490abb35d3bc4748d48578199cae58b7ea790951e6f15aa81f2a5d8cb1e70df4ccde057f5c4c6b97ea07",
				Hash:         "0xf43c4c4bd7c917cfc9680201951ec74c27104721122ea7d743e1c1b4f9a69d31",
			},
		},
	},
	&wire.Block{
		Transactions: []*wire.Transaction{
			&wire.Transaction{
				Type:         1,
				Nonce:        1234,
				To:           "eQ9TnvMUUsB8ztZchSe3o7f5XfifEmZvJR",
				From:         "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad",
				SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
				Value:        "1",
				Timestamp:    1532730723,
				Fee:          "0.1",
				InvokeArgs:   (*wire.InvokeArgs)(nil),
				Sig:          "0xd8c9f3c01b7b1fc56a69a22ec21e939990de6911f75ff98fe67b29677931d12172a3f436469e32e2e4511aae65d964756cc6bb4fe1653fce16d867f00b83c40d",
				Hash:         "0xc90eb0d12fe7d98a3c7d7eea54529fdc08b4e704eb093243df4a88999ed93a0e",
			},
			&wire.Transaction{
				Type:         1,
				Nonce:        123,
				To:           "eQ9TnvMUUsB8ztZchSe3o7f5XfifEmZvJR",
				From:         "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad",
				SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
				Value:        "1",
				Timestamp:    1532730722,
				Fee:          "0.1",
				InvokeArgs:   (*wire.InvokeArgs)(nil),
				Sig:          "0xa3e0edce436fe4ba2dc5df3ab9a41d5be1d399da8b5a4e1a54ea19b7b528eb4fb279c26d2a57151e4461e7a6032333a37f23bc3ed1e3cd794786a1fc94fc4508",
				Hash:         "0xf767f0b1a94b03c731752f9dffa50eefed1a1edcc153cf25b2b6fc7c02f50fa6",
			},
		},
	},
}

var OrphanBlock1 = &wire.Block{
	Header: &wire.Header{ParentHash: "0xsomething1cdf0e214bcdb7af36885316506f7388f262f7b710a28a00d217065", CreatorPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC", Number: 0x2, StateRoot: "0x802d600887c1fba7187c36180cb6e6e9218f94e58e8a56f66561cbf8452f36f9", TransactionsRoot: "0xe15bb15726de09feecb364c9b66865a0a95ca9213c46a805df719eabdcce1db7", Nonce: 0x5df04, MixHash: "0x1cdf0e214bcdb7af36885316506f7388f262f7b710a28a00d21706550cdd72c2", Difficulty: "102994", Timestamp: 1532776721},
	Transactions: []*wire.Transaction{
		&wire.Transaction{
			Type:         1,
			Nonce:        123,
			To:           "e6i7rxApBYUt7w94gGDKTz45A5J567JfkS",
			From:         "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad",
			SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
			Value:        "1",
			Timestamp:    1532730722,
			Fee:          "0.1",
			InvokeArgs:   (*wire.InvokeArgs)(nil),
			Sig:          "0x5512603b23fc924cffaa907970c43e1943af00df5c1216755bf012f87b013c070eccf8d2d5fff34b3570f9d74ec159526621fb93adcd836e6c67d0cec0519e03",
			Hash:         "0x1c58e122238d0a715a3ead4747baf03a81b4aa4dc1da933fba376fe277d8f0ca",
		},
	},
	Hash: "0x6f535131116ad4e64fe2353a329e01c0e1b0de07856a0294d3cc3dad60aac96e",
	Sig:  "0xf8c0af171007592e1f06ba6f8b749d522e7c54067436105631c0c7e4cc9ac7eac198c0bf6bc1429e26456fcf3aef66b5d7620fc8142c46e2c535fd1ea4450f02",
}

var Stale1 = []*wire.Block{
	&wire.Block{
		Header: &wire.Header{ParentHash: "0x4d2a36390fd9423c141770f14c96ece0d03cd46137f8d8aceb0bf024992c8df9", CreatorPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC", Number: 0x2, StateRoot: "0x2fbf11beeb8d78b2d81456a730e8a57b0503c1182f88c135e43d5d4ac02258bd", TransactionsRoot: "0xe15bb15726de09feecb364c9b66865a0a95ca9213c46a805df719eabdcce1db7", Nonce: 0x5def8, MixHash: "0x1cdf0e214bcdb7af36885316506f7388f262f7b710a28a00d21706550cdd72c2", Difficulty: "102994", Timestamp: 1532819725},
		Transactions: []*wire.Transaction{
			&wire.Transaction{
				Type:         1,
				Nonce:        123,
				To:           "e6i7rxApBYUt7w94gGDKTz45A5J567JfkS",
				From:         "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad",
				SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
				Value:        "1",
				Timestamp:    1532730722,
				Fee:          "0.1",
				InvokeArgs:   (*wire.InvokeArgs)(nil),
				Sig:          "0x5512603b23fc924cffaa907970c43e1943af00df5c1216755bf012f87b013c070eccf8d2d5fff34b3570f9d74ec159526621fb93adcd836e6c67d0cec0519e03",
				Hash:         "0x1c58e122238d0a715a3ead4747baf03a81b4aa4dc1da933fba376fe277d8f0ca",
			},
		},
		Hash: "0xf7837d0dde2f6ff3328a98e43c80d2b81b68a0d78faff7a07a16883e83f5c22b",
		Sig:  "0x61633b6145b0b28b2f326ed5ce32bba2c5a7c32a0f98432ea26ae25421d9f43ece6e244701d7b61e32f74af24c9034883d01c82355280f8df9f4c2e689cedd0f",
	},
	&wire.Block{
		Header: &wire.Header{ParentHash: "0x4d2a36390fd9423c141770f14c96ece0d03cd46137f8d8aceb0bf024992c8df9", CreatorPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC", Number: 0x2, StateRoot: "0x2fbf11beeb8d78b2d81456a730e8a57b0503c1182f88c135e43d5d4ac02258bd", TransactionsRoot: "0xe15bb15726de09feecb364c9b66865a0a95ca9213c46a805df719eabdcce1db7", Nonce: 0x5def9, MixHash: "0x1cdf0e214bcdb7af36885316506f7388f262f7b710a28a00d21706550cdd72c2", Difficulty: "102994", Timestamp: 1532819797},
		Transactions: []*wire.Transaction{
			&wire.Transaction{
				Type:         1,
				Nonce:        123,
				To:           "e6i7rxApBYUt7w94gGDKTz45A5J567JfkS",
				From:         "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad",
				SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
				Value:        "1",
				Timestamp:    1532730722,
				Fee:          "0.1",
				InvokeArgs:   (*wire.InvokeArgs)(nil),
				Sig:          "0x5512603b23fc924cffaa907970c43e1943af00df5c1216755bf012f87b013c070eccf8d2d5fff34b3570f9d74ec159526621fb93adcd836e6c67d0cec0519e03",
				Hash:         "0x1c58e122238d0a715a3ead4747baf03a81b4aa4dc1da933fba376fe277d8f0ca",
			},
		},
		Hash: "0x2e62dffeccd7c0d6423b681201f7173d939d2a6d2f6ae714a9d42ffdc9cdf5d3",
		Sig:  "0xd1645f6ae6b089c71ae6b38c1b63e357dee97fbfa516a733b1b10ce9622597a10a75f3e9f0046681404a304bb63b88f2f50ec65430cfda5642754baf1e154c03",
	},
	&wire.Block{
		Header: &wire.Header{ParentHash: "0xf7837d0dde2f6ff3328a98e43c80d2b81b68a0d78faff7a07a16883e83f5c22b", CreatorPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC", Number: 0x3, StateRoot: "0xe1b864f373cd882d955723b4403fc6367dbc2eb20fa9a51c91a87af4c5037e23", TransactionsRoot: "0xe15bb15726de09feecb364c9b66865a0a95ca9213c46a805df719eabdcce1db7", Nonce: 0x5def9, MixHash: "0x1cdf0e214bcdb7af36885316506f7388f262f7b710a28a00d21706550cdd72c2", Difficulty: "102994", Timestamp: 1532819880},
		Transactions: []*wire.Transaction{
			&wire.Transaction{
				Type:         1,
				Nonce:        123,
				To:           "e6i7rxApBYUt7w94gGDKTz45A5J567JfkS",
				From:         "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad",
				SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
				Value:        "1",
				Timestamp:    1532730722,
				Fee:          "0.1",
				InvokeArgs:   (*wire.InvokeArgs)(nil),
				Sig:          "0x5512603b23fc924cffaa907970c43e1943af00df5c1216755bf012f87b013c070eccf8d2d5fff34b3570f9d74ec159526621fb93adcd836e6c67d0cec0519e03",
				Hash:         "0x1c58e122238d0a715a3ead4747baf03a81b4aa4dc1da933fba376fe277d8f0ca",
			},
		},
		Hash: "0xa236aad5d25eed4ebf28bcb9570f8dbdb2ddeec21eead154545c532a3563f5b9",
		Sig:  "0xc07f39ac894486cb680ecf0c193dec65b51a67820321ddf524dda124904b04591a6c8c85cd08afb913c464e0d2c31dba506f3ed47ba8d008a4f3edd9a49c140d",
	},
	&wire.Block{
		Header: &wire.Header{ParentHash: "0xf7837d0dde2f6ff3328a98e43c80d2b81b68a0d78faff7a07a16883e83f5c22b", CreatorPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC", Number: 0x4, StateRoot: "0xbbdd1c9c85d587b2c04771f87a63d0f3a3ffb3fed4eb81e619a3fc895c62bc5e", TransactionsRoot: "0xe15bb15726de09feecb364c9b66865a0a95ca9213c46a805df719eabdcce1db7", Nonce: 0x5def9, MixHash: "0x1cdf0e214bcdb7af36885316506f7388f262f7b710a28a00d21706550cdd72c2", Difficulty: "102994", Timestamp: 1532821762},
		Transactions: []*wire.Transaction{
			&wire.Transaction{
				Type:         1,
				Nonce:        123,
				To:           "e6i7rxApBYUt7w94gGDKTz45A5J567JfkS",
				From:         "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad",
				SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
				Value:        "1",
				Timestamp:    1532730722,
				Fee:          "0.1",
				InvokeArgs:   (*wire.InvokeArgs)(nil),
				Sig:          "0x5512603b23fc924cffaa907970c43e1943af00df5c1216755bf012f87b013c070eccf8d2d5fff34b3570f9d74ec159526621fb93adcd836e6c67d0cec0519e03",
				Hash:         "0x1c58e122238d0a715a3ead4747baf03a81b4aa4dc1da933fba376fe277d8f0ca",
			},
		},
		Hash: "0x5d7c756e14b2b8da611d1feafb1f03e7a0a7de65f191d28c95c62155af6c65c2",
		Sig:  "0x46491b31c0ed58a211e5acf4687da2c2bd88b4c42ab441cf5f44c8e3ba4534e30649f9fd1eba90c712fc2e9797bf921c08af6ddf17c8dc4e81c7bdd09570770b",
	},
}

var StateRoot1 = []*wire.Block{
	&wire.Block{
		Header: &wire.Header{ParentHash: "0x4d2a36390fd9423c141770f14c96ece0d03cd46137f8d8aceb0bf024992c8df9", CreatorPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC", Number: 0x2, StateRoot: "0xabdd1c9c85d587b2c04771f87a63d0f3a3ffb3fed4eb81e619a3fc895c62bc5e", TransactionsRoot: "0xe15bb15726de09feecb364c9b66865a0a95ca9213c46a805df719eabdcce1db7", Nonce: 0x5def9, MixHash: "0x1cdf0e214bcdb7af36885316506f7388f262f7b710a28a00d21706550cdd72c2", Difficulty: "102994", Timestamp: 1532857352},
		Transactions: []*wire.Transaction{
			&wire.Transaction{
				Type:         1,
				Nonce:        123,
				To:           "e6i7rxApBYUt7w94gGDKTz45A5J567JfkS",
				From:         "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad",
				SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
				Value:        "1",
				Timestamp:    1532730722,
				Fee:          "0.1",
				InvokeArgs:   (*wire.InvokeArgs)(nil),
				Sig:          "0x5512603b23fc924cffaa907970c43e1943af00df5c1216755bf012f87b013c070eccf8d2d5fff34b3570f9d74ec159526621fb93adcd836e6c67d0cec0519e03",
				Hash:         "0x1c58e122238d0a715a3ead4747baf03a81b4aa4dc1da933fba376fe277d8f0ca",
			},
		},
		Hash: "0xdb772701a974585bb40c63d23daa66b286df147771c135b65062186f9f92d6f0",
		Sig:  "0x34c237175310f39c04a8f1ca1b91bfe69fee89c43a15ffbd6a1ebae6acf05df376dcc938e815ccb2e043e7e0cd1e2d54c9e38043c32633d40ab491bed284fc09",
	},
	&wire.Block{
		Header: &wire.Header{ParentHash: "0x4d2a36390fd9423c141770f14c96ece0d03cd46137f8d8aceb0bf024992c8df9", CreatorPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC", Number: 0x2, StateRoot: "0x2fbf11beeb8d78b2d81456a730e8a57b0503c1182f88c135e43d5d4ac02258bd", TransactionsRoot: "0xe15bb15726de09feecb364c9b66865a0a95ca9213c46a805df719eabdcce1db7", Nonce: 0x5def9, MixHash: "0x1cdf0e214bcdb7af36885316506f7388f262f7b710a28a00d21706550cdd72c2", Difficulty: "102994", Timestamp: 1532857439},
		Transactions: []*wire.Transaction{
			&wire.Transaction{
				Type:         1,
				Nonce:        123,
				To:           "e6i7rxApBYUt7w94gGDKTz45A5J567JfkS",
				From:         "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad",
				SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
				Value:        "1",
				Timestamp:    1532730722,
				Fee:          "0.1",
				InvokeArgs:   (*wire.InvokeArgs)(nil),
				Sig:          "0x5512603b23fc924cffaa907970c43e1943af00df5c1216755bf012f87b013c070eccf8d2d5fff34b3570f9d74ec159526621fb93adcd836e6c67d0cec0519e03",
				Hash:         "0x1c58e122238d0a715a3ead4747baf03a81b4aa4dc1da933fba376fe277d8f0ca",
			},
		},
		Hash: "0x88fc616535ad8c7b99d73d2bc4534b2f9a2b115110cd90b70cfad002c908753a",
		Sig:  "0xd864a7826346973886a8c3632facbcd0be8a15ccb0b1af216ccc4b9f7e038f1db31e5b4fd670ad47d7d8afb375938fdab992b84408907942fcd709b1dc89a70f",
	},
}

var Orphans = []*wire.Block{
	&wire.Block{
		Header: &wire.Header{ParentHash: "0x4d2a36390fd9423c141770f14c96ece0d03cd46137f8d8aceb0bf024992c8df9", CreatorPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC", Number: 0x2, StateRoot: "0x2fbf11beeb8d78b2d81456a730e8a57b0503c1182f88c135e43d5d4ac02258bd", TransactionsRoot: "0xe15bb15726de09feecb364c9b66865a0a95ca9213c46a805df719eabdcce1db7", Nonce: 0x5def9, MixHash: "0x1cdf0e214bcdb7af36885316506f7388f262f7b710a28a00d21706550cdd72c2", Difficulty: "102994", Timestamp: 1532857834},
		Transactions: []*wire.Transaction{
			&wire.Transaction{
				Type:         1,
				Nonce:        123,
				To:           "e6i7rxApBYUt7w94gGDKTz45A5J567JfkS",
				From:         "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad",
				SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
				Value:        "1",
				Timestamp:    1532730722,
				Fee:          "0.1",
				InvokeArgs:   (*wire.InvokeArgs)(nil),
				Sig:          "0x5512603b23fc924cffaa907970c43e1943af00df5c1216755bf012f87b013c070eccf8d2d5fff34b3570f9d74ec159526621fb93adcd836e6c67d0cec0519e03",
				Hash:         "0x1c58e122238d0a715a3ead4747baf03a81b4aa4dc1da933fba376fe277d8f0ca",
			},
		},
		Hash: "0xe2c352e43275b85426dbd5f9e395d65b6a7fcd2997d5959e3d92e75827043e8c",
		Sig:  "0x0f5576749d60278495e6e52a468808c2a1a334f6dc64fd74eec1bcfc06a3994e657c63d5bda2c754d1065453797eccaa980da9ca3d7fd85a8aa355bbc9720f03",
	},
	&wire.Block{
		Header: &wire.Header{ParentHash: "0xe2c352e43275b85426dbd5f9e395d65b6a7fcd2997d5959e3d92e75827043e8c", CreatorPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC", Number: 0x3, StateRoot: "0x8b0a4568e98ae07868f8864d8548e1c0b818d0b1fb7642ca8957c12dc0b15c3d", TransactionsRoot: "0x8c7de88cbceb30c565db180b1626f58ebc48186c94eb24f15ddd9e097ecbfcbb", Nonce: 0x5defb, MixHash: "0x1cdf0e214bcdb7af36885316506f7388f262f7b710a28a00d21706550cdd72c2", Difficulty: "102994", Timestamp: 1532861125},
		Transactions: []*wire.Transaction{
			&wire.Transaction{
				Type:         1,
				Nonce:        123,
				To:           "e6i7rxApBYUt7w94gGDKTz45A5J567JfkS",
				From:         "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad",
				SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
				Value:        "1",
				Timestamp:    1532730723,
				Fee:          "0.1",
				InvokeArgs:   (*wire.InvokeArgs)(nil),
				Sig:          "0xf1501aa8a0f6538625951b4bc9f57a814460cb8ba2ab03b69aa6d136382e791dac0a23aedbaad2fefa33fb715ffcf854c8665855498485aa65102828e8e01808",
				Hash:         "0x2e7e6f8867526c18896c59f4530d8630cae587c7a3a64fe6adc43aecde353f16",
			},
		},
		Hash: "0xb4e15e6238dbc408ac2bef25786f5ce215798454618e700e869bbb9784fdc3d3",
		Sig:  "0x0d1ec3745badc2b0a2240b76ee8600db6da353d84e25c39c0c7c904ae7337b22f1b1e11f1178324f5d937c90e10fa1e17419e28e7327757e623aacacada4e90e",
	},
	&wire.Block{
		Header: &wire.Header{ParentHash: "0xb4e15e6238dbc408ac2bef25786f5ce215798454618e700e869bbb9784fdc3d3", CreatorPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC", Number: 0x4, StateRoot: "0xa33a388b8bcdd8545daf036ad353e87002935a834c25a18777302f5638e84a1f", TransactionsRoot: "0x26de988c07f6765072c4d12aa3d5cdf79653dec26a513b30728df4deb0a5c216", Nonce: 0x5defb, MixHash: "0x1cdf0e214bcdb7af36885316506f7388f262f7b710a28a00d21706550cdd72c2", Difficulty: "102994", Timestamp: 1532861278},
		Transactions: []*wire.Transaction{
			&wire.Transaction{
				Type:         1,
				Nonce:        123,
				To:           "e6i7rxApBYUt7w94gGDKTz45A5J567JfkS",
				From:         "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad",
				SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
				Value:        "1",
				Timestamp:    1532730724,
				Fee:          "0.1",
				InvokeArgs:   (*wire.InvokeArgs)(nil),
				Sig:          "0x223ec5d3548e92efbe60a5aee24766f37201d50355e36e410383c5d3f44fc60a94043f6d082e9b7735a11486fb7e30df9eb4a739ebc463e978cf25b558e6f500",
				Hash:         "0x605a4799b91cc9ef7e677b2a86ed1c8cdad35c10f155567dfa206b46b5a97251",
			},
		},
		Hash: "0x4bf27ac8452cca81f17079cd49e8314b8c1b0a5d9058c5a5cf6477f365c2ba3f",
		Sig:  "0x83ea6bdb942b1f513798069ad8fa68bbd00f82dfc5431a91af48e1f2e8a065cc7b0f1fc79db6b4ada3ec4cf59331c6edfb8f9d96c5055466f880e787c0380f05",
	},
	&wire.Block{
		Header: &wire.Header{ParentHash: "0x4bf27ac8452cca81f17079cd49e8314b8c1b0a5d9058c5a5cf6477f365c2ba3f", CreatorPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC", Number: 0x5, StateRoot: "0xcd77a64da943de74cb9c71ed4d8844459b353690fb3c032eedd58b24d6c7a4a6", TransactionsRoot: "0xf40b36dae9e32c6c632e1bd667797f1867a1b2333599edb3350628c279672101", Nonce: 0x5defb, MixHash: "0x1cdf0e214bcdb7af36885316506f7388f262f7b710a28a00d21706550cdd72c2", Difficulty: "102994", Timestamp: 1532861346},
		Transactions: []*wire.Transaction{
			&wire.Transaction{
				Type:         1,
				Nonce:        123,
				To:           "e6i7rxApBYUt7w94gGDKTz45A5J567JfkS",
				From:         "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad",
				SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
				Value:        "1",
				Timestamp:    1532730725,
				Fee:          "0.1",
				InvokeArgs:   (*wire.InvokeArgs)(nil),
				Sig:          "0x3dee87954795785daabae738f9c6fd3e1f01b19fb56e9aa776580136d3e90a629ab9d42495acb53f24132b5c89089dd1c07c76ee99948377cbddcbe29b58d90c",
				Hash:         "0xccc1bd0e81cb57f43a719b580180d53293623b2f7de2a04c2e5abc683d720a72",
			},
		},
		Hash: "0x4cd6f638986c8bce1ca6ea113964d87127c64982daff24dd5c03cdbe49209a58",
		Sig:  "0x67ffd86d168e3bc0a29617ca3743d15f19ac2b95e26f6e1f48c9d5f68805530892d3c064003bfd4798a5f63a530a77a850f1004ec920112b5656af74c6559d0b",
	},
}

var TestGenesisBlock = `{
    "header": {
        "parentHash": null,
        "creatorPubKey": "49VzsGezoNjJQxHjekoCQP9CXZUs34CmCY53kGaHyR9rCJQJbJW",
        "number": 1,
        "stateRoot": "0xd66d3a287ca3730d2eeae550f765627c767a6b809cf65db1221879b532534297",
        "transactionsRoot": "0x",
        "nonce": 0,
        "mixHash": "",
        "difficulty": "500000",
        "timestamp": 0
    },
    "transactions": [
        {
            "type": 1,
            "nonce": 2,
            "to": "e6i7rxApBYUt7w94gGDKTz45A5J567JfkS",
            "senderPubKey": "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
            "from": "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad",
            "value": "1",
            "timestamp": 1508673895,
            "fee": "0.1",
            "invokeArgs": null,
            "sig": "0xe9821f4d75acf65cab8ba7df82b161c2cc5fb5acb56cb4c54d8bb0425131106489d1041ed964875d7416cd25baf792a2068dc27e07b91110b3e8469f903a420b",
            "hash": "0xc3ca4c0910a49009a51af2d43d01d5c5cead0f1abd9529301c82dec871383886"
        }
    ],
    "hash": "hash_1",
    "sig": "abc"
}`

var TestBlocks = []string{
	`{
		"header": {
			"parentHash": null,
			"creatorPubKey": "49VzsGezoNjJQxHjekoCQP9CXZUs34CmCY53kGaHyR9rCJQJbJW",
			"number": 1,
			"stateRoot": "abc",
			"transactionsRoot": "jsjhf9e3i9nfi",
			"nonce": 0,
			"mixHash": "",
			"difficulty": "500000",
			"timestamp": 0
		},
		"transactions": [
			{
				"type": 1,
				"nonce": 2,
				"to": "efjshfhh389djn29snmnvuis",
				"senderPubKey": "xsj2909jfhhjskmj99k",
				"from": "",
				"value": "100.333",
				"timestamp": 1508673895,
				"fee": "0.00003",
				"invokeArgs": null,
				"sig": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs",
				"hash": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs"
			},
			{
				"type": 2,
				"nonce": 2,
				"to": null,
				"senderPubKey": "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
				"from": "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad",
				"value": "10",
				"timestamp": 1508673895,
				"fee": "0.00003",
				"invokeArgs": null,
				"sig": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs",
				"hash": "shhfd7387ydhudhsy8ehhhfjsg748hd"
			}
		],
		"hash": "lkssjajsdnaskomcskosks",
		"sig": "abc"
	}`,
	`{
		"header": {
			"parentHash": null,
			"creatorPubKey": "49VzsGezoNjJQxHjekoCQP9CXZUs34CmCY53kGaHyR9rCJQJbJW",
			"number": 2,
			"stateRoot": "abc",
			"transactionsRoot": "jsjhf9e3i9nfi",
			"nonce": 0,
			"mixHash": "",
			"difficulty": "500000",
			"timestamp": 0
		},
		"transactions": [
			{
				"type": 1,
				"nonce": 2,
				"to": "efjshfhh389djn29snmnvuis",
				"senderPubKey": "xsj2909jfhhjskmj99k",
				"from": "",
				"value": "100.333",
				"timestamp": 1508673895,
				"fee": "0.00003",
				"invokeArgs": null,
				"sig": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs",
				"hash": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs"
			},
			{
				"type": 2,
				"nonce": 2,
				"to": null,
				"senderPubKey": "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
				"from": "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad",
				"value": "10",
				"timestamp": 1508673895,
				"fee": "0.00003",
				"invokeArgs": null,
				"sig": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs",
				"hash": "shhfd7387ydhudhsy8ehhhfjsg748hd"
			}
		],
		"hash": "lkssjajsdnaskomcskosks",
		"sig": "abc"
	}`,
	`{
		"header": {
			"parentHash": null,
			"creatorPubKey": "49VzsGezoNjJQxHjekoCQP9CXZUs34CmCY53kGaHyR9rCJQJbJW",
			"number": 3,
			"stateRoot": "abc",
			"transactionsRoot": "jsjhf9e3i9nfi",
			"nonce": 0,
			"mixHash": "",
			"difficulty": "500000",
			"timestamp": 0
		},
		"transactions": [
			{
				"type": 1,
				"nonce": 2,
				"to": "efjshfhh389djn29snmnvuis",
				"senderPubKey": "xsj2909jfhhjskmj99k",
				"from": "",
				"value": "100.333",
				"timestamp": 1508673895,
				"fee": "0.00003",
				"invokeArgs": null,
				"sig": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs",
				"hash": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs"
			},
			{
				"type": 2,
				"nonce": 2,
				"to": null,
				"senderPubKey": "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
				"from": "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad",
				"value": "10",
				"timestamp": 1508673895,
				"fee": "0.00003",
				"invokeArgs": null,
				"sig": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs",
				"hash": "shhfd7387ydhudhsy8ehhhfjsg748hd"
			}
		],
		"hash": "lkssjajsdnaskomcskosks",
		"sig": "abc"
	}`,
}

var TestBlock1 = `{
    "header": {
        "parentHash": null,
        "creatorPubKey": "49VzsGezoNjJQxHjekoCQP9CXZUs34CmCY53kGaHyR9rCJQJbJW",
        "number": 1,
        "stateRoot": "abc",
        "transactionsRoot": "jsjhf9e3i9nfi",
        "nonce": 0,
        "mixHash": "",
        "difficulty": "500000",
        "timestamp": 0
    },
    "transactions": [
        {
            "type": 1,
            "nonce": 2,
            "to": "efjshfhh389djn29snmnvuis",
            "senderPubKey": "xsj2909jfhhjskmj99k",
            "from": "",
            "value": "100.333",
            "timestamp": 1508673895,
            "fee": "0.00003",
            "invokeArgs": null,
            "sig": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs",
            "hash": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs"
        },
        {
            "type": 2,
            "nonce": 2,
            "to": null,
            "senderPubKey": "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
            "from": "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad",
            "value": "10",
            "timestamp": 1508673895,
            "fee": "0.00003",
            "invokeArgs": null,
            "sig": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs",
            "hash": "shhfd7387ydhudhsy8ehhhfjsg748hd"
        }
    ],
    "hash": "lkssjajsdnaskomcskosks",
    "sig": "abc"
}`

var TestBlock2 = `{
    "header": {
        "parentHash": null,
        "creatorPubKey": "49VzsGezoNjJQxHjekoCQP9CXZUs34CmCY53kGaHyR9rCJQJbJW",
        "number": 2,
        "stateRoot": "abc",
        "transactionsRoot": "sjifhahe3i9nfi",
        "nonce": 0,
        "mixHash": "",
        "difficulty": "500000",
        "timestamp": 0
    },
    "transactions": [
        {
            "type": 1,
            "nonce": 2,
            "to": "efjshfhh389djn29snmnvuis",
            "senderPubKey": "xsj2909jfhhjskmj99k",
            "from": "",
            "value": "100.333",
            "timestamp": 1508673895,
            "fee": "0.00003",
            "invokeArgs": null,
            "sig": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs",
            "hash": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs"
        },
        {
            "type": 2,
            "nonce": 2,
            "to": null,
            "senderPubKey": "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
            "from": "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad",
            "value": "10",
            "timestamp": 1508673895,
            "fee": "0.00003",
            "invokeArgs": null,
            "sig": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs",
            "hash": "shhfd7387ydhudhsy8ehhhfjsg748hd"
        }
    ],
    "hash": "lkssjajsdnaskomcskosks",
    "sig": "abc"
}`

var TestBlock3 = `{
    "header": {
        "parentHash": null,
        "creatorPubKey": "49VzsGezoNjJQxHjekoCQP9CXZUs34CmCY53kGaHyR9rCJQJbJW",
        "number": 3,
        "stateRoot": "abc",
        "transactionsRoot": "sjifhahe3dasasdai9nfi",
        "nonce": 0,
        "mixHash": "",
        "difficulty": "500000",
        "timestamp": 0
    },
    "transactions": [
        {
            "type": 1,
            "nonce": 2,
            "to": "efjshfhh389djn29snmnvuis",
            "senderPubKey": "xsj2909jfhhjskmj99k",
            "from": "",
            "value": "100.333",
            "timestamp": 1508673895,
            "fee": "0.00003",
            "invokeArgs": null,
            "sig": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs",
            "hash": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs"
        },
        {
            "type": 2,
            "nonce": 2,
            "to": null,
            "senderPubKey": "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
            "from": "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad",
            "value": "10",
            "timestamp": 1508673895,
            "fee": "0.00003",
            "invokeArgs": null,
            "sig": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs",
            "hash": "shhfd7387ydhudhsy8ehhhfjsg748hd"
        }
    ],
    "hash": "lkssjajsdnaskomcskosks",
    "sig": "abc"
}`
