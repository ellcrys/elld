This snippet includes a codes to create blocks within the test environment. This will come
handy when there is need to create a test block for custom purposes.

```go
BeforeEach(func() {
    blk, err := bc.GenerateBlock(&GenerateBlockParams{
        Transactions: []*wire.Transaction{wire.NewTx(wire.TxTypeBalance, 123, receiver.Addr(), sender, "0.1", "0.1", time.Now().Unix())},
        Creator:      sender,
        Nonce:        384772,
        MixHash:      "0x1cdf0e214bcdb7af36885316506f7388f262f7b710a28a00d21706550cdd72c2",
        Difficulty:   "102994",
    }, ChainOp{Chain: chain})
    Expect(err).To(BeNil())
    Expect(blk).ToNot(BeNil())
    pretty.Println(blk)
})
```