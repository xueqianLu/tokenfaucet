package common

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/hpb-project/tokenfaucet/types"
	"log"
)

func GetCacheConfig() *types.CacheConfig {
	conf := &types.CacheConfig{}
	conf.Conn = beego.AppConfig.String("cache::conn")
	conf.DBNum = beego.AppConfig.String("cache::dbNum")
	conf.Passworkd = beego.AppConfig.String("cache::password")
	return conf
}

func GetNetworkConfig(networkName string) *types.Network {
	name := fmt.Sprintf("network-%s", networkName)
	url := beego.AppConfig.String(fmt.Sprintf("%s::url", name))
	if len(url) == 0 {
		return nil
	}
	network := &types.Network{Name: networkName, Url: url}

	network.Token = beego.AppConfig.String(fmt.Sprintf("%s::token", name))
	network.Gaslimit, _ = beego.AppConfig.Int(fmt.Sprintf("%s::GasLimit", name))
	network.Chainid, _ = beego.AppConfig.Int(fmt.Sprintf("%s::chainID", name))
	network.Coin, _ = beego.AppConfig.Int64(fmt.Sprintf("%s::coin", name))

	network.Mykey = GetPrivateKeyConfig()
	network.Address, _ = GetAddrFromPrivk(network.Mykey)
	log.Println("use account ", network.Address)

	return network
}

func GetAllNetworkConfig() map[string]*types.Network {
	var networks = make(map[string]*types.Network)
	if mainnet := GetNetworkConfig("mainnet"); mainnet != nil {
		networks["mainnet"] = mainnet
	}
	if testnet := GetNetworkConfig("testnet"); testnet != nil {
		networks["testnet"] = testnet
	}
	return networks
}

func GetPrivateKeyConfig() string {
	return beego.AppConfig.String("account::MYKEY")
}

func GetLimitAmount() int64 {
	a, _ := beego.AppConfig.Int64("limit::amount")
	return a
}

func GetLimitTime() int {
	a, _ := beego.AppConfig.Int("limit::time")
	return a
}

func GetLimitCount() string {
	return beego.AppConfig.String("limit::count")
}

func GetLimitWarningValue() int64 {
	a, _ := beego.AppConfig.Int64("limit::warningValue")
	return a
}

// GetGithubClientID github
func GetGithubClientID() string {
	return beego.AppConfig.String("github::clientID")
}

func GetGithubClientSecret() string {
	return beego.AppConfig.String("github::clientSecret")
}

func GetGithubCallback() string {
	return beego.AppConfig.String("github::callback")
}

func GetGithubRedirect() string {
	return beego.AppConfig.String("github::redirect")
}

