package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	ethcom "github.com/ethereum/go-ethereum/common"
	rds_conn "github.com/hpb-project/tokenfaucet/cacheTools"
	"github.com/hpb-project/tokenfaucet/common"
	"github.com/hpb-project/tokenfaucet/common/ethrpc"
	"github.com/hpb-project/tokenfaucet/contracts"
	"github.com/hpb-project/tokenfaucet/models"
	"github.com/hpb-project/tokenfaucet/types"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
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

func (f *FaucetController) Default() {
	f.githubCallbackRes(200, "", "")
}

func (f *FaucetController) GithubLoginHandler() {
	key := f.GetString("key")

	redirectURL := fmt.Sprintf("https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s",
		common.GetGithubClientID(),
		common.GetGithubCallback()+"?key="+key)

	s := strings.Split(key, "-")
	log.Println("GithubLoginHandler net:  ", s)

	log.Println("redirectURL", redirectURL)
	f.Ctx.Redirect(302, redirectURL)
}


func (f *FaucetController) GithubCallbackHandler() {
	key := f.GetString("key")
	code := f.GetString("code")

	s := strings.Split(key, "-")
	log.Println("GithubCallbackHandler net:  ", s[1])

	githubAccessToken, err := getGithubAccessToken(code)
	if err != nil {
		f.ResponseInfo(500, err.Error(), "")
		return
	}
	githubData, err := getGithubData(githubAccessToken)
	if err != nil {
		f.ResponseInfo(500, err.Error(), "")
		return
	}

	log.Println("GithubCallbackHandler login:  ", githubData.Login)

	//check github account
	f.githubTransfer(s[0], s[1], githubData.Login)
}

func (f *FaucetController) githubTransfer(to string, net string, login string) {
	network, exist := f.Networks[net]
	if !exist {
		beego.Error("Not supported network", "param.network", net)
		f.githubCallbackRes(500, "Not supported network", "")
		return
	}

	//验证地址是否正确
	if boo := common.CheckAddress(to); !boo {
		beego.Error("check address", "addr", to)
		f.githubCallbackRes(500, "Request address format exception, please re-enter.", "")
		return
	}
	checkkey := fmt.Sprintf("%s-%s-%s", "github", net, login)

	if boo := rds_conn.SR.IsKeyExists(checkkey); boo {
		beego.Error("check time limit ", "resut ", boo)
		f.githubCallbackRes(500, "Exceeding the daily limit.", "")
		return
	}

	url := network.Url
	myKey := common.GetPrivateKeyConfig()
	amount := common.GetLimitAmount()
	tie := common.GetLimitTime()

	var txHash string
	toAddr := ethcom.HexToAddress(to)
	ctx := context.Background()
	{
		tokenAddr := ethcom.HexToAddress(network.Token)
		tx, err := contracts.TokenTransfer(ctx, url, tokenAddr, myKey, toAddr, amount, ethrpc.GetNonce(network))
		if err != nil {
			beego.Error("token transfer failed ", " err ", err)
			f.githubCallbackRes(500, "token transfer failed", "")
			return
		}
		txHash = tx.Hash().String()
	}
	if network.Coin > 0 {
		tx, err := ethrpc.EthTransfer(network, myKey, toAddr, network.Coin)
		if err != nil {
			beego.Error("send coin transfer failed ", " err ", err)
			f.githubCallbackRes(400, "send token transfer failed", "")
			return
		}
		txHash = tx.Hash().String()
	}

	//After the account address is successfully collected,
	//it is saved in redis to limit the collection frequency of users
	if len(txHash) > 0 {
		rds_conn.SR.SetKvAndExp(checkkey, to, tie)
		f.githubCallbackRes(200, "", txHash)

		return
	}

	f.githubCallbackRes(500, "transfer failure", "")
	return
}

func (f *FaucetController) TokenTransfer() {
	ip := common.GetClientIP(f.Ctx)

	var param models.Param
	data := f.Ctx.Input.RequestBody
	json.Unmarshal(data, &param)

	network, exist := f.Networks[param.Network]
	if !exist {
		beego.Error("Not supported network", "param.network", param.Network)
		f.ResponseInfo(500, "Not supported network.", "")
		return
	}

	t, c, err := rds_conn.CheckIpAddress(ip)
	if err != nil {
		beego.Error("check ipaddress", "err", err)
		f.ResponseInfo(500, err.Error(), "")
		return
	}
	//验证地址是否正确
	if boo := common.CheckAddress(param.To); !boo {
		beego.Error("check address", "addr", param.To)
		f.ResponseInfo(500, "Request address format exception, please re-enter.", "")
		return
	}
	checkkey := fmt.Sprintf("%s-%s", param.Network, param.To)

	if boo := rds_conn.SR.IsKeyExists(checkkey); boo {
		beego.Error("check time limit ", "resut ", boo)
		f.ResponseInfo(500, "", "Exceeding the daily limit.")
		return
	}
	url := network.Url
	myKey := common.GetPrivateKeyConfig()
	amount := common.GetLimitAmount()
	tie := common.GetLimitTime()

	var txHash string
	toAddr := ethcom.HexToAddress(param.To)
	ctx := context.Background()
	{
		tokenAddr := ethcom.HexToAddress(network.Token)
		tx, err := contracts.TokenTransfer(ctx, url, tokenAddr, myKey, toAddr, amount, ethrpc.GetNonce(network))
		if err != nil {
			beego.Error("token transfer failed ", " err ", err)
			f.ResponseInfo(500, fmt.Sprintf("token transfer failed"), fmt.Sprintf("%s", err))
			return
		}
		txHash = tx.Hash().String()
	}
	if network.Coin > 0 {
		tx, err := ethrpc.EthTransfer(network, myKey, toAddr, network.Coin)
		if err != nil {
			beego.Error("send coin transfer failed ", " err ", err)
			f.ResponseInfo(500, fmt.Sprintf("coin transfer failed"), fmt.Sprintf("%s", err))
			return
		}
		txHash = tx.Hash().String()
	}

	rds_conn.SR.SetKvAndExp(ip, c, t)

	//After the account address is successfully collected,
	//it is saved in redis to limit the collection frequency of users
	if len(txHash) > 0 {
		rds_conn.SR.SetKvAndExp(checkkey, param.To, tie)
		f.ResponseInfo(200, "", txHash)

		return
	}

	f.ResponseInfo(500, "transfer failure", "false")
	return
}

func getGithubAccessToken(code string) (string, error) {
	requestBodyMap := map[string]string{"client_id": common.GetGithubClientID(), "client_secret": common.GetGithubClientSecret(), "code": code}
	requestJSON, _ := json.Marshal(requestBodyMap)

	req, reqerr := http.NewRequest("POST", "https://github.com/login/oauth/access_token", bytes.NewBuffer(requestJSON))
	if reqerr != nil {
		return "", errors.New("request creation failed")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, resperr := http.DefaultClient.Do(req)
	if resperr != nil {
		return "", errors.New("request failed")
	}

	respbody, _ := ioutil.ReadAll(resp.Body)

	// Represents the response received from Github
	type githubAccessTokenResponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
	}

	var ghresp githubAccessTokenResponse
	json.Unmarshal(respbody, &ghresp)

	return ghresp.AccessToken, nil
}

func getGithubData(accessToken string) (*UserData, error) {
	// Get request to a set URL
	req, reqerr := http.NewRequest(
		"GET",
		"https://api.github.com/user",
		nil,
	)
	if reqerr != nil {
		return nil, errors.New("api request creation failed")
	}

	// Set the Authorization header before sending the request
	// Authorization: token XXXXXXXXXXXXXXXXXXXXXXXXXXX
	authorizationHeaderValue := fmt.Sprintf("token %s", accessToken)
	req.Header.Set("Authorization", authorizationHeaderValue)
	req.Header.Set("accept", "application/vnd.github.v3+json")
	// Make the request		q
	resp, resperr := http.DefaultClient.Do(req)
	if resperr != nil {
		return nil, errors.New("request failed")
	}

	var t UserData
	if err := json.NewDecoder(resp.Body).Decode(&t); err != nil {
		fmt.Fprintf(os.Stdout, "could not parse JSON response: %v", err)
		return nil, errors.New("could not parse JSON")
	}

	// Convert byte slice to string and return
	return &t, nil
}

type UserData struct {
	Login string `json:"login"`
	Id    int    `json:"id"`
}

type resRedirect struct {
	Code   int    `json:"code"`
	ErrMSG string `json:"err_msg"`
	Data   string `json:"data"`
}

func (f *FaucetController) githubCallbackRes(code int, errMsg string, data string) {
	rd := resRedirect{Code: code, ErrMSG: errMsg, Data: data}
	jrd, _ := json.Marshal(rd)
	r := fmt.Sprintf(common.GetGithubRedirect()+"?data=%s", jrd)
	f.Redirect(r, 302)
}

