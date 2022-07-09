package ethrpc

import (
	"context"
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/prometheus/common/log"
	"github.com/hpb-project/tokenfaucet/types"
	"math/big"
	"strconv"
)

var (
	big10 = big.NewInt(10)
)

func newTransactOpts(privKey *ecdsa.PrivateKey, chainId *big.Int) *bind.TransactOpts {
	transactOpts, err := bind.NewKeyedTransactorWithChainID(privKey, chainId)
	if err != nil {
		panic(err)
	}

	return transactOpts
}

func GetTransactionReceipt(network *types.Network, hash string) (*ethtypes.Receipt, error) {
	client, err := ethclient.Dial(network.Url)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	txhash := common.HexToHash(hash)
	return client.TransactionReceipt(context.Background(), txhash)
}

func EthGetTransactionCount(network *types.Network, addr string) (uint64, error) {
	address := common.HexToAddress(addr)
	client, err := ethclient.Dial(network.Url)
	if err != nil {
		return 0, err
	}
	defer client.Close()
	return client.NonceAt(context.Background(), address, nil)
}

func EthTransfer(network *types.Network, privk string, toAddress common.Address, amount int64) (*ethtypes.Transaction, error) {
	client, err := ethclient.Dial(network.Url)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	priv, err := crypto.HexToECDSA(privk)
	if err != nil {
		log.Errorf("invalid private key")
		return nil, err
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		gasPrice = big.NewInt(500000000000)
	}

	nonce := GetNonce(network)

	unit := new(big.Int).Exp(big10, big.NewInt(18), nil)
	tokenValue := new(big.Int).Mul(big.NewInt(int64(amount)), unit)
	signer := ethtypes.NewEIP155Signer(big.NewInt(int64(network.Chainid)))
	tx := ethtypes.NewTransaction(uint64(nonce), toAddress, tokenValue, uint64(network.Gaslimit), gasPrice, nil)
	signTx, err := ethtypes.SignTx(tx, signer, priv)
	if err != nil {
		return nil, err
	}
	err = client.SendTransaction(context.Background(), signTx)
	if err != nil {
		return nil, err
	}
	return signTx, nil
}

func GetBalance(network *types.Network, addr string) (*big.Int, error) {
	address := common.HexToAddress(addr)
	client, err := ethclient.Dial(network.Url)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	balance, err := client.BalanceAt(context.Background(), address, nil)
	return balance, err
}

func Hex2Dec(val string) uint64 {
	val = val[2:]
	n, _ := strconv.ParseUint(val, 16, 32)
	return n
}

func GetNonce(n *types.Network) uint64 {
	n.Mux.Lock()
	defer n.Mux.Unlock()

	onchainNonce, err := EthGetTransactionCount(n, n.Address)
	if err == nil && n.Nonce < onchainNonce {
		n.Nonce = onchainNonce
	}
	nc := n.Nonce
	n.Nonce += 1
	return nc
}
