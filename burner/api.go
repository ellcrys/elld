package burner

import (
	"fmt"
	"path"

	"github.com/ellcrys/elld/ltcsuite/ltcd/rpcclient"

	"github.com/ellcrys/elld/ltcsuite/ltcd"
	"github.com/ellcrys/elld/ltcsuite/ltcd/txscript"

	"github.com/ellcrys/elld/ltcsuite/ltcd/chaincfg/chainhash"

	"github.com/ellcrys/elld/ltcsuite/ltcd/wire"

	"github.com/ellcrys/elld/ltcsuite/ltcutil"

	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/params"
	"github.com/ellcrys/elld/util"
	"github.com/shopspring/decimal"
	"gopkg.in/asaskevich/govalidator.v4"

	"github.com/ellcrys/elld/accountmgr"
	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/rpc"
	"github.com/ellcrys/elld/rpc/jsonrpc"
	"github.com/ellcrys/elld/types"
)

// API exposes JSON-RPC methods that perform
// coin burning operations
type API struct {
	db         elldb.DB
	cfg        *config.EngineConfig
	rpcClient  *rpcclient.Client
	accountMgr *accountmgr.AccountManager
}

type burnBody struct {
	WIF      string  `json:"wif"`
	Producer string  `json:"producer"`
	Amount   string  `json:"amount"`
	Fee      float64 `json:"fee"`
}

var jsonErr = jsonrpc.Error

// NewAPI creates an instance of API
func NewAPI(db elldb.DB, cfg *config.EngineConfig,
	rpcClient *rpcclient.Client, am *accountmgr.AccountManager) *API {
	return &API{
		db:         db,
		cfg:        cfg,
		rpcClient:  rpcClient,
		accountMgr: am,
	}
}

// apiBurnerAccountBalance gets the balance of a burner account
func (api *API) apiBurnerAccountBalance(arg interface{}) *jsonrpc.Response {

	burnerAddress, ok := arg.(string)
	if !ok {
		return jsonErr(types.ErrCodeUnexpectedArgType, rpc.ErrMethodArgType("String").Error(), nil)
	}

	// Get the burner account
	am := accountmgr.New(path.Join(api.cfg.DataDir(), "accounts"))
	_, err := am.GetBurnerAccountByAddress(burnerAddress)
	if err != nil {
		return jsonErr(types.ErrCodeAccountNotFound, err.Error(), nil)
	}

	return jsonrpc.Success(util.EncodeForJS(map[string]interface{}{
		"balance": BalanceOf(api.db, burnerAddress),
	}))
}

func validateBurnBody(body *burnBody) *jsonrpc.Response {

	// Expects wif with valid type
	if body.WIF == "" {
		return jsonErr(types.ErrCodeCallParamError, "WIF is required", nil)
	}

	// Expects a valid producer address
	if body.Producer == "" {
		return jsonErr(types.ErrCodeCallParamError, "Producer address is required", nil)
	} else if crypto.IsValidAddr(body.Producer) != nil {
		return jsonErr(types.ErrCodeCallParamError, "Producer address is not a valid "+
			"network address", nil)
	}

	// Amount must be a numeric string and not less than the allowed minimum
	if body.Amount == "" {
		return jsonErr(types.ErrCodeCallParamError, "Amount is required", nil)
	} else if !govalidator.IsFloat(body.Amount) && !govalidator.IsNumeric(body.Amount) {
		return jsonErr(types.ErrCodeCallParamError, "Amount must be numeric", nil)
	} else {
		amt, _ := decimal.NewFromString(body.Amount)
		if amt.LessThan(params.MinimumBurnAmt) {
			msg := fmt.Sprintf("Amount cannot be less than the minimum (%s)",
				params.MinimumBurnAmt.String())
			return jsonErr(types.ErrCodeCallParamError, msg, nil)
		}
	}

	return nil
}

// apiBurn burns coins.
// It interacts with the burner chain via its RPC service.
// An error is returned if the burner chain RPC service is not running.
// The expected arguments are the:
// - `wif` key which must belong to an account on the node.
// - `producer` is the network address of a prospective block producer.
// - `amount` is the number of coins to be burned.
// - `fee` (optional) is the fee to pay to miners on the burn chain.
func (api *API) apiBurn(arg interface{}) *jsonrpc.Response {

	// Ensure burner chain RPC service is enabled.
	// We cannot perform this operation without access to RPC service.
	if !ltcd.IsRPCOn() {
		return jsonErr(types.ErrCodeUnexpected, "burner chain rpc service is disabled", nil)
	}

	// So here, the RPC is enabled. But we need to ensure the
	// API has a valid RPC client reference
	if api.rpcClient == nil {
		return jsonErr(types.ErrCodeUnexpected, "unable to proceed. uninitialized rpc client", nil)
	}

	body, ok := arg.(map[string]interface{})
	if !ok {
		return jsonErr(types.ErrCodeUnexpectedArgType, rpc.ErrMethodArgType("JSON").Error(), nil)
	}

	var burnBody burnBody
	if err := util.MapDecode(body, &burnBody); err != nil {
		return jsonErr(types.ErrCodeCallParamTypeError, err.Error(), nil)
	}

	// If fee is equal or less than zero, use default fee
	if burnBody.Fee <= 0 {
		burnBody.Fee = params.DefaultBurnTxFee
	}

	if valRes := validateBurnBody(&burnBody); valRes != nil {
		return valRes
	}

	// Decode the wif private key
	wif, err := ltcutil.DecodeWIF(burnBody.WIF)
	if err != nil {
		return jsonErr(types.ErrCodeCallParamError, "wif is not valid", nil)
	}

	wifToKey := crypto.NewSecp256k1FromWIF(wif)
	address := wifToKey.Addr()

	// Ensure the address matches that of a known account
	_, err = api.accountMgr.GetBurnerAccountByAddress(address)
	if err != nil {
		if err == accountmgr.ErrAccountNotFound {
			return jsonErr(types.ErrCodeCallParamError, "burner account is unknown", nil)
		}
		return jsonErr(types.ErrCodeUnexpected, "failed to get list of burner accounts", nil)
	}

	amount, _ := decimal.NewFromString(burnBody.Amount)
	requiredAmt := amount.Add(decimal.NewFromFloat(burnBody.Fee))

	// Find utxos that can satisfy the burn amount
	totalAmountUTXOS := decimal.Zero
	var candidateUTXOs = []*DocUTXO{}
	for _, utxo := range GetAllUTXO(api.db, address) {

		// Stop collecting UTXO if we have collected enough to cover
		// the burn amount and transaction fee
		if totalAmountUTXOS.GreaterThanOrEqual(requiredAmt) {
			break
		}

		value := decimal.NewFromFloat(utxo.Value)
		candidateUTXOs = append(candidateUTXOs, utxo)
		totalAmountUTXOS = totalAmountUTXOS.Add(value)
	}

	// Ensure we have enough UTXOs to continue
	if totalAmountUTXOS.LessThan(requiredAmt) {
		return jsonErr(types.ErrCodeUnexpected, "insufficient balance", nil)
	}

	tx := wire.NewMsgTx(2)

	// Build the input
	for _, utxo := range candidateUTXOs {
		outHash, err := chainhash.NewHashFromStr(utxo.TxHash)
		if err != nil {
			return jsonErr(types.ErrCodeUnexpected, "invalid utxo id", nil)
		}
		output := wire.NewOutPoint(outHash, utxo.Index)
		tx.AddTxIn(wire.NewTxIn(output, nil, nil))
	}

	// Build the outputs.
	// One will be the OP_RETURN output that spends coins.
	// The other will send change to the change address.
	decodedAddr, _ := crypto.DecodeAddrOnly(address)
	opReturnOutData := decodedAddr[:]
	opReturnOutScript, err := txscript.NewScriptBuilder().AddOp(txscript.OP_RETURN).
		AddData(opReturnOutData).Script()
	if err != nil {
		return jsonErr(types.ErrCodeUnexpected, "failed to create op_return output", nil)
	}

	// Build the op_return output and set the burn amount
	amtFloat, _ := amount.Float64()
	amtToSat, err := ltcutil.NewAmount(amtFloat)
	if err != nil {
		return jsonErr(types.ErrCodeUnexpected, err.Error(), nil)
	}

	tx.AddTxOut(wire.NewTxOut(int64(amtToSat), opReturnOutScript))

	// Calculate the change, if change is not equal to 0
	changeDec := totalAmountUTXOS.Sub(requiredAmt)
	if !changeDec.IsZero() {

		// Build the change output script
		changeScript, err := txscript.PayToAddrScript(wifToKey.Address())
		if err != nil {
			return jsonErr(types.ErrCodeUnexpected, "failed to create change script", nil)
		}

		changeFloat, _ := changeDec.Float64()
		changeSat, err := ltcutil.NewAmount(changeFloat)
		tx.AddTxOut(wire.NewTxOut(int64(changeSat), changeScript))
	}

	// Create spend signature for input at index 0
	spendFromScript, err := txscript.PayToAddrScript(wifToKey.Address())
	if err != nil {
		return jsonErr(types.ErrCodeUnexpected, err.Error(), nil)
	}
	for i := range candidateUTXOs {
		pubSig, err := txscript.SignatureScript(tx, i, spendFromScript,
			txscript.SigHashAll, wif.PrivKey, true)
		if err != nil {
			msg := fmt.Sprintf("failed to create signature script for input (%d)", i)
			return jsonErr(types.ErrCodeUnexpected, msg, nil)
		}
		tx.TxIn[i].SignatureScript = pubSig
	}

	// Send the transaction
	txHash, err := api.rpcClient.SendRawTransaction(tx, false)
	if err != nil {
		return jsonErr(types.ErrCodeUnexpected, "failed to send transaction: "+err.Error(), nil)
	}

	return jsonrpc.Success(map[string]interface{}{
		"hash": txHash.String(),
	})
}

// APIs returns all API handlers
func (api *API) APIs() jsonrpc.APISet {
	return map[string]jsonrpc.APIInfo{

		"getBalance": {
			Namespace:   types.NamespaceBurner,
			Description: "Get the balance of a burner account.",
			Func:        api.apiBurnerAccountBalance,
		},

		"burn": {
			Namespace:   types.NamespaceBurner,
			Description: "Creates a Litecoin transaction that burns litecoins.",
			Func:        api.apiBurn,
		},
	}
}
