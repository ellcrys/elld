This snippet includes a codes to create blocks within the test environment. This will come
handy when there is need to create a test block for custom purposes.

```go
BeforeEach(func() {
    BeforeEach(func() {
		blk, err := bc.GenerateBlock(&GenerateBlockParams{
			Transactions: []*wire.Transaction{
				wire.NewTx(wire.TxTypeBalance, 123, receiver.Addr(), sender, "1", "0.1", 1532730722),
			},
			Creator:    sender,
			Nonce:      wire.EncodeNonce(1),
			MixHash:    util.BytesToHash([]byte("mix hash")),
			Difficulty: new(big.Int).SetInt64(500),
		}, ChainOp{Chain: chain})
		Expect(err).To(BeNil())
		Expect(blk).ToNot(BeNil())
		pp.Println(blk)
	})
})
```