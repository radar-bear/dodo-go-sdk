package dodo_sdk

import (
	"errors"
	"github.com/radar-bear/goWeb3"
	"os"
	"strings"
)

type DODOZooContract struct {
	Address  string
	Contract *goWeb3.Contract
	Web3     *goWeb3.Web3
}

func NewDODOZooContract() (zoo *DODOZooContract, err error) {
	info, err := GetDepolyedInfo()
	if err != nil {
		return
	}

	network := strings.ToLower(os.Getenv("NETWORK"))
	var address = info.Mainnet["DODOZoo"]
	if network == "kovan" {
		address = info.Kovan["DODOZoo"]
	}

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
