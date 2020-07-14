package dodo_go_sdk

import (
	"errors"
	"github.com/radar-bear/goWeb3"
	"github.com/radar-bear/goWeb3/helper"
	"github.com/shopspring/decimal"
	"os"
)

type ERC20Contract struct {
	Address  string
	Contract *goWeb3.Contract
	Web3     *goWeb3.Web3
	Decimals int
}

func NewERC20Contract(address string) (erc20 *ERC20Contract, err error) {
	nodeUrl := os.Getenv("ETH_NODE_URL")
	if nodeUrl == "" {
		err = errors.New("Require ENV ETH_NODE_URL")
		return
	}

	web3 := goWeb3.NewWeb3(nodeUrl)
	contract, err := web3.NewContract(ERC20Abi, address)
	if err != nil {
		return
	}

	res, err := contract.Call("decimals")
	decimals, err := helper.HexString2Int(res)
	if err != nil {
		return
	}

	erc20 = &ERC20Contract{
		address,
		contract,
		web3,
		decimals,
	}

	return
}

func (token *ERC20Contract) BalanceOf(address string) (rawBalance decimal.Decimal, balance decimal.Decimal, err error) {
	res, err := token.Contract.Call("balanceOf", goWeb3.HexToAddress(address))
	if err != nil {
		return
	}
	rawBalance = helper.HexString2Decimal(res, 0)
	balance = token.ToReadableBalance(rawBalance)
	return
}

func (token *ERC20Contract) ToRawBalance(balance decimal.Decimal) decimal.Decimal {
	return balance.Mul(decimal.New(1, int32(token.Decimals)))
}

func (token *ERC20Contract) ToReadableBalance(balance decimal.Decimal) decimal.Decimal {
	return balance.Mul(decimal.New(1, int32(token.Decimals)*-1))
}
