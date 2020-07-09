package dodo_sdk

import (
	"encoding/base64"
	"encoding/json"
	"github.com/radar-bear/goWeb3/helper"
	"strings"
)

const EmptyAddress = "0x0000000000000000000000000000000000000000"
const SELL = "sell"
const BUY = "BUY"

func StdAddr(address string) string {
	if len(address) > 40 {
		return "0x" + strings.ToLower(address[len(address)-40:])
	} else {
		return strings.ToLower(Add0xPrefix(address))
	}
}

func Remove0xPrefix(originStr string) string {
	if len(originStr) < 2 {
		return originStr
	}
	if originStr[0:2] == "0x" {
		return originStr[2:]
	} else {
		return originStr
	}
}

func Add0xPrefix(originStr string) string {
	return "0x" + Remove0xPrefix(originStr)
}

func SplitWeb3ReturnValue(returnValue string, position int) string {
	return returnValue[2+64*position : 66+64*position]
}

type DeployedInfo struct {
	Mainnet map[string]string `json:"mainnet"`
	Kovan   map[string]string `json:"kovan"`
	Ropsten map[string]string `json:"ropsten"`
}

type gitResp struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Sha         string `json:"sha"`
	Size        int    `json:"size"`
	URL         string `json:"url"`
	HTMLURL     string `json:"html_url"`
	GitURL      string `json:"git_url"`
	DownloadURL string `json:"download_url"`
	Type        string `json:"type"`
	Content     string `json:"content"`
	Encoding    string `json:"encoding"`
	Links       struct {
		Self string `json:"self"`
		Git  string `json:"git"`
		HTML string `json:"html"`
	} `json:"_links"`
}

func GetDepolyedInfo() (info DeployedInfo, err error) {
	resp, err := helper.Get(
		"https://api.github.com/repos/radar-bear/dodo-docs/contents/DeployedInfo.json",
		"",
		helper.EmptyKeyPairList,
		helper.EmptyKeyPairList,
	)
	if err != nil {
		return
	}
	var dataContainer gitResp
	json.Unmarshal([]byte(resp), &dataContainer)
	infoByte, err := base64.StdEncoding.DecodeString(dataContainer.Content)
	if err != nil {
		return
	}
	json.Unmarshal(infoByte, &info)
	return
}
