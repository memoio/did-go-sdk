package memo

import (
	"context"
	"crypto/ecdsa"
	"encoding/binary"
	"encoding/hex"
	"math/big"
	"math/rand"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/memoio/go-did/types"
	"golang.org/x/xerrors"

	com "github.com/memoio/contractsv2/common"
	inst "github.com/memoio/contractsv2/go_contracts/instance"
	"github.com/memoio/did-solidity/go-contracts/proxy"
)

var (
	checkTxSleepTime = 6 // wait 6s
	nextBlockTime    = 5 // blocktime
)

type MemoDIDController struct {
	instanceAddr  common.Address
	endpoint      string
	privateKey    *ecdsa.PrivateKey
	didTransactor *bind.TransactOpts
	proxyAddr     common.Address
}

var _ DIDController = &MemoDIDController{}

func NewMemoDIDController(privateKey *ecdsa.PrivateKey, chain string) (*MemoDIDController, error) {
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
	auth.Value = big.NewInt(0) // in wei
	auth.GasPrice = big.NewInt(2000)

	return &MemoDIDController{
		instanceAddr:  instanceAddr,
		endpoint:      endpoint,
		privateKey:    privateKey,
		didTransactor: auth,
		proxyAddr:     proxyAddr,
	}, err
}

func (c *MemoDIDController) CreateUnregisteredDID(methodType string, publicKeyBytes []byte) (*types.MemoDID, error) {
	_, err := getVerificationMethod(methodType, publicKeyBytes)
	if err != nil {
		return nil, err
	}

	salt, err := c.generateSalt("blockNumber")
	if err != nil {
		return nil, err
	}

	identifier := hex.EncodeToString(crypto.Keccak256(binary.BigEndian.AppendUint64(publicKeyBytes, salt)))

	return &types.MemoDID{
		Method:      "memo",
		Identifier:  identifier,
		Identifiers: []string{identifier},
	}, nil
}

// GetRegisteredMessage get the message to be signed by the user
// Use EIP-191 to sign the message
// See more details: https://eips.ethereum.org/EIPS/eip-191
func (c *MemoDIDController) GetRegisterMessage(did *types.MemoDID, methodType string, publicKeyBytes []byte) (string, error) {
	_, err := getVerificationMethod(methodType, publicKeyBytes)
	if err != nil {
		return "", err
	}

	var nonceBuf = make([]byte, 8)
	binary.BigEndian.PutUint64(nonceBuf, 0)

	message := string("createDID") + did.Identifier + methodType + string(publicKeyBytes) + string(nonceBuf)

	return message, nil
}

func (c *MemoDIDController) RegisterDID(did *types.MemoDID, methodType string, publicKeyBytes []byte, signature []byte) error {
	client, err := ethclient.DialContext(context.TODO(), c.endpoint)
	if err != nil {
		return err
	}
	defer client.Close()

	proxyIns, err := proxy.NewProxy(c.proxyAddr, client)
	if err != nil {
		return err
	}

	tx, err := proxyIns.CreateDID(c.didTransactor, did.Identifier, methodType, publicKeyBytes, signature, big.NewInt(1000000000000000000))
	if err != nil {
		return err
	}

	return CheckTx(c.endpoint, tx.Hash(), "RegisterDID")
}

// GetAddVerificationMethodMessage get the message to be signed by the user
// Use EIP-191 to sign the message
// See more details: https://eips.ethereum.org/EIPS/eip-191
func (c *MemoDIDController) GetAddVerificationMethodMessage(did *types.MemoDID, vtype string, controller types.MemoDID, publicKeyBytes []byte) (string, error) {
	_, err := getVerificationMethod(vtype, publicKeyBytes)
	if err != nil {
		return "", err
	}

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

	message := string("addVerificationMethod") + did.Identifier + vtype + controller.Identifier + string(publicKeyBytes) + string(nonceBuf)

	return message, nil
}

func (c *MemoDIDController) AddVerificationMethod(did *types.MemoDID, vtype string, controller types.MemoDID, publicKeyBytes []byte, signature []byte) error {
	publicKey := proxy.IAccountDidPublicKey{
		MethodType:  vtype,
		Controller:  controller.Identifier,
		PubKeyData:  publicKeyBytes,
		Deactivated: false,
	}

	client, err := ethclient.DialContext(context.TODO(), c.endpoint)
	if err != nil {
		return err
	}
	defer client.Close()

	proxyIns, err := proxy.NewProxy(c.proxyAddr, client)
	if err != nil {
		return err
	}

	tx, err := proxyIns.AddVeri(c.didTransactor, did.Identifier, publicKey, signature)
	if err != nil {
		return err
	}

	return CheckTx(c.endpoint, tx.Hash(), "AddVerificationMethod")
}

// GetUpdateVerificationMethodMessage get the message to be signed by the user
// Use EIP-191 to sign the message
// See more details: https://eips.ethereum.org/EIPS/eip-191
func (c *MemoDIDController) GetUpdateVerificationMethodMessage(didUrl types.MemoDIDUrl, vtype string, publicKeyBytes []byte) (string, error) {
	_, err := getVerificationMethod(vtype, publicKeyBytes)
	if err != nil {
		return "", err
	}

	client, err := ethclient.DialContext(context.TODO(), c.endpoint)
	if err != nil {
		return "", err
	}
	defer client.Close()

	proxyIns, err := proxy.NewProxy(c.proxyAddr, client)
	if err != nil {
		return "", err
	}

	nonce, err := proxyIns.GetNonce(&bind.CallOpts{}, didUrl.Identifier)
	if err != nil {
		return "", err
	}

	var nonceBuf = make([]byte, 8)
	binary.BigEndian.PutUint64(nonceBuf, nonce)
	var indexBuf = make([]byte, 8)
	binary.BigEndian.PutUint64(indexBuf, uint64(didUrl.GetMethodIndex()))

	message := string("updateVerificationMethod") + didUrl.Identifier + string(indexBuf) + vtype + string(publicKeyBytes) + string(nonceBuf)

	return message, nil
}

func (c *MemoDIDController) UpdateVerificationMethod(didUrl types.MemoDIDUrl, vtype string, publicKeyBytes []byte, signature []byte) error {
	client, err := ethclient.DialContext(context.TODO(), c.endpoint)
	if err != nil {
		return err
	}
	defer client.Close()

	proxyIns, err := proxy.NewProxy(c.proxyAddr, client)
	if err != nil {
		return err
	}

	tx, err := proxyIns.UpdateVeri(c.didTransactor, didUrl.Identifier, big.NewInt(int64(didUrl.GetMethodIndex())), vtype, publicKeyBytes, signature)
	if err != nil {
		return err
	}
	return CheckTx(c.endpoint, tx.Hash(), "UpdateVerificationMethod")
}

// GetDeactivateVerificationMethodMessage get the message to be signed by the user
// Use EIP-191 to sign the message
// See more details: https://eips.ethereum.org/EIPS/eip-191
func (c *MemoDIDController) GetDeactivateVerificationMethodMessage(didUrl types.MemoDIDUrl) (string, error) {
	client, err := ethclient.DialContext(context.TODO(), c.endpoint)
	if err != nil {
		return "", err
	}
	defer client.Close()

	proxyIns, err := proxy.NewProxy(c.proxyAddr, client)
	if err != nil {
		return "", err
	}

	nonce, err := proxyIns.GetNonce(&bind.CallOpts{}, didUrl.Identifier)
	if err != nil {
		return "", err
	}

	var nonceBuf = make([]byte, 8)
	binary.BigEndian.PutUint64(nonceBuf, nonce)
	var indexBuf = make([]byte, 8)
	binary.BigEndian.PutUint64(indexBuf, uint64(didUrl.GetMethodIndex()))

	message := string("deactivateVerificationMethod") + didUrl.Identifier + string(indexBuf) + string(nonceBuf)

	return message, nil
}

func (c *MemoDIDController) DeactivateVerificationMethod(didUrl types.MemoDIDUrl, signature []byte) error {
	client, err := ethclient.DialContext(context.TODO(), c.endpoint)
	if err != nil {
		return err
	}
	defer client.Close()

	proxyIns, err := proxy.NewProxy(c.proxyAddr, client)
	if err != nil {
		return err
	}

	tx, err := proxyIns.DeactivateVeri(c.didTransactor, didUrl.Identifier, big.NewInt(int64(didUrl.GetMethodIndex())), true, signature)
	if err != nil {
		return err
	}

	return CheckTx(c.endpoint, tx.Hash(), "DeactivateVerificationMethod")
}

// GetAddRelationShipMessage get the message to be signed by the user
// Use EIP-191 to sign the message
// See more details: https://eips.ethereum.org/EIPS/eip-191
func (c *MemoDIDController) GetAddRelationShipMessage(did *types.MemoDID, relationType int, didUrl types.MemoDIDUrl, expireTime int64) (string, error) {
	client, err := ethclient.DialContext(context.TODO(), c.endpoint)
	if err != nil {
		return "", err
	}
	defer client.Close()

	proxyIns, err := proxy.NewProxy(c.proxyAddr, client)
	if err != nil {
		return "", err
	}

	nonce, err := proxyIns.GetNonce(&bind.CallOpts{}, didUrl.Identifier)
	if err != nil {
		return "", err
	}

	var nonceBuf = make([]byte, 8)
	binary.BigEndian.PutUint64(nonceBuf, nonce)
	var expireBuf = make([]byte, 8)
	binary.BigEndian.PutUint64(expireBuf, uint64(expireTime))

	var message string
	switch relationType {
	case types.Authentication:
		message = string("addAuthentication") + did.Identifier + didUrl.String() + string(nonceBuf)
	case types.AssertionMethod:
		message = string("addAssertion") + did.Identifier + didUrl.String() + string(nonceBuf)
	case types.CapabilityDelegation:
		message = string("addDelegation") + did.Identifier + didUrl.String() + string(expireBuf) + string(nonceBuf)
	case types.Recovery:
		message = string("addRecovery") + did.Identifier + didUrl.String() + string(nonceBuf)
	default:
		return "", xerrors.Errorf("unsupported relation type")
	}

	return message, nil
}

func (c *MemoDIDController) AddRelationShip(did *types.MemoDID, relationType int, didUrl types.MemoDIDUrl, expireTime int64, signature []byte) error {
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
	case types.Authentication:
		tx, err = proxyIns.AddAuth(c.didTransactor, did.Identifier, didUrl.String(), signature)
	case types.AssertionMethod:
		tx, err = proxyIns.AddAssertion(c.didTransactor, did.Identifier, didUrl.String(), signature)
	case types.CapabilityDelegation:
		tx, err = proxyIns.AddDelegation(c.didTransactor, did.Identifier, didUrl.String(), big.NewInt(expireTime+time.Now().Unix()), signature)
	case types.Recovery:
		tx, err = proxyIns.AddRecovery(c.didTransactor, did.Identifier, didUrl.String(), signature)
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
func (c *MemoDIDController) GetDeactivateRelationShipMessage(did *types.MemoDID, relationType int, didUrl types.MemoDIDUrl) (string, error) {
	client, err := ethclient.DialContext(context.TODO(), c.endpoint)
	if err != nil {
		return "", err
	}
	defer client.Close()

	proxyIns, err := proxy.NewProxy(c.proxyAddr, client)
	if err != nil {
		return "", err
	}

	nonce, err := proxyIns.GetNonce(&bind.CallOpts{}, didUrl.Identifier)
	if err != nil {
		return "", err
	}

	var nonceBuf = make([]byte, 8)
	binary.BigEndian.PutUint64(nonceBuf, nonce)

	var message string
	switch relationType {
	case types.Authentication:
		message = string("deactivateAuthentication") + did.Identifier + didUrl.String() + string(nonceBuf)
	case types.AssertionMethod:
		message = string("deactivateAssertion") + did.Identifier + didUrl.String() + string(nonceBuf)
	case types.CapabilityDelegation:
		message = string("deactivateDelegation") + did.Identifier + didUrl.String() + string(nonceBuf)
	case types.Recovery:
		message = string("deactivateRecovery") + did.Identifier + didUrl.String() + string(nonceBuf)
	default:
		return "", xerrors.Errorf("unsupported relation type")
	}

	return message, nil
}

func (c *MemoDIDController) DeactivateRelationShip(did *types.MemoDID, relationType int, didUrl types.MemoDIDUrl, signature []byte) error {
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
	case types.Authentication:
		tx, err = proxyIns.RemoveAuth(c.didTransactor, did.Identifier, didUrl.String(), signature)
	case types.AssertionMethod:
		tx, err = proxyIns.RemoveAssertion(c.didTransactor, did.Identifier, didUrl.String(), signature)
	case types.CapabilityDelegation:
		tx, err = proxyIns.RemoveDelegation(c.didTransactor, did.Identifier, didUrl.String(), signature)
	case types.Recovery:
		tx, err = proxyIns.RemoveRecovery(c.didTransactor, did.Identifier, didUrl.String(), signature)
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
func (c *MemoDIDController) GetDeactivateDIDMessage(did *types.MemoDID) (string, error) {
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

	message := string("deactivateDID") + did.Identifier + string(nonceBuf)

	return message, nil
}

func (c *MemoDIDController) DeactivateDID(did *types.MemoDID, signature []byte) error {
	client, err := ethclient.DialContext(context.TODO(), c.endpoint)
	if err != nil {
		return err
	}
	defer client.Close()

	proxyIns, err := proxy.NewProxy(c.proxyAddr, client)
	if err != nil {
		return err
	}

	tx, err := proxyIns.DeactivateDID(c.didTransactor, did.Identifier, true, signature)
	if err != nil {
		return err
	}

	return CheckTx(c.endpoint, tx.Hash(), "DeactivateDID")
}

// func (c *MemoDIDController) ApproveOfMfileContract(amount int) error {
// 	client, err := ethclient.DialContext(context.TODO(), c.endpoint)
// 	if err != nil {
// 		return err
// 	}
// 	defer client.Close()

// 	instanceIns, err := inst.NewInstance(c.instanceAddr, client)
// 	if err != nil {
// 		return err
// 	}

// 	// get fileDIDCtrAddr
// 	fileDIDCtrAddr, err := instanceIns.Instances(&bind.CallOpts{}, com.TypeFileDidControl)
// 	if err != nil {
// 		return err
// 	}

// 	// get ERC20Addr
// 	ERC20Addr, err := instanceIns.Instances(&bind.CallOpts{}, com.TypeERC20)
// 	if err != nil {
// 		return err
// 	}

// 	// get ERC20Ins
// 	ERC20Ins, err := erc.NewERC20(ERC20Addr, client)
// 	if err != nil {
// 		return err
// 	}

// 	tx, err := ERC20Ins.Approve(c.didTransactor, fileDIDCtrAddr, big.NewInt(int64(amount)))
// 	if err != nil {
// 		return err
// 	}

// 	return CheckTx(c.endpoint, tx.Hash(), "ApproveOfMfileContract")
// }

// func (c *MemoDIDController) BuyReadPermission(mfileDID types.MfileDID, memoDID *types.MemoDID) error {
// 	client, err := ethclient.DialContext(context.TODO(), c.endpoint)
// 	if err != nil {
// 		return err
// 	}
// 	defer client.Close()

// 	proxyIns, err := proxy.NewProxy(c.proxyAddr, client)
// 	if err != nil {
// 		return err
// 	}

// 	tx, err := proxyIns.BuyRead(c.didTransactor, mfileDID.Identifier, memoDID.Identifier)
// 	if err != nil {
// 		return err
// 	}

// 	return CheckTx(c.endpoint, tx.Hash(), "BuyReadPermission")
// }

// CheckTx check whether transaction is successful through receipt
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

// generateSalt generate a random salt
// 1. use crypto.Rand to generate a random salt
// 2. use pending nonce to generate a salt
// 3. use the block number to generate a salt
func (c *MemoDIDController) generateSalt(genType string) (uint64, error) {
	client, err := ethclient.DialContext(context.TODO(), c.endpoint)
	if err != nil {
		return 0, err
	}
	defer client.Close()

	switch genType {
	case "random":
		salt := make([]byte, 32)
		_, err = rand.Read(salt)
		if err != nil {
			return 0, err
		}
		return uint64(binary.BigEndian.Uint64(salt)), nil
	case "pendingNonce":
		address := crypto.PubkeyToAddress(c.privateKey.PublicKey)
		return client.PendingNonceAt(context.TODO(), address)
	case "blockNumber":
		return client.BlockNumber(context.TODO())
	}
	return 0, xerrors.Errorf("unsupported generate type: %s", genType)
}

func getVerificationMethod(methodType string, publicKeyData []byte) (*types.VerificationMethod, error) {
	switch methodType {
	case "EcdsaSecp256k1VerificationKey2019":
		return &types.VerificationMethod{
			PublicKey: types.PublicKey{
				Type:         methodType,
				PublicKeyHex: hex.EncodeToString(publicKeyData),
			},
		}, nil
	case "Ed25519VerificationKey2018":
		return &types.VerificationMethod{
			PublicKey: types.PublicKey{
				Type:         methodType,
				PublicKeyHex: hex.EncodeToString(publicKeyData),
			},
		}, nil
	case "EcdsaSecp256k1RecoveryMethod2020":
		return &types.VerificationMethod{
			PublicKey: types.PublicKey{
				Type:         methodType,
				PublicKeyHex: hex.EncodeToString(publicKeyData),
			},
		}, nil
	default:
		return nil, xerrors.Errorf("unsupported method type: %s", methodType)
	}
}
