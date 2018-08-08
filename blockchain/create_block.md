This snippet includes a codes to create blocks within the test environment. This will come
handy when there is need to create a test block for custom purposes.

```go
BeforeEach(func() {
    bc.chains = make(map[string]*Chain)
    chain := NewChain("c1", db, cfg, log)
    block2 = makeTestBlock(bc, chain, &common.GenerateBlockParams{
        Transactions: []*wire.Transaction{
            wire.NewTx(wire.TxTypeAllocCoin, 123, sender.Addr(), sender, "1", "0.1", 1532730722),
        },
        Creator:    sender,
        Nonce:      wire.EncodeNonce(1),
        MixHash:    util.BytesToHash([]byte("mix hash")),
        Difficulty: new(big.Int).SetInt64(500),
    })
})
```