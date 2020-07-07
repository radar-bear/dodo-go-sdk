package dodo_sdk

import (
	"errors"
	"github.com/radar-bear/goWeb3"
	"github.com/radar-bear/goWeb3/helper"
	"github.com/shopspring/decimal"
	"math/big"
)

type DODOContract struct {
	Address      string
	Contract     *goWeb3.Contract
	Base         *ERC20Contract
	Quote        *ERC20Contract
	BaseDecimal  int
	QuoteDecimal int
	Web3         *goWeb3.Web3
}

func NewDODOContract(baseTokenAddress string, quoteTokenAddress string) (dodo *DODOContract, err error) {
	DODOZoo, err := NewDODOZooContract()
	if err != nil {
		return
	}

	DODOAddress, err := DODOZoo.GetDODOAddress(baseTokenAddress, quoteTokenAddress)
	if err != nil {
		return
	}

	contract, err := DODOZoo.Web3.NewContract(DODOAbi, DODOAddress)
	if err != nil {
		return
	}

	baseRawAddress, err := contract.Call("_BASE_TOKEN_")
	if err != nil {
		return
	}
	quoteRawAddress, err := contract.Call("_QUOTE_TOKEN_")
	if err != nil {
		return
	}

	Base, err := NewERC20Contract(StdAddr(baseRawAddress))
	if err != nil {
		return
	}
	Quote, err := NewERC20Contract(StdAddr(quoteRawAddress))
	if err != nil {
		return
	}

	dodo = &DODOContract{
		DODOAddress,
		contract,
		Base,
		Quote,
		Base.Decimals,
		Quote.Decimals,
		DODOZoo.Web3,
	}
	return
}

// ================= Getters =================

func (d *DODOContract) GetBalancedStatus() (B0 decimal.Decimal, Q0 decimal.Decimal, err error) {
	res, err := d.Contract.Call("getExpectedTarget")
	if err != nil {
		return
	}
	B0 = helper.HexString2Decimal(SplitWeb3ReturnValue(res, 0), int32(d.BaseDecimal)*-1)
	Q0 = helper.HexString2Decimal(SplitWeb3ReturnValue(res, 1), int32(d.QuoteDecimal)*-1)
	return
}

func (d *DODOContract) GetDODOBalances() (B decimal.Decimal, Q decimal.Decimal, err error) {
	res, err := d.Contract.Call("_BASE_BALANCE_")
	if err != nil {
		return
	}
	B = helper.HexString2Decimal(res, int32(d.BaseDecimal)*-1)
	Q = helper.HexString2Decimal(res, int32(d.QuoteDecimal)*-1)
	return
}

func (d *DODOContract) GetFeeRate() (feeRate decimal.Decimal, err error) {
	rawLpFeeRate, err := d.Contract.Call("_LP_FEE_RATE_")
	if err != nil {
		return
	}
	rawMtFeeRate, err := d.Contract.Call("_MT_FEE_RATE_")
	if err != nil {
		return
	}
	feeRate = helper.HexString2Decimal(rawLpFeeRate, -18).Add(helper.HexString2Decimal(rawMtFeeRate, -18))
	return
}

func (d *DODOContract) GetK() (k decimal.Decimal, err error) {
	res, err := d.Contract.Call("_K_")
	if err != nil {
		return
	}
	k = helper.HexString2Decimal(res, -18)
	return
}

func (d *DODOContract) GetOraclePrice() (price decimal.Decimal, err error) {
	res, err := d.Contract.Call("getOraclePrice")
	if err != nil {
		return
	}
	// e.g. if base decimal is 9 and quote decimal is 18, oracle price has decimal 27
	price = helper.HexString2Decimal(res, int32(d.BaseDecimal-d.QuoteDecimal-18))
	return
}

func (d *DODOContract) GetWithdrawPenalty(isBaseToken bool, amount decimal.Decimal) (penalty decimal.Decimal, err error) {
	var decimals int
	var funcName string
	if isBaseToken {
		decimals = d.BaseDecimal
		funcName = "getWithdrawBasePenalty"
	} else {
		decimals = d.QuoteDecimal
		funcName = "getWithdrawQuotePenalty"
	}
	rawAmount := helper.DecimalToBigInt(amount.Div(decimal.New(1, int32(decimals))))
	rawPenalty, err := d.Contract.Call(funcName, rawAmount)
	if err != nil {
		return
	}
	penalty = helper.HexString2Decimal(rawPenalty, int32(decimals)*-1)
	return
}

func (d *DODOContract) QuerySellPrice(amount decimal.Decimal) (price decimal.Decimal, err error) {
	rawPayQuote, err := d.Contract.Call("querySellBaseToken", helper.DecimalToBigInt(amount.Div(decimal.New(1, int32(d.BaseDecimal)))))
	if err != nil {
		return
	}
	payQuote := helper.HexString2Decimal(rawPayQuote, int32(d.QuoteDecimal)*-1)
	price = payQuote.Div(amount)
	return
}

func (d *DODOContract) QueryBuyPrice(amount decimal.Decimal) (price decimal.Decimal, err error) {
	rawReceiveQuote, err := d.Contract.Call("queryBuyBaseToken", helper.DecimalToBigInt(amount.Div(decimal.New(1, int32(d.BaseDecimal)))))
	if err != nil {
		return
	}
	receiveQuote := helper.HexString2Decimal(rawReceiveQuote, int32(d.QuoteDecimal)*-1)
	price = receiveQuote.Div(amount)
	return
}

// ================= Trader =================

func (d *DODOContract) Trade(side string, baseTokenAmount decimal.Decimal, priceLimit decimal.Decimal, sendParams *goWeb3.SendTxParams) (txHash string, err error) {
	var funcName string
	switch side {
	case SELL:
		funcName = "sellBaseToken"
		break
	case BUY:
		funcName = "buyBaseToken"
		break
	default:
		err = errors.New("DODO TRADE SIDE WRONG")
		return
	}
	quoteTokenAmount := baseTokenAmount.Mul(priceLimit)
	rawQuoteTokenAmount := helper.DecimalToBigInt(quoteTokenAmount.Mul(decimal.New(1, int32(d.QuoteDecimal))))
	rawBaseTokenAmount := helper.DecimalToBigInt(baseTokenAmount.Mul(decimal.New(1, int32(d.BaseDecimal))))
	return d.Contract.Send(sendParams, big.NewInt(0), funcName, rawBaseTokenAmount, rawQuoteTokenAmount)
}

// ================= Liquidity Provider =================

func (d *DODOContract) Deposit(isBaseToken bool, amount decimal.Decimal, sendParams *goWeb3.SendTxParams) (txHash string, err error) {
	var funcName string
	var decimals int
	if isBaseToken {
		funcName = "depositBase"
		decimals = d.BaseDecimal
	} else {
		funcName = "depositQuote"
		decimals = d.QuoteDecimal
	}
	rawAmount := helper.DecimalToBigInt(amount.Mul(decimal.New(1, int32(decimals))))
	return d.Contract.Send(sendParams, big.NewInt(0), funcName, rawAmount)
}

func (d *DODOContract) Withdraw(isBaseToken bool, amount decimal.Decimal, sendParams *goWeb3.SendTxParams) (txHash string, err error) {
	var funcName string
	var decimals int
	if isBaseToken {
		funcName = "withdrawBase"
		decimals = d.BaseDecimal
	} else {
		funcName = "withdrawQuote"
		decimals = d.QuoteDecimal
	}
	rawAmount := helper.DecimalToBigInt(amount.Mul(decimal.New(1, int32(decimals))))
	return d.Contract.Send(sendParams, big.NewInt(0), funcName, rawAmount)
}

// ================= Receipt Parser =================

type DODOTradeLog struct {
	Trader string
	Amount decimal.Decimal
	Price  decimal.Decimal
	Side   string
}

func (d *DODOContract) ParseTradeTx(txHash string) (trades []DODOTradeLog, err error) {
	receipt, err := d.Web3.GetRecipt(txHash)
	if err != nil {
		return
	}
	for _, log := range receipt.Logs {
		// sell base token
		if log.Topics[0] == "0xd8648b6ac54162763c86fd54bf2005af8ecd2f9cb273a5775921fd7f91e17b2d" {
			baseToken := helper.HexString2Decimal(SplitWeb3ReturnValue(log.Data, 0), int32(d.BaseDecimal)*-1)
			quoteToken := helper.HexString2Decimal(SplitWeb3ReturnValue(log.Data, 1), int32(d.QuoteDecimal)*-1)
			trade := DODOTradeLog{
				StdAddr(log.Topics[1]),
				baseToken,
				quoteToken.Div(baseToken),
				SELL,
			}
			trades = append(trades, trade)
		}
		// buy base token
		if log.Topics[0] == "0xe93ad76094f247c0dafc1c61adc2187de1ac2738f7a3b49cb20b2263420251a3" {
			baseToken := helper.HexString2Decimal(SplitWeb3ReturnValue(log.Data, 0), int32(d.BaseDecimal)*-1)
			quoteToken := helper.HexString2Decimal(SplitWeb3ReturnValue(log.Data, 1), int32(d.QuoteDecimal)*-1)
			trade := DODOTradeLog{
				StdAddr(log.Topics[1]),
				baseToken,
				quoteToken.Div(baseToken),
				BUY,
			}
			trades = append(trades, trade)
		}
	}
	return
}

type DODODepositLog struct {
	Payer             string
	LiquidityProvider string
	Amount            decimal.Decimal
	IsBaseToken       bool
}

func (d *DODOContract) ParseDepositTx(txHash string) (depositLogs []DODODepositLog, err error) {
	receipt, err := d.Web3.GetRecipt(txHash)
	if err != nil {
		return
	}
	for _, log := range receipt.Logs {
		// deposit base
		if log.Topics[0] == "0xb0f1d6b2bf09eb5e858f8722141866730907dbac3748137e2c733caebe552e0d" {
			depositLog := DODODepositLog{
				StdAddr(log.Topics[1]),
				StdAddr(log.Topics[2]),
				helper.HexString2Decimal(log.Data, int32(d.BaseDecimal)*-1),
				true,
			}
			depositLogs = append(depositLogs, depositLog)
		}
		// deposit quote
		if log.Topics[0] == "0xda08e2ce8fe6d34374c45827209f01c55962f5d3a2e60b7adaddab0a34a9c50d" {
			depositLog := DODODepositLog{
				StdAddr(log.Topics[1]),
				StdAddr(log.Topics[2]),
				helper.HexString2Decimal(log.Data, int32(d.QuoteDecimal)*-1),
				false,
			}
			depositLogs = append(depositLogs, depositLog)
		}
	}
	return
}

type DODOWithdrawLog struct {
	LiquidityProvider string
	Receiver          string
	Amount            decimal.Decimal
	IsBaseToken       bool
}

func (d *DODOContract) ParseWithdrawTx(txHash string) (withdrawLogs []DODOWithdrawLog, err error) {
	receipt, err := d.Web3.GetRecipt(txHash)
	if err != nil {
		return
	}
	for _, log := range receipt.Logs {
		// withdraw base
		if log.Topics[0] == "0x8fd51fc63578638d083f35f5eb02f543a9877a54ca3fbd7e085e6e4c8fdcc42d" {
			withdrawLog := DODOWithdrawLog{
				StdAddr(log.Topics[1]),
				StdAddr(log.Topics[2]),
				helper.HexString2Decimal(log.Data, int32(d.BaseDecimal)*-1),
				true,
			}
			withdrawLogs = append(withdrawLogs, withdrawLog)
		} // withdraw quote
		if log.Topics[0] == "0x47663350bc8ab956c1618a2efdfdd9ed10c970c0b40f077dd3e708dd58e67517" {
			withdrawLog := DODOWithdrawLog{
				StdAddr(log.Topics[1]),
				StdAddr(log.Topics[2]),
				helper.HexString2Decimal(log.Data, int32(d.QuoteDecimal)*-1),
				false,
			}
			withdrawLogs = append(withdrawLogs, withdrawLog)
		}
	}
	return
}
