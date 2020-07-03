package dodo_sdk

import (
	"github.com/radar-bear/goWeb3"
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

// func (d *DODOContract) queryBuyPrice(amount decimal.Decimal) (price decimal.Decimal, err error) {
//
// }
//
// func (d *DODOContract) querySellPrice(amount decimal.Decimal) (price decimal.Decimal, err error) {
//
// }
