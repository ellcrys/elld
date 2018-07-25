package blockchain

import (
	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/wire"
)

// SeedTestGenesisState creates the initial accounts references
// in transactions of test genesis block.Should only be used for tests
func SeedTestGenesisState(store common.Store, gBlock *wire.Block) error {

	// create private key for initial accounts

	// Addr: eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad
	// PublicKey: 48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC
	// PrivateKey: wU7ckbRBWevtkoT9QoET1adGCsABPRtyDx5T9EHZ4paP78EQ1w5sFM2sZg87fm1N2Np586c98GkYwywvtgy9d2gEpWbsbU
	key := crypto.NewKeyFromIntSeed(1)

	// Addr: e6i7rxApBYUt7w94gGDKTz45A5J567JfkS
	// PublicKey: 48s9G48LD5eo5YMjJWmRjPaoDZJRNTuiscHMov6zDGMEqUg4vbG
	// PrivateKey: wPUNjfUCUmu2SMWgwharmEkP7hCFF2xAQsCEXurw5A1W2kaLBSNxowZJ8QiZ5a1rBZQXGRJSRACMzenuHPdhzu3SNmKaVk
	key2 := crypto.NewKeyFromIntSeed(2)

	// compute the genesis chain ID
	gChainID := util.ToHex(util.Blake2b256([]byte(gBlock.Hash)))

	// make the account keys. Set block number to the number of
	// the genesis block (1).
	acct1Key := common.MakeAccountKey(gBlock.GetNumber(), gChainID, key.Addr())
	acct2Key := common.MakeAccountKey(gBlock.GetNumber(), gChainID, key2.Addr())

	// Create account objects
	acct1 := &wire.Account{
		Type:    wire.AccountTypeBalance,
		Address: key.Addr(),
		Balance: "100",
	}
	acct2 := &wire.Account{
		Type:    wire.AccountTypeBalance,
		Address: key2.Addr(),
		Balance: "0",
	}

	// add account to the store
	err := store.Put(acct1Key, util.ObjectToBytes(acct1))
	if err != nil {
		return nil
	}
	err = store.Put(acct2Key, util.ObjectToBytes(acct2))
	if err != nil {
		return nil
	}

	return nil
}
