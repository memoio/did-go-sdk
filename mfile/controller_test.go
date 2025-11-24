package mfile

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ipfs/go-cid"
	com "github.com/memoio/contractsv2/common"
	inst "github.com/memoio/contractsv2/go_contracts/instance"
	"github.com/memoio/go-did/memo"
	"github.com/memoio/go-did/types"
	"github.com/multiformats/go-multicodec"
	mh "github.com/multiformats/go-multihash"
	"golang.org/x/xerrors"
)

var globalPrivateKeys []string

func init() {
	content, err := ioutil.ReadFile("../key.json")
	if err != nil {
		log.Fatal(err.Error())
	}

	err = json.Unmarshal(content, &globalPrivateKeys)
	if err != nil {
		log.Fatal(err.Error())
	}

	// log.Println(globalPrivateKeys)

}

func ToPublicKey(privateKeyHex string) (string, error) {
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return "", err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", xerrors.Errorf("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	return hex.EncodeToString(crypto.CompressPubkey(publicKeyECDSA)), nil
}

func ToPublicKeys(privateKeyHex []string) ([]*ecdsa.PrivateKey, []string, error) {
	var sks []*ecdsa.PrivateKey
	var pks []string
	for _, sk := range privateKeyHex {
		publicKey, err := ToPublicKey(sk)
		if err != nil {
			return nil, nil, err
		}

		privateKey, err := crypto.HexToECDSA(sk)
		if err != nil {
			return nil, nil, err
		}

		sks = append(sks, privateKey)
		pks = append(pks, publicKey)
	}

	return sks, pks, nil
}

func genCid() cid.Cid {
	data := make([]byte, 32)
	rand.Read(data)

	prifix := cid.NewPrefixV1(uint64(multicodec.Raw), mh.SHA2_256)
	cid, _ := prifix.Sum(data)

	return cid
}

func TestBasic(t *testing.T) {
	sks, _, err := ToPublicKeys(globalPrivateKeys)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Used for signing transactions and paying gas fees for users.
	masterKey1 := sks[0]
	// Used for signing user operations, the owner of the DID.
	userSK := sks[1]

	cid := genCid()
	mfileDIDString := "did:mfile:" + cid.String()

	resolver, _ := NewMfileDIDResolver("dev")

	mfilecontroller, err := NewMfileDIDController(masterKey1, "dev", mfileDIDString)
	if err != nil {
		t.Fatal(err.Error())
	}

	memocontroller, err := memo.NewMemoDIDController(masterKey1, "dev")
	if err != nil {
		t.Fatal(err.Error())
	}

	// Create memo DID first
	memoDID, err := memocontroller.CreateUnregisteredDID("EcdsaSecp256k1VerificationKey2019", crypto.CompressPubkey(&userSK.PublicKey))
	if err != nil {
		t.Fatal(err.Error())
	}

	// Register memo DID
	message, err := memocontroller.GetRegisterMessage(memoDID, "EcdsaSecp256k1VerificationKey2019", crypto.CompressPubkey(&userSK.PublicKey))
	if err != nil {
		t.Fatal(err.Error())
	}

	hash := crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err := crypto.Sign(hash, userSK)
	if err != nil {
		t.Fatal(err.Error())
	}

	err = memocontroller.RegisterDID(memoDID, "EcdsaSecp256k1VerificationKey2019", crypto.CompressPubkey(&userSK.PublicKey), signature)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Try to register mfile DID before controller is registered (should fail)
	mfileDID, err := types.ParseMfileDID(mfileDIDString)
	if err != nil {
		t.Fatal(err.Error())
	}

	message, err = mfilecontroller.GetRegisterDIDMessage(mfileDID, "cid", 0, big.NewInt(10), []string{"test"}, *memoDID)
	if err == nil {
		hash = crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
		signature, _ = crypto.Sign(hash, userSK)
		err = mfilecontroller.RegisterDID(mfileDID, "cid", 0, big.NewInt(10), []string{"test"}, *memoDID, signature)
		if err == nil {
			t.Fatal("should report an error when controller is not registered")
		}
	}

	// Now register mfile DID (should succeed)
	message, err = mfilecontroller.GetRegisterDIDMessage(mfileDID, "cid", 0, big.NewInt(10), []string{"test"}, *memoDID)
	if err != nil {
		t.Fatal(err.Error())
	}

	hash = crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err = crypto.Sign(hash, userSK)
	if err != nil {
		t.Fatal(err.Error())
	}

	err = mfilecontroller.RegisterDID(mfileDID, "cid", 0, big.NewInt(10), []string{"test"}, *memoDID, signature)
	if err != nil {
		t.Fatal("should not report an error when controller is registered: " + err.Error())
	}

	// Change price
	message, err = mfilecontroller.GetChangePriceMessage(mfileDID, big.NewInt(50))
	if err != nil {
		t.Fatal(err.Error())
	}

	hash = crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err = crypto.Sign(hash, userSK)
	if err != nil {
		t.Fatal(err.Error())
	}

	err = mfilecontroller.ChangePrice(mfileDID, big.NewInt(50), signature)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Add relation ship
	message, err = mfilecontroller.GetAddRelationShipMessage(mfileDID, types.Read, *memoDID)
	if err != nil {
		t.Fatal(err.Error())
	}

	hash = crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err = crypto.Sign(hash, userSK)
	if err != nil {
		t.Fatal(err.Error())
	}

	err = mfilecontroller.AddRelationShip(mfileDID, types.Read, *memoDID, signature)
	if err != nil {
		t.Fatal(err.Error())
	}

	document, err := resolver.Resolve(mfilecontroller.DID().String())
	if err != nil {
		t.Fatal(err.Error())
	}

	data, err := json.MarshalIndent(document, "", "\t")
	if err != nil {
		t.Error(err.Error())
		return
	}

	t.Log(string(data))
}

func TestResolve(t *testing.T) {
	resolver, _ := NewMfileDIDResolver("dev")
	// document, err := resolver.Resolve("did:mfile:bafkreic7emp2v6ofwkpiiqmrbjq2m6sgyws4eyq5jbphqiywkqyxzbags4")
	document, err := resolver.Resolve("did:mfile:bafkreidujh6bvgbe2wkgqbjmfhixqtywrynrxqkokgwirsnnpslymgcr6a")
	if err != nil {
		t.Fatal(err.Error())
	}

	data, err := json.MarshalIndent(document, "", "\t")
	if err != nil {
		t.Error(err.Error())
		return
	}

	t.Log(string(data))
}

func TestBuyReadPermission(t *testing.T) {
	// sks, _, err := ToPublicKeys(globalPrivateKeys)
	// if err != nil {
	// 	t.Fatal(err.Error())
	// }

	resolver, err := NewMfileDIDResolver("dev")
	if err != nil {
		t.Fatal(err.Error())
	}

	// Note: BuyReadPermission is commented out in memo controller
	// This test may need to be updated based on the actual implementation
	// controller, err := memo.NewMemoDIDController(sks[1], "dev")
	// if err != nil {
	// 	t.Fatal(err.Error())
	// }

	// memoDID, err := types.ParseMemoDID("did:memo:deb3d9ca231caca8c03edad42d03c4ccb7ddd8eae81373267c3b484cc62a8d13")
	// if err != nil {
	// 	t.Fatal(err.Error())
	// }

	// did, _ := types.ParseMfileDID("did:mfile:bafkreic7emp2v6ofwkpiiqmrbjq2m6sgyws4eyq5jbphqiywkqyxzbags4")

	// Note: ApproveOfMfileContract and BuyReadPermission are commented out in memo controller
	// Uncomment these if they are re-enabled
	// err = controller.ApproveOfMfileContract(1000)
	// if err != nil {
	// 	t.Fatal(err.Error())
	// }

	// err = controller.BuyReadPermission(*did, memoDID)
	// if err != nil {
	// 	t.Fatal(err.Error())
	// }

	document, err := resolver.Resolve("did:mfile:bafkreic7emp2v6ofwkpiiqmrbjq2m6sgyws4eyq5jbphqiywkqyxzbags4")
	if err != nil {
		t.Fatal(err.Error())
	}

	data, err := json.MarshalIndent(document, "", "\t")
	if err != nil {
		t.Error(err.Error())
		return
	}

	t.Log(string(data))
}

func TestGetInstanceAddress(t *testing.T) {
	instanceAddr, endpoint := com.GetInsEndPointByChain("dev")

	client, err := ethclient.DialContext(context.TODO(), endpoint)
	if err != nil {
		t.Error(err.Error())
		return
	}

	num, err := client.BlockNumber(context.TODO())
	if err != nil {
		t.Error(err.Error())
		return
	}
	t.Log(num)

	// new instanceIns
	instanceIns, err := inst.NewInstance(instanceAddr, client)
	if err != nil {
		t.Error(err.Error())
		return
	}

	address, err := instanceIns.Instances(&bind.CallOpts{}, com.TypeDidProxy)
	if err != nil {
		t.Error(err.Error())
		return
	}
	t.Log(address)

	// // get ERC20Addr
	// Erc20Addr, err := instanceIns.Instances(&bind.CallOpts{}, com.TypeERC20)
	// if err != nil {
	// 	t.Error(err.Error())
	// 	return
	// }

	// t.Log(instanceAddr)

	// Erc20Ins, err := erc.NewERC20(Erc20Addr, client)
	// if err != nil {
	// 	t.Error(err.Error())
	// 	return
	// }

	// balance, err := Erc20Ins.BalanceOf(&bind.CallOpts{}, common.HexToAddress("0x7C0491aE63e3816F96B777340b1571feA7bB21dE"))
	// if err != nil {
	// 	t.Error(err.Error())
	// 	return
	// }

	// t.Log(balance)
}
