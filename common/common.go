package common

import (
	"crypto/ecdsa"
	"errors"
	"github.com/astaxie/beego/context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
	"net"
	"strings"
)

func TransferAmount(amount *big.Int) *big.Int {
	n := new(big.Int)
	na, _ := n.SetString("1000000000000000000", 0)
	return na.Mul(na, amount)
}

func CheckAddress(address string) bool {
	if !strings.HasPrefix(address, "0x") {
		return false
	}
	return common.IsHexAddress(address)
}

func GetClientIP(ctx *context.Context) string {
	r := ctx.Request
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	ip := strings.TrimSpace(strings.Split(xForwardedFor, ",")[0])
	if ip != "" {
		return ip
	}
	ip = strings.TrimSpace(r.Header.Get("X-Real-Ip"))
	if ip != "" {
		return ip
	}
	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		return ip
	}
	return ""
}


func GetAddrFromPrivk(privk string) (string, error) {
	if strings.HasPrefix(privk, "0x") || strings.HasPrefix(privk, "0X") {
		privk = privk[2:]
	}

	priv, err := crypto.HexToECDSA(privk)
	if err != nil {
		return "", err
	}
	publicKey := priv.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", errors.New("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	return fromAddress.String(), nil
}
