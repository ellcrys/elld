package blockchain

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
