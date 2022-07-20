package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	ethcom "github.com/ethereum/go-ethereum/common"
	rds_conn "github.com/hpb-project/tokenfaucet/cacheTools"
	"github.com/hpb-project/tokenfaucet/common"
	"github.com/hpb-project/tokenfaucet/common/ethrpc"
	"github.com/hpb-project/tokenfaucet/contracts"
	"github.com/hpb-project/tokenfaucet/models"
	"github.com/hpb-project/tokenfaucet/types"
	"strconv"
	"sync"
)

var (
	txMux = sync.Mutex{}
)

type FaucetController struct {
	beego.Controller
	Networks map[string]*types.Network
}

func (f *FaucetController) ResponseInfo(code int, err_msg string, result interface{}) {
	switch code {
	case 200:
		f.Data["json"] = map[string]interface{}{"error": "200", "err_msg": err_msg, "data": result}
	default:
		f.Data["json"] = map[string]interface{}{"error": "500", "err_msg": err_msg, "data": result}
	}
	f.ServeJSON()
}

func (f *FaucetController) TokenTransfer() {
	beego.Info("got request token transfer")
	txMux.Lock()
	ip := common.GetClientIP(f.Ctx)

	var param models.Param
	data := f.Ctx.Input.RequestBody
	json.Unmarshal(data, &param)

	network, exist := f.Networks[param.Network]
	if !exist {
		txMux.Unlock()
		beego.Error("Not supported network", "param.network", param.Network)
		f.ResponseInfo(500, "Not supported network.", "")
		return
	}

	t, c, err := rds_conn.CheckIpAddress(ip)
	if err != nil {
		txMux.Unlock()
		beego.Error("check ipaddress", "err", err)
		f.ResponseInfo(500, err.Error(), "")
		return
	}
	logs.Info("current ip is ", ip, "count is ", c)
	var succeed bool = false
	oldC := c
	rds_conn.SR.SetKvAndExp(ip, c, t)
	txMux.Unlock()
	defer func() {
		if !succeed {
			// recover set ip
			intC, _ := strconv.Atoi(oldC)
			lastC := intC - 1
			rds_conn.SR.SetKvAndExp(ip, strconv.Itoa(lastC), t)
		}
	}()

	if boo := common.CheckAddress(param.To); !boo {
		beego.Error("check address", "addr", param.To)
		f.ResponseInfo(500, "Request address format exception, please re-enter.", "")
		return
	}
	checkkey := fmt.Sprintf("%s-%s", param.Network, param.To)

	if boo := rds_conn.SR.IsKeyExists(checkkey); boo {
		beego.Error("check time limit ", "resut ", boo)
		f.ResponseInfo(500, "Exceeding the daily limit.", "")
		return
	}
	url := network.Url
	myKey := common.GetPrivateKeyConfig()
	amount := common.GetLimitAmount()
	tie := common.GetLimitTime()

	rds_conn.SR.SetKvAndExp(checkkey, param.To, tie)
	defer func() {
		if !succeed {
			//clear address.
			rds_conn.SR.Del(param.To)
		}
	}()

	var txHash string
	toAddr := ethcom.HexToAddress(param.To)
	ctx := context.Background()
	{
		tokenAddr := ethcom.HexToAddress(network.Token)
		tx, err := contracts.TokenTransfer(ctx, url, tokenAddr, myKey, toAddr, amount, ethrpc.GetNonce(network))
		if err != nil {
			beego.Error("token transfer failed ", " err ", err)
			f.ResponseInfo(500, fmt.Sprintf("token transfer failed: %s", err), "")
			return
		}
		txHash = tx.Hash().String()
	}
	if network.Coin > 0 {
		tx, err := ethrpc.EthTransfer(network, myKey, toAddr, network.Coin)
		if err != nil {
			beego.Error("send coin transfer failed ", " err ", err)
			f.ResponseInfo(500, fmt.Sprintf("coin transfer failed: %s", err), "")
			return
		}
		txHash = tx.Hash().String()
	}

	//After the account address is successfully collected,
	//it is saved in redis to limit the collection frequency of users
	if len(txHash) > 0 {
		succeed = true
		rds_conn.SR.SetKvAndExp(checkkey, param.To, tie)
		f.ResponseInfo(200, "", txHash)

		return
	}

	f.ResponseInfo(500, "transfer failure", "")
	return
}
