package memo

import (
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"reflect"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/memoio/go-did/types"
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

	log.Println(globalPrivateKeys)

}

// var address1 string = "0xe89971bfeEA7381d47fE608d676dfb5440F0fD2E"
// var address2 string = "0x7C0491aE63e3816F96B777340b1571feA7bB21dE"
// var address3 string = "0xc0FF8898729d543c197Fb8b8ef7EE2f39024e1e8"
// var address4 string = "0x53F76F77DeC24D601ad3001114C9a35EfD4A5F5F"
// var address5 string = "0xC44F1bccDb80F266b727c5B3f2839AA3a2FEf1d1"

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

func TestGetPK(t *testing.T) {
	_, pks, _ := ToPublicKeys(globalPrivateKeys)
	t.Log(pks)
}

func genVerificationMethod(did *types.MemoDID, methodIndex int64, controller *types.MemoDID, vtype, publicKeyHex string) (types.VerificationMethod, error) {
	if controller == nil {
		controller, _ = types.ParseMemoDID("did:memo:0000000000000000000000000000000000000000000000000000000000000000")
	}

	didUrl, err := did.DIDUrl(methodIndex)
	if err != nil {
		return types.VerificationMethod{}, err
	}

	if publicKeyHex[:2] != "0x" {
		publicKeyHex = "0x" + publicKeyHex
	}

	return types.VerificationMethod{
		ID:         didUrl,
		Controller: *controller,
		PublicKey: types.PublicKey{
			Type:         vtype,
			PublicKeyHex: publicKeyHex,
		},
	}, nil
}

// test creat, read, update and delete
func TestBasic(t *testing.T) {
	sks, pks, err := ToPublicKeys(globalPrivateKeys)
	if err != nil {
		t.Error(err.Error())
		return
	}

	// Used for signing transactions and paying gas fees for users.
	privateKey := sks[0]

	// Used for signing user operations, the owner of the DID.
	userSK := sks[1]
	masterKey := pks[1]

	pk1 := pks[2]
	pk2 := pks[3]

	d := &types.MemoDIDDocument{}

	controller, err := NewMemoDIDController(privateKey, "dev")
	if err != nil {
		t.Fatal(err.Error())
	}

	resolver, err := NewMemoDIDResolver("dev")
	if err != nil {
		t.Fatal(err.Error())
	}

	did, err := controller.CreateUnregisteredDID("EcdsaSecp256k1VerificationKey2019", crypto.CompressPubkey(&userSK.PublicKey))
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(did)

	// register did(userSK)
	message, err := controller.GetRegisterMessage(did, "EcdsaSecp256k1VerificationKey2019", crypto.CompressPubkey(&userSK.PublicKey))
	if err != nil {
		t.Fatal(err.Error())
	}

	hash := crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err := crypto.Sign(hash, userSK)
	if err != nil {
		t.Fatal(err.Error())
	}

	err = controller.RegisterDID(did, "EcdsaSecp256k1VerificationKey2019", crypto.CompressPubkey(&userSK.PublicKey), signature)
	if err != nil {
		t.Fatal(err.Error())
	}

	document, err := resolver.Resolve(did.String())
	if err != nil {
		t.Fatal(err.Error())
	}

	verificationMethod, err := genVerificationMethod(did, 0, nil, "EcdsaSecp256k1VerificationKey2019", masterKey)
	if err != nil {
		t.Fatal(err.Error())
	}

	d.Context = DefaultContext
	d.ID = *did
	d.VerificationMethod = append(d.VerificationMethod, verificationMethod)
	if !reflect.DeepEqual(document, d) {
		t.Fatal("Unexpect RegisterDID result")
	}

	//
	// add verification method(key-1)
	publicKey1Bytes, _ := hex.DecodeString(pk1)
	message, err = controller.GetAddVerificationMethodMessage(did, "EcdsaSecp256k1VerificationKey2019", *did, publicKey1Bytes)
	if err != nil {
		t.Fatal(err.Error())
	}

	hash = crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err = crypto.Sign(hash, userSK)
	if err != nil {
		t.Fatal(err.Error())
	}

	err = controller.AddVerificationMethod(did, "EcdsaSecp256k1VerificationKey2019", *did, publicKey1Bytes, signature)
	if err != nil {
		t.Fatal(err.Error())
	}

	document, err = resolver.Resolve(did.String())
	if err != nil {
		t.Fatal(err.Error())
	}

	verificationMethod, err = genVerificationMethod(did, 1, did, "EcdsaSecp256k1VerificationKey2019", pk1)
	if err != nil {
		t.Fatal(err.Error())
	}
	d.VerificationMethod = append(d.VerificationMethod, verificationMethod)
	if !reflect.DeepEqual(document, d) {
		t.Fatal("Unexpect RegisterDID result")
	}

	//
	// add authentication(key-1)
	message, err = controller.GetAddRelationShipMessage(did, types.Authentication, document.VerificationMethod[0].ID, 0)
	if err != nil {
		t.Fatal(err.Error())
	}
	hash = crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err = crypto.Sign(hash, userSK)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = controller.AddRelationShip(did, types.Authentication, document.VerificationMethod[0].ID, 0, signature)
	if err != nil {
		t.Fatal(err.Error())
	}

	document, err = resolver.Resolve(did.String())
	if err != nil {
		t.Error(err.Error())
		return
	}

	d.Authentication = append(d.Authentication, d.VerificationMethod[0].ID)
	if !reflect.DeepEqual(document, d) {
		t.Error("Unexpect RegisterDID result")
		return
	}

	//
	// add verification method(key-2)
	publicKey2Bytes, _ := hex.DecodeString(pk2)
	message, err = controller.GetAddVerificationMethodMessage(did, "EcdsaSecp256k1VerificationKey2019", *did, publicKey2Bytes)
	if err != nil {
		t.Error(err.Error())
		return
	}

	hash = crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err = crypto.Sign(hash, userSK)
	if err != nil {
		t.Error(err.Error())
		return
	}

	err = controller.AddVerificationMethod(did, "EcdsaSecp256k1VerificationKey2019", *did, publicKey2Bytes, signature)
	if err != nil {
		t.Error(err.Error())
		return
	}

	document, err = resolver.Resolve(did.String())
	if err != nil {
		t.Error(err.Error())
		return
	}

	verificationMethod, err = genVerificationMethod(did, 2, did, "EcdsaSecp256k1VerificationKey2019", pk2)
	if err != nil {
		t.Error(err.Error())
		return
	}
	d.VerificationMethod = append(d.VerificationMethod, verificationMethod)
	if !reflect.DeepEqual(document, d) {
		t.Error("Unexpect result")
		return
	}

	//
	// add assertion method(masterKey)
	message, err = controller.GetAddRelationShipMessage(did, types.AssertionMethod, document.VerificationMethod[0].ID, 0)
	if err != nil {
		t.Error(err.Error())
		return
	}

	hash = crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err = crypto.Sign(hash, userSK)
	if err != nil {
		t.Error(err.Error())
		return
	}

	err = controller.AddRelationShip(did, types.AssertionMethod, document.VerificationMethod[0].ID, 0, signature)
	if err != nil {
		t.Error(err.Error())
		return
	}

	document, err = resolver.Resolve(did.String())
	if err != nil {
		t.Error(err.Error())
		return
	}

	d.AssertionMethod = append(d.AssertionMethod, d.VerificationMethod[0].ID)
	if !reflect.DeepEqual(document, d) {
		t.Error("Unexpect result")
		return
	}

	//
	// add delegation(key-1, key-2)
	expireTime1 := time.Now().Unix() + 7*24*int64(time.Hour.Seconds())
	message, err = controller.GetAddRelationShipMessage(did, types.CapabilityDelegation, document.VerificationMethod[1].ID, expireTime1)
	if err != nil {
		t.Error(err.Error())
		return
	}

	hash = crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err = crypto.Sign(hash, userSK)
	if err != nil {
		t.Error(err.Error())
		return
	}

	err = controller.AddRelationShip(did, types.CapabilityDelegation, document.VerificationMethod[1].ID, expireTime1, signature)
	if err != nil {
		t.Error(err.Error())
		return
	}

	expireTime2 := time.Now().Unix() + int64(time.Minute.Seconds())
	message, err = controller.GetAddRelationShipMessage(did, types.CapabilityDelegation, document.VerificationMethod[2].ID, expireTime2)
	if err != nil {
		t.Error(err.Error())
		return
	}

	hash = crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err = crypto.Sign(hash, userSK)
	if err != nil {
		t.Error(err.Error())
		return
	}

	err = controller.AddRelationShip(did, types.CapabilityDelegation, document.VerificationMethod[2].ID, expireTime2, signature)
	if err != nil {
		t.Error(err.Error())
		return
	}

	document, err = resolver.Resolve(did.String())
	if err != nil {
		t.Error(err.Error())
		return
	}

	d.CapabilityDelegation = append(d.CapabilityDelegation, d.VerificationMethod[1].ID)
	d.CapabilityDelegation = append(d.CapabilityDelegation, d.VerificationMethod[2].ID)
	if !reflect.DeepEqual(document, d) {
		t.Error("Unexpect result")
		return
	}

	//
	// add recovery(masterKey, key-1, key-2)
	message, err = controller.GetAddRelationShipMessage(did, types.Recovery, document.VerificationMethod[0].ID, 0)
	if err != nil {
		t.Error(err.Error())
		return
	}

	hash = crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err = crypto.Sign(hash, userSK)
	if err != nil {
		t.Error(err.Error())
		return
	}

	err = controller.AddRelationShip(did, types.Recovery, document.VerificationMethod[0].ID, 0, signature)
	if err != nil {
		t.Error(err.Error())
		return
	}

	message, err = controller.GetAddRelationShipMessage(did, types.Recovery, document.VerificationMethod[1].ID, 0)
	if err != nil {
		t.Error(err.Error())
		return
	}

	hash = crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err = crypto.Sign(hash, userSK)
	if err != nil {
		t.Error(err.Error())
		return
	}

	err = controller.AddRelationShip(did, types.Recovery, document.VerificationMethod[1].ID, 0, signature)
	if err != nil {
		t.Error(err.Error())
		return
	}

	message, err = controller.GetAddRelationShipMessage(did, types.Recovery, document.VerificationMethod[2].ID, 0)
	if err != nil {
		t.Error(err.Error())
		return
	}

	hash = crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err = crypto.Sign(hash, userSK)
	if err != nil {
		t.Error(err.Error())
		return
	}

	err = controller.AddRelationShip(did, types.Recovery, document.VerificationMethod[2].ID, 0, signature)
	if err != nil {
		t.Error(err.Error())
		return
	}

	document, err = resolver.Resolve(did.String())
	if err != nil {
		t.Error(err.Error())
		return
	}

	d.Recovery = append(d.Recovery, d.VerificationMethod[0].ID)
	d.Recovery = append(d.Recovery, d.VerificationMethod[1].ID)
	d.Recovery = append(d.Recovery, d.VerificationMethod[2].ID)

	if !reflect.DeepEqual(document, d) {
		t.Error("Unexpect result")
		return
	}

	//
	// delegation(key-2) expires automatically
	time.Sleep(time.Minute)

	document, err = resolver.Resolve(did.String())
	if err != nil {
		t.Error(err.Error())
		return
	}

	for i, delegation := range d.CapabilityDelegation {
		if delegation.String() == d.VerificationMethod[2].ID.String() {
			d.CapabilityDelegation = append(d.CapabilityDelegation[:i], d.CapabilityDelegation[i+1:]...)
			break
		}
	}
	if !reflect.DeepEqual(document, d) {
		t.Error("Unexpect result")
		return
	}

	// deactivate recovery(key-1)
	message, err = controller.GetDeactivateRelationShipMessage(did, types.Recovery, document.VerificationMethod[1].ID)
	if err != nil {
		t.Error(err.Error())
		return
	}

	hash = crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err = crypto.Sign(hash, userSK)
	if err != nil {
		t.Error(err.Error())
		return
	}

	err = controller.DeactivateRelationShip(did, types.Recovery, document.VerificationMethod[1].ID, signature)
	if err != nil {
		t.Error(err.Error())
		return
	}

	document, err = resolver.Resolve(did.String())
	if err != nil {
		t.Error(err.Error())
		return
	}

	for i, recovery := range d.Recovery {
		if recovery.String() == d.VerificationMethod[1].ID.String() {
			d.Recovery = append(d.Recovery[:i], d.Recovery[i+1:]...)
			break
		}
	}
	if !reflect.DeepEqual(document, d) {
		t.Error("Unexpect result")
		return
	}

	//
	// deactivate verification method(key-2), also deactivate recovery(key-2).
	message, err = controller.GetDeactivateVerificationMethodMessage(document.VerificationMethod[2].ID)
	if err != nil {
		t.Error(err.Error())
		return
	}

	hash = crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err = crypto.Sign(hash, userSK)
	if err != nil {
		t.Error(err.Error())
		return
	}

	err = controller.DeactivateVerificationMethod(document.VerificationMethod[2].ID, signature)
	if err != nil {
		t.Error(err.Error())
		return
	}

	document, err = resolver.Resolve(did.String())
	if err != nil {
		t.Error(err.Error())
		return
	}

	d.VerificationMethod = d.VerificationMethod[:2]
	d.Recovery = d.Recovery[:1]
	if !reflect.DeepEqual(document, d) {
		t.Error("Unexpect result")
		return
	}

	data, _ := json.MarshalIndent(document, " ", "\t")
	t.Log(string(data))
	data, _ = json.MarshalIndent(d, " ", "\t")
	t.Log(string(data))

	remainUrl := d.Recovery[0]

	// deactivate did
	message, err = controller.GetDeactivateDIDMessage(did)
	if err != nil {
		t.Error(err.Error())
		return
	}

	hash = crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err = crypto.Sign(hash, userSK)
	if err != nil {
		t.Error(err.Error())
		return
	}

	err = controller.DeactivateDID(did, signature)
	if err != nil {
		t.Error(err.Error())
		return
	}

	document, err = resolver.Resolve(did.String())
	if err != nil {
		t.Error(err.Error())
		return
	}

	d = &types.MemoDIDDocument{}
	if !reflect.DeepEqual(document, d) {
		t.Error("Unexpect result")
		return
	}

	// Trying to update deactivated did
	publicKey1Bytes, _ = hex.DecodeString(pk1)
	message, err = controller.GetAddVerificationMethodMessage(did, "EcdsaSecp256k1VerificationKey2019", *did, publicKey1Bytes)
	if err == nil {
		hash = crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
		signature, _ = crypto.Sign(hash, userSK)
		err = controller.AddVerificationMethod(did, "EcdsaSecp256k1VerificationKey2019", *did, publicKey1Bytes, signature)
		if err == nil {
			t.Error("There should report an error when trying to update deactivated did")
			return
		}
	}

	message, err = controller.GetAddRelationShipMessage(did, types.Authentication, remainUrl, 0)
	if err == nil {
		hash = crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
		signature, _ = crypto.Sign(hash, userSK)
		err = controller.AddRelationShip(did, types.Authentication, remainUrl, 0, signature)
		if err == nil {
			t.Error("There should report an error when trying to update deactivated did")
			return
		}
	}

	message, err = controller.GetAddRelationShipMessage(did, types.AssertionMethod, remainUrl, 0)
	if err == nil {
		hash = crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
		signature, _ = crypto.Sign(hash, userSK)
		err = controller.AddRelationShip(did, types.AssertionMethod, remainUrl, 0, signature)
		if err == nil {
			t.Error("There should report an error when trying to update deactivated did")
			return
		}
	}

	message, err = controller.GetAddRelationShipMessage(did, types.CapabilityDelegation, remainUrl, 0)
	if err == nil {
		hash = crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
		signature, _ = crypto.Sign(hash, userSK)
		err = controller.AddRelationShip(did, types.CapabilityDelegation, remainUrl, 0, signature)
		if err == nil {
			t.Error("There should report an error when trying to update deactivated did")
			return
		}
	}

	message, err = controller.GetAddRelationShipMessage(did, types.Recovery, remainUrl, 0)
	if err == nil {
		hash = crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
		signature, _ = crypto.Sign(hash, userSK)
		err = controller.AddRelationShip(did, types.Recovery, remainUrl, 0, signature)
		if err == nil {
			t.Error("There should report an error when trying to update deactivated did")
			return
		}
	}
}

func TestResolve(t *testing.T) {
	resolver, err := NewMemoDIDResolver("dev")
	if err != nil {
		t.Error(err.Error())
		return
	}

	start := time.Now()
	document, err := resolver.Resolve("did:memo:d687daa192ffa26373395872191e8502cc41fbfbf27dc07d3da3a35de57c2d96")
	if err != nil {
		t.Error(err.Error())
		return
	}
	t.Log(time.Since(start).Seconds())

	documentBytes, err := json.MarshalIndent(document, "", "\t")
	if err != nil {
		t.Error(err.Error())
		return
	}

	fmt.Println(string(documentBytes))
}

func TestDerefrence(t *testing.T) {
	resolver, err := NewMemoDIDResolver("dev")
	if err != nil {
		t.Error(err.Error())
		return
	}

	publicKey, err := resolver.Dereference("did:memo:d687daa192ffa26373395872191e8502cc41fbfbf27dc07d3da3a35de57c2d96#masterKey")
	if err != nil {
		t.Error(err.Error())
		return
	}

	t.Log(publicKey)
}
