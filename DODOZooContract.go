package dodo_go_sdk

import (
	"errors"
	"github.com/radar-bear/goWeb3"
	"os"
)

type DODOZooContract struct {
	Address  string
	Contract *goWeb3.Contract
	Web3     *goWeb3.Web3
}

func NewDODOZooContract(address string) (zoo *DODOZooContract, err error) {

	nodeUrl := os.Getenv("ETH_NODE_URL")
	if nodeUrl == "" {
		err = errors.New("Require ENV ETH_NODE_URL")
		return
	}

	web3 := goWeb3.NewWeb3(nodeUrl)
	contract, err := web3.NewContract(DODOZooAbi, address)
	if err != nil {
		return
	}
	zoo = &DODOZooContract{
		address,
		contract,
		web3,
	}
	return
}

func (z *DODOZooContract) GetDODOAddress(baseTokenAddress string, quoteTokenAddress string) (address string, err error) {
	resp, err := z.Contract.Call("getDODO", goWeb3.HexToAddress(baseTokenAddress), goWeb3.HexToAddress(quoteTokenAddress))
	if err != nil {
		return
	}
	address = StdAddr(resp)
	if address == EmptyAddress {
		err = errors.New("DODO NOT EXIST")
	}
	return
}
