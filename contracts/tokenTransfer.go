package contracts

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/prometheus/common/log"
	"github.com/hpb-project/tokenfaucet/contracts/erc20"
	"math/big"
)

var (
	big10 = big.NewInt(10)
)

func newTransactOpts(privKey *ecdsa.PrivateKey, chainId *big.Int, nonce uint64) *bind.TransactOpts {
	transactOpts, err := bind.NewKeyedTransactorWithChainID(privKey, chainId)
	if err != nil {
		panic(err)
	}
	transactOpts.Nonce = big.NewInt(int64(nonce))

	return transactOpts
}

func TokenTransfer(ctx context.Context, url string, tokenAddr common.Address, privk string, to common.Address, amount int64, nonce uint64) (*types.Transaction, error) {
	priv, err := crypto.HexToECDSA(privk)
	if err != nil {
		log.Errorf("invalid private key")
		return nil, err
	}

	client, err := ethclient.Dial(url)
	if err != nil {
		log.Errorf("dial rpc url failed ", err)
		return nil, err
	}
	defer client.Close()

	chainId, err := client.ChainID(ctx)
	if err != nil {
		log.Errorf("get chain id failed", err)
		return nil, err
	}

	contract, err := erc20.NewErc20(tokenAddr, client)
	if err != nil {
		log.Errorf("build contract failed ", err)
		return nil, err
	}
	defaultCall := &bind.CallOpts{
		Context: ctx,
	}
	decimal, err := contract.Decimals(defaultCall)
	if err != nil {
		log.Errorf("get decimal failed ", err)
		return nil, err
	}

	unit := new(big.Int).Exp(big10, big.NewInt(int64(decimal)), nil)
	tokenValue := new(big.Int).Mul(big.NewInt(int64(amount)), unit)

	txOpt := newTransactOpts(priv, chainId, nonce)
	tx, err := contract.Transfer(txOpt, to, tokenValue)
	if err != nil {
		log.Errorf("token transfer failed err ", err)
		return nil, err
	}
	return tx, err
}

func getAddrFromPrivk(priv *ecdsa.PrivateKey) (common.Address, error) {
	publicKey := priv.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return common.Address{}, errors.New("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	return fromAddress, nil
}
