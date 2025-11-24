package mfile

import (
	"context"
	"crypto/ecdsa"
	"encoding/binary"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	com "github.com/memoio/contractsv2/common"
	inst "github.com/memoio/contractsv2/go_contracts/instance"
	"github.com/memoio/did-solidity/go-contracts/proxy"
	"github.com/memoio/go-did/types"
	"golang.org/x/xerrors"
)

var (
	checkTxSleepTime = 6 // wait 6s(blocktime + 1s)
	nextBlockTime    = 5 // blocktime
)

type MfileDIDController struct {
	did           *types.MfileDID
	endpoint      string
	privateKey    *ecdsa.PrivateKey
	didTransactor *bind.TransactOpts
	proxyAddr     common.Address
}

var _ MfileStore = &MfileDIDController{}

func NewMfileDIDController(privateKey *ecdsa.PrivateKey, chain, didString string) (*MfileDIDController, error) {
	instanceAddr, endpoint := com.GetInsEndPointByChain(chain)

	client, err := ethclient.DialContext(context.TODO(), endpoint)
	if err != nil {
		return nil, err
	}

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		chainID = big.NewInt(985)
	}

	// new instanceIns
	instanceIns, err := inst.NewInstance(instanceAddr, client)
	if err != nil {
		return nil, err
	}

	// get proxyAddr
	proxyAddr, err := instanceIns.Instances(&bind.CallOpts{}, com.TypeDidProxy)
	if err != nil {
		return nil, err
	}

	// new auth
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		return nil, err
	}
	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(300000) // in units
	auth.GasPrice = big.NewInt(1000)

	did, err := types.ParseMfileDID(didString)
	return &MfileDIDController{
		did:           did,
		endpoint:      endpoint,
		privateKey:    privateKey,
		didTransactor: auth,
		proxyAddr:     proxyAddr,
	}, err
}

func (c *MfileDIDController) DID() *types.MfileDID {
	return c.did
}

// GetRegisterDIDMessage get the message to be signed by the user
// Use EIP-191 to sign the message
// See more details: https://eips.ethereum.org/EIPS/eip-191
func (c *MfileDIDController) GetRegisterDIDMessage(did *types.MfileDID, encode string, ftype uint8, price *big.Int, keywords []string, controller types.MemoDID) (string, error) {
	client, err := ethclient.DialContext(context.TODO(), c.endpoint)
	if err != nil {
		return "", err
	}
	defer client.Close()

	proxyIns, err := proxy.NewProxy(c.proxyAddr, client)
	if err != nil {
		return "", err
	}

	nonce, err := proxyIns.GetNonce(&bind.CallOpts{}, did.Identifier)
	if err != nil {
		return "", err
	}

	var nonceBuf = make([]byte, 8)
	binary.BigEndian.PutUint64(nonceBuf, nonce)
	var ftypeBuf = make([]byte, 8)
	binary.BigEndian.PutUint64(ftypeBuf, uint64(ftype))
	var priceBuf = make([]byte, 32)
	price.FillBytes(priceBuf)

	keywordsStr := ""
	for _, kw := range keywords {
		keywordsStr += kw
	}

	message := string("registerMfileDID") + did.Identifier + encode + string(ftypeBuf) + string(priceBuf) + keywordsStr + controller.Identifier + string(nonceBuf)

	return message, nil
}

func (c *MfileDIDController) RegisterDID(did *types.MfileDID, encode string, ftype uint8, price *big.Int, keywords []string, controller types.MemoDID, signature []byte) error {
	client, err := ethclient.DialContext(context.TODO(), c.endpoint)
	if err != nil {
		return err
	}
	defer client.Close()

	proxyIns, err := proxy.NewProxy(c.proxyAddr, client)
	if err != nil {
		return err
	}

	tx, err := proxyIns.RegisterMfileDid(c.didTransactor, did.Identifier, encode, ftype, controller.Identifier, price, keywords, signature)
	if err != nil {
		return err
	}

	return CheckTx(c.endpoint, tx.Hash(), "RegisterDID")
}

// GetChangeControllerMessage get the message to be signed by the user
// Use EIP-191 to sign the message
// See more details: https://eips.ethereum.org/EIPS/eip-191
func (c *MfileDIDController) GetChangeControllerMessage(did *types.MfileDID, controller types.MemoDID) (string, error) {
	client, err := ethclient.DialContext(context.TODO(), c.endpoint)
	if err != nil {
		return "", err
	}
	defer client.Close()

	proxyIns, err := proxy.NewProxy(c.proxyAddr, client)
	if err != nil {
		return "", err
	}

	nonce, err := proxyIns.GetNonce(&bind.CallOpts{}, did.Identifier)
	if err != nil {
		return "", err
	}

	var nonceBuf = make([]byte, 8)
	binary.BigEndian.PutUint64(nonceBuf, nonce)

	message := string("changeController") + did.Identifier + controller.Identifier + string(nonceBuf)

	return message, nil
}

func (c *MfileDIDController) ChangeController(did *types.MfileDID, controller types.MemoDID, signature []byte) error {
	client, err := ethclient.DialContext(context.TODO(), c.endpoint)
	if err != nil {
		return err
	}
	defer client.Close()

	proxyIns, err := proxy.NewProxy(c.proxyAddr, client)
	if err != nil {
		return err
	}

	tx, err := proxyIns.ChangeController(c.didTransactor, did.Identifier, controller.Identifier, signature)
	if err != nil {
		return err
	}

	return CheckTx(c.endpoint, tx.Hash(), "ChangeController")
}

// GetChangeFileTypeMessage get the message to be signed by the user
// Use EIP-191 to sign the message
// See more details: https://eips.ethereum.org/EIPS/eip-191
func (c *MfileDIDController) GetChangeFileTypeMessage(did *types.MfileDID, ftype uint8) (string, error) {
	client, err := ethclient.DialContext(context.TODO(), c.endpoint)
	if err != nil {
		return "", err
	}
	defer client.Close()

	proxyIns, err := proxy.NewProxy(c.proxyAddr, client)
	if err != nil {
		return "", err
	}

	nonce, err := proxyIns.GetNonce(&bind.CallOpts{}, did.Identifier)
	if err != nil {
		return "", err
	}

	var nonceBuf = make([]byte, 8)
	binary.BigEndian.PutUint64(nonceBuf, nonce)
	var ftypeBuf = make([]byte, 8)
	binary.BigEndian.PutUint64(ftypeBuf, uint64(ftype))

	message := string("changeFileType") + did.Identifier + string(ftypeBuf) + string(nonceBuf)

	return message, nil
}

func (c *MfileDIDController) ChangeFileType(did *types.MfileDID, ftype uint8, signature []byte) error {
	client, err := ethclient.DialContext(context.TODO(), c.endpoint)
	if err != nil {
		return err
	}
	defer client.Close()

	proxyIns, err := proxy.NewProxy(c.proxyAddr, client)
	if err != nil {
		return err
	}

	tx, err := proxyIns.ChangeFtype(c.didTransactor, did.Identifier, ftype, signature)
	if err != nil {
		return err
	}

	return CheckTx(c.endpoint, tx.Hash(), "ChangeFileType")
}

// GetChangePriceMessage get the message to be signed by the user
// Use EIP-191 to sign the message
// See more details: https://eips.ethereum.org/EIPS/eip-191
func (c *MfileDIDController) GetChangePriceMessage(did *types.MfileDID, price *big.Int) (string, error) {
	client, err := ethclient.DialContext(context.TODO(), c.endpoint)
	if err != nil {
		return "", err
	}
	defer client.Close()

	proxyIns, err := proxy.NewProxy(c.proxyAddr, client)
	if err != nil {
		return "", err
	}

	nonce, err := proxyIns.GetNonce(&bind.CallOpts{}, did.Identifier)
	if err != nil {
		return "", err
	}

	var nonceBuf = make([]byte, 8)
	binary.BigEndian.PutUint64(nonceBuf, nonce)
	var priceBuf = make([]byte, 32)
	price.FillBytes(priceBuf)

	message := string("changePrice") + did.Identifier + string(priceBuf) + string(nonceBuf)

	return message, nil
}

func (c *MfileDIDController) ChangePrice(did *types.MfileDID, price *big.Int, signature []byte) error {
	client, err := ethclient.DialContext(context.TODO(), c.endpoint)
	if err != nil {
		return err
	}
	defer client.Close()

	proxyIns, err := proxy.NewProxy(c.proxyAddr, client)
	if err != nil {
		return err
	}

	tx, err := proxyIns.ChangePrice(c.didTransactor, did.Identifier, price, signature)
	if err != nil {
		return err
	}

	return CheckTx(c.endpoint, tx.Hash(), "ChangePrice")
}

// GetChangeKeywordsMessage get the message to be signed by the user
// Use EIP-191 to sign the message
// See more details: https://eips.ethereum.org/EIPS/eip-191
func (c *MfileDIDController) GetChangeKeywordsMessage(did *types.MfileDID, keywords []string) (string, error) {
	client, err := ethclient.DialContext(context.TODO(), c.endpoint)
	if err != nil {
		return "", err
	}
	defer client.Close()

	proxyIns, err := proxy.NewProxy(c.proxyAddr, client)
	if err != nil {
		return "", err
	}

	nonce, err := proxyIns.GetNonce(&bind.CallOpts{}, did.Identifier)
	if err != nil {
		return "", err
	}

	var nonceBuf = make([]byte, 8)
	binary.BigEndian.PutUint64(nonceBuf, nonce)

	keywordsStr := ""
	for _, kw := range keywords {
		keywordsStr += kw
	}

	message := string("changeKeywords") + did.Identifier + keywordsStr + string(nonceBuf)

	return message, nil
}

func (c *MfileDIDController) ChangeKeywords(did *types.MfileDID, keywords []string, signature []byte) error {
	client, err := ethclient.DialContext(context.TODO(), c.endpoint)
	if err != nil {
		return err
	}
	defer client.Close()

	proxyIns, err := proxy.NewProxy(c.proxyAddr, client)
	if err != nil {
		return err
	}

	tx, err := proxyIns.ChangeKeywords(c.didTransactor, did.Identifier, keywords, signature)
	if err != nil {
		return err
	}

	return CheckTx(c.endpoint, tx.Hash(), "ChangeKeywords")
}

// GetAddRelationShipMessage get the message to be signed by the user
// Use EIP-191 to sign the message
// See more details: https://eips.ethereum.org/EIPS/eip-191
func (c *MfileDIDController) GetAddRelationShipMessage(did *types.MfileDID, relationType int, memoDID types.MemoDID) (string, error) {
	client, err := ethclient.DialContext(context.TODO(), c.endpoint)
	if err != nil {
		return "", err
	}
	defer client.Close()

	proxyIns, err := proxy.NewProxy(c.proxyAddr, client)
	if err != nil {
		return "", err
	}

	nonce, err := proxyIns.GetNonce(&bind.CallOpts{}, did.Identifier)
	if err != nil {
		return "", err
	}

	var nonceBuf = make([]byte, 8)
	binary.BigEndian.PutUint64(nonceBuf, nonce)

	var message string
	switch relationType {
	case types.Read:
		message = string("grantRead") + did.Identifier + memoDID.Identifier + string(nonceBuf)
	default:
		return "", xerrors.Errorf("unsupported relation type")
	}

	return message, nil
}

func (c *MfileDIDController) AddRelationShip(did *types.MfileDID, relationType int, memoDID types.MemoDID, signature []byte) error {
	client, err := ethclient.DialContext(context.TODO(), c.endpoint)
	if err != nil {
		return err
	}
	defer client.Close()

	proxyIns, err := proxy.NewProxy(c.proxyAddr, client)
	if err != nil {
		return err
	}

	var tx *etypes.Transaction
	switch relationType {
	case types.Read:
		tx, err = proxyIns.GrantRead(c.didTransactor, did.Identifier, memoDID.Identifier, signature)
	default:
		return xerrors.Errorf("unsupported relation ships")
	}
	if err != nil {
		return err
	}

	return CheckTx(c.endpoint, tx.Hash(), "AddRelationShip")
}

// GetDeactivateRelationShipMessage get the message to be signed by the user
// Use EIP-191 to sign the message
// See more details: https://eips.ethereum.org/EIPS/eip-191
func (c *MfileDIDController) GetDeactivateRelationShipMessage(did *types.MfileDID, relationType int, memoDID types.MemoDID) (string, error) {
	client, err := ethclient.DialContext(context.TODO(), c.endpoint)
	if err != nil {
		return "", err
	}
	defer client.Close()

	proxyIns, err := proxy.NewProxy(c.proxyAddr, client)
	if err != nil {
		return "", err
	}

	nonce, err := proxyIns.GetNonce(&bind.CallOpts{}, did.Identifier)
	if err != nil {
		return "", err
	}

	var nonceBuf = make([]byte, 8)
	binary.BigEndian.PutUint64(nonceBuf, nonce)

	var message string
	switch relationType {
	case types.Read:
		message = string("deactivateRead") + did.Identifier + memoDID.Identifier + string(nonceBuf)
	default:
		return "", xerrors.Errorf("unsupported relation type")
	}

	return message, nil
}

func (c *MfileDIDController) DeactivateRelationShip(did *types.MfileDID, relationType int, memoDID types.MemoDID, signature []byte) error {
	client, err := ethclient.DialContext(context.TODO(), c.endpoint)
	if err != nil {
		return err
	}
	defer client.Close()

	proxyIns, err := proxy.NewProxy(c.proxyAddr, client)
	if err != nil {
		return err
	}

	var tx *etypes.Transaction
	switch relationType {
	case types.Read:
		tx, err = proxyIns.DeactivateRead(c.didTransactor, did.Identifier, memoDID.Identifier, signature)
	default:
		return xerrors.Errorf("unsupported relation ships")
	}
	if err != nil {
		return err
	}

	return CheckTx(c.endpoint, tx.Hash(), "DeactivateRelationShip")
}

// GetDeactivateDIDMessage get the message to be signed by the user
// Use EIP-191 to sign the message
// See more details: https://eips.ethereum.org/EIPS/eip-191
func (c *MfileDIDController) GetDeactivateDIDMessage(did *types.MfileDID) (string, error) {
	client, err := ethclient.DialContext(context.TODO(), c.endpoint)
	if err != nil {
		return "", err
	}
	defer client.Close()

	proxyIns, err := proxy.NewProxy(c.proxyAddr, client)
	if err != nil {
		return "", err
	}

	nonce, err := proxyIns.GetNonce(&bind.CallOpts{}, did.Identifier)
	if err != nil {
		return "", err
	}

	var nonceBuf = make([]byte, 8)
	binary.BigEndian.PutUint64(nonceBuf, nonce)

	message := string("deactivateMfileDID") + did.Identifier + string(nonceBuf)

	return message, nil
}

func (c *MfileDIDController) DeactivateDID(did *types.MfileDID, signature []byte) error {
	client, err := ethclient.DialContext(context.TODO(), c.endpoint)
	if err != nil {
		return err
	}
	defer client.Close()

	proxyIns, err := proxy.NewProxy(c.proxyAddr, client)
	if err != nil {
		return err
	}

	tx, err := proxyIns.DeactivateMfileDid(c.didTransactor, did.Identifier, true, signature)
	if err != nil {
		return err
	}

	return CheckTx(c.endpoint, tx.Hash(), "DeactivateDID")
}

func CheckTx(endPoint string, txHash common.Hash, name string) error {
	var receipt *etypes.Receipt

	t := checkTxSleepTime
	for i := 0; i < 10; i++ {
		time.Sleep(time.Duration(t) * time.Second)
		receipt = com.GetTransactionReceipt(endPoint, txHash)
		if receipt != nil {
			break
		}
		t = nextBlockTime
	}

	if receipt == nil {
		return xerrors.Errorf("%s: cann't get transaction(%s) receipt, not packaged", name, txHash)
	}

	// 0 means fail
	if receipt.Status == 0 {
		if receipt.GasUsed != receipt.CumulativeGasUsed {
			return xerrors.Errorf("%s: transaction(%s) exceed gas limit", name, txHash)
		}
		return xerrors.Errorf("%s: transaction(%s) mined but execution failed, please check your tx input", name, txHash)
	}
	return nil
}
