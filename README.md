# Memo DID And Mfile DID Golang SDK

## Introduction

This document mainly introduces how to use go-did to interact with Memo DID and Mfile DID contracts. For more information about DID, please refer to the [DID document](https://www.w3.org/TR/did-core/), for more information about Memo DID, please refer to the [Memo DID document](https://github.com/memoio/did-docs/blob/master/memo-did-design.md), for more information about Mfile DID, please refer to the [Mfile DID document](https://github.com/memoio/did-docs/blob/master/mfile-did-design.md).

## Install

-

## Interact with Memo DID Contracts

In go-did, the `MemoDIDController` class is provided to control the Memo DID document saved in the contract, thereby realizing the control of Memo DID permissions. Currently the following chains are supported:

- dev：https://devchain.metamemo.one:8501
- megrez: https://chain.metamemo.one:8501

### Create DID

Create a new DID.

```go
package main

import (
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/memoio/go-did/memo"
)

func main() {
	// Used for signing transactions and paying gas fees
	sk, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	// Used for signing user operations, the owner of the DID
	userSK, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	controller, err := memo.NewMemoDIDController(sk, "dev")
	if err != nil {
		panic(err.Error())
	}

	// Create unregistered DID
	did, err := controller.CreateUnregisteredDID("EcdsaSecp256k1VerificationKey2019", crypto.CompressPubkey(&userSK.PublicKey))
	if err != nil {
		panic(err.Error())
	}

	// Get message to be signed
	message, err := controller.GetRegisterMessage(did, "EcdsaSecp256k1VerificationKey2019", crypto.CompressPubkey(&userSK.PublicKey))
	if err != nil {
		panic(err.Error())
	}

	// Sign the message using EIP-191
	hash := crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err := crypto.Sign(hash, userSK)
	if err != nil {
		panic(err.Error())
	}

	// Register DID
	err = controller.RegisterDID(did, "EcdsaSecp256k1VerificationKey2019", crypto.CompressPubkey(&userSK.PublicKey), signature)
	if err != nil {
		panic(err.Error())
	}

	log.Println(did.String())
}
```

### View DID document details

If the DID has been created, you can view the complete DID document.

```go
package main

import (
	"encoding/json"
	"log"

	"github.com/memoio/go-did/memo"
)

func main() {
	did := "did:memo:d687daa192ffa26373395872191e8502cc41fbfbf27dc07d3da3a35de57c2d96"

	resolver, err := memo.NewMemoDIDResolver("dev")
	if err != nil {
		panic(err.Error())
	}

	document, err := resolver.Resolve(did)
	if err != nil {
		panic(err.Error())
	}

	data, err := json.Marshal(document)
	if err != nil {
		panic(err.Error())
	}

	log.Println(string(data))
}
```

### Add new VerificationMethod

It is possible to add new verification methods to an existing Memo DID.

```go
package main

import (
	"encoding/hex"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/memoio/go-did/memo"
	"github.com/memoio/go-did/types"
)

func main() {
	did := "did:memo:d687daa192ffa26373395872191e8502cc41fbfbf27dc07d3da3a35de57c2d96"
	// Used for signing transactions and paying gas fees
	sk, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	// Used for signing user operations, the owner of the DID
	userSK, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	controller, err := memo.NewMemoDIDController(sk, "dev")
	if err != nil {
		panic(err.Error())
	}

	DID, err := types.ParseMemoDID(did)
	if err != nil {
		panic(err.Error())
	}

	publicKeyHex := "0x02d78b20654eb7a5d58d83b25d090a338eff18f0b5f919777c9d894c2e161b4b52"
	publicKeyBytes, err := hex.DecodeString(publicKeyHex[2:]) // Remove 0x prefix
	if err != nil {
		panic(err.Error())
	}

	// Get message to be signed
	message, err := controller.GetAddVerificationMethodMessage(DID, "EcdsaSecp256k1VerificationKey2019", *DID, publicKeyBytes)
	if err != nil {
		panic(err.Error())
	}

	// Sign the message using EIP-191
	hash := crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err := crypto.Sign(hash, userSK)
	if err != nil {
		panic(err.Error())
	}

	// Add verification method
	err = controller.AddVerificationMethod(DID, "EcdsaSecp256k1VerificationKey2019", *DID, publicKeyBytes, signature)
	if err != nil {
		panic(err.Error())
	}
}
```

### Update VerificationMethod

Existing authentication methods can be modified.

```go
package main

import (
	"encoding/hex"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/memoio/go-did/memo"
	"github.com/memoio/go-did/types"
)

func main() {
	did := "did:memo:d687daa192ffa26373395872191e8502cc41fbfbf27dc07d3da3a35de57c2d96"
	// Used for signing transactions and paying gas fees
	sk, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	// Used for signing user operations, the owner of the DID
	userSK, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	controller, err := memo.NewMemoDIDController(sk, "dev")
	if err != nil {
		panic(err.Error())
	}

	DID, err := types.ParseMemoDID(did)
	if err != nil {
		panic(err.Error())
	}
	didUrl, _ := DID.DIDUrl(1)

	publicKeyHex := "0x03d21e6c4843fa3f5d019e551131106e2075925b01da2a83dc177879a512eb608f"
	publicKeyBytes, err := hex.DecodeString(publicKeyHex[2:]) // Remove 0x prefix
	if err != nil {
		panic(err.Error())
	}

	// Get message to be signed
	message, err := controller.GetUpdateVerificationMethodMessage(didUrl, "EcdsaSecp256k1VerificationKey2019", publicKeyBytes)
	if err != nil {
		panic(err.Error())
	}

	// Sign the message using EIP-191
	hash := crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err := crypto.Sign(hash, userSK)
	if err != nil {
		panic(err.Error())
	}

	// Update verification method
	err = controller.UpdateVerificationMethod(didUrl, "EcdsaSecp256k1VerificationKey2019", publicKeyBytes, signature)
	if err != nil {
		panic(err.Error())
	}
}
```

### Delete VerificationMethod

You can delete existing authentication methods.

```go
package main

import (
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/memoio/go-did/memo"
	"github.com/memoio/go-did/types"
)

func main() {
	did := "did:memo:d687daa192ffa26373395872191e8502cc41fbfbf27dc07d3da3a35de57c2d96"
	// Used for signing transactions and paying gas fees
	sk, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	// Used for signing user operations, the owner of the DID
	userSK, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	controller, err := memo.NewMemoDIDController(sk, "dev")
	if err != nil {
		panic(err.Error())
	}

	DID, err := types.ParseMemoDID(did)
	if err != nil {
		panic(err.Error())
	}
	didUrl, _ := DID.DIDUrl(1)

	// Get message to be signed
	message, err := controller.GetDeactivateVerificationMethodMessage(didUrl)
	if err != nil {
		panic(err.Error())
	}

	// Sign the message using EIP-191
	hash := crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err := crypto.Sign(hash, userSK)
	if err != nil {
		panic(err.Error())
	}

	// Deactivate verification method
	err = controller.DeactivateVerificationMethod(didUrl, signature)
	if err != nil {
		panic(err.Error())
	}
}
```

### Add login verification method

After creating a Memo DID, you can add a new login verification method, which includes public key information, etc. After successfully adding, you can use the signature of the corresponding private key to log in to a third-party application offline as the DID, such as the memo middleware.

```go
package main

import (
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/memoio/go-did/memo"
	"github.com/memoio/go-did/types"
)

func main() {
	did := "did:memo:d687daa192ffa26373395872191e8502cc41fbfbf27dc07d3da3a35de57c2d96"
	// Used for signing transactions and paying gas fees
	sk, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	// Used for signing user operations, the owner of the DID
	userSK, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	controller, err := memo.NewMemoDIDController(sk, "dev")
	if err != nil {
		panic(err.Error())
	}

	DID, err := types.ParseMemoDID(did)
	if err != nil {
		panic(err.Error())
	}
	didUrl, _ := DID.DIDUrl(0)

	// Get message to be signed
	message, err := controller.GetAddRelationShipMessage(DID, types.Authentication, didUrl, 0)
	if err != nil {
		panic(err.Error())
	}

	// Sign the message using EIP-191
	hash := crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err := crypto.Sign(hash, userSK)
	if err != nil {
		panic(err.Error())
	}

	// Add relation ship
	err = controller.AddRelationShip(DID, types.Authentication, didUrl, 0, signature)
	if err != nil {
		panic(err.Error())
	}
}
```

### Deleting a login verification method

You can delete an existing login method. After the deletion is successful, you will not be able to log in to third-party applications offline with the DID using the signature of the corresponding private key.

```go
package main

import (
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/memoio/go-did/memo"
	"github.com/memoio/go-did/types"
)

func main() {
	did := "did:memo:d687daa192ffa26373395872191e8502cc41fbfbf27dc07d3da3a35de57c2d96"
	// Used for signing transactions and paying gas fees
	sk, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	// Used for signing user operations, the owner of the DID
	userSK, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	controller, err := memo.NewMemoDIDController(sk, "dev")
	if err != nil {
		panic(err.Error())
	}

	DID, err := types.ParseMemoDID(did)
	if err != nil {
		panic(err.Error())
	}
	didUrl, _ := DID.DIDUrl(0)

	// Get message to be signed
	message, err := controller.GetDeactivateRelationShipMessage(DID, types.Authentication, didUrl)
	if err != nil {
		panic(err.Error())
	}

	// Sign the message using EIP-191
	hash := crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err := crypto.Sign(hash, userSK)
	if err != nil {
		panic(err.Error())
	}

	// Deactivate relation ship
	err = controller.DeactivateRelationShip(DID, types.Authentication, didUrl, signature)
	if err != nil {
		panic(err.Error())
	}
}
```

### Add proxy access authentication method

After creating a Memo DID, you can add a new proxy access verification method, which includes public key information, etc. After successfully adding, you can use the signature of the corresponding private key to access resources that require permissions as the DID, such as the user's private files in the Memo middleware.

```go
package main

import (
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/memoio/go-did/memo"
	"github.com/memoio/go-did/types"
)

func main() {
	did := "did:memo:d687daa192ffa26373395872191e8502cc41fbfbf27dc07d3da3a35de57c2d96"
	// Used for signing transactions and paying gas fees
	sk, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	// Used for signing user operations, the owner of the DID
	userSK, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	controller, err := memo.NewMemoDIDController(sk, "dev")
	if err != nil {
		panic(err.Error())
	}

	DID, err := types.ParseMemoDID(did)
	if err != nil {
		panic(err.Error())
	}
	didUrl, _ := DID.DIDUrl(0)

	// Get message to be signed
	message, err := controller.GetAddRelationShipMessage(DID, types.CapabilityDelegation, didUrl, 0)
	if err != nil {
		panic(err.Error())
	}

	// Sign the message using EIP-191
	hash := crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err := crypto.Sign(hash, userSK)
	if err != nil {
		panic(err.Error())
	}

	// Add relation ship
	err = controller.AddRelationShip(DID, types.CapabilityDelegation, didUrl, 0, signature)
	if err != nil {
		panic(err.Error())
	}
}
```

### Deleting a proxy access authentication method

The original proxy access verification method can be deleted. After the deletion is successful, the corresponding private key signature will be used to sign the DID and the resources that require permission will not be accessible.

```go
package main

import (
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/memoio/go-did/memo"
	"github.com/memoio/go-did/types"
)

func main() {
	did := "did:memo:d687daa192ffa26373395872191e8502cc41fbfbf27dc07d3da3a35de57c2d96"
	// Used for signing transactions and paying gas fees
	sk, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	// Used for signing user operations, the owner of the DID
	userSK, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	controller, err := memo.NewMemoDIDController(sk, "dev")
	if err != nil {
		panic(err.Error())
	}

	DID, err := types.ParseMemoDID(did)
	if err != nil {
		panic(err.Error())
	}
	didUrl, _ := DID.DIDUrl(0)

	// Get message to be signed
	message, err := controller.GetDeactivateRelationShipMessage(DID, types.CapabilityDelegation, didUrl)
	if err != nil {
		panic(err.Error())
	}

	// Sign the message using EIP-191
	hash := crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err := crypto.Sign(hash, userSK)
	if err != nil {
		panic(err.Error())
	}

	// Deactivate relation ship
	err = controller.DeactivateRelationShip(DID, types.CapabilityDelegation, didUrl, signature)
	if err != nil {
		panic(err.Error())
	}
}
```

### Purchase read permissions

You can purchase the read permission of private files by paying. After purchasing the read permission, memo did will be added to the read field of mfile did, so that you can request the file corresponding to mfile did offline. Before purchasing the read permission, you need to call the approve method.

**Note:** `ApproveOfMfileContract` and `BuyReadPermission` methods are currently commented out in the memo controller. Please check the implementation for the latest API.

```go
package main

import (
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/memoio/go-did/memo"
	"github.com/memoio/go-did/types"
)

func main() {
	did := "did:memo:d687daa192ffa26373395872191e8502cc41fbfbf27dc07d3da3a35de57c2d96"
	sk, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	controller, err := memo.NewMemoDIDController(sk, "dev")
	if err != nil {
		panic(err.Error())
	}

	mfiledid, _ := types.ParseMfileDID("did:mfile:bafkreic7emp2v6ofwkpiiqmrbjq2m6sgyws4eyq5jbphqiywkqyxzbags4")
	memoDID, _ := types.ParseMemoDID(did)

	// Note: These methods are currently commented out
	// err = controller.ApproveOfMfileContract(1000)
	// if err != nil {
	// 	panic(err.Error())
	// }

	// err = controller.BuyReadPermission(*mfiledid, memoDID)
	// if err != nil {
	// 	panic(err.Error())
	// }
}
```

### Deleting a DID

Delete the created DID. Once deleted, the DID will be unavailable and cannot be recreated.

```go
package main

import (
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/memoio/go-did/memo"
	"github.com/memoio/go-did/types"
)

func main() {
	did := "did:memo:d687daa192ffa26373395872191e8502cc41fbfbf27dc07d3da3a35de57c2d96"
	// Used for signing transactions and paying gas fees
	sk, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	// Used for signing user operations, the owner of the DID
	userSK, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	controller, err := memo.NewMemoDIDController(sk, "dev")
	if err != nil {
		panic(err.Error())
	}

	DID, err := types.ParseMemoDID(did)
	if err != nil {
		panic(err.Error())
	}

	// Get message to be signed
	message, err := controller.GetDeactivateDIDMessage(DID)
	if err != nil {
		panic(err.Error())
	}

	// Sign the message using EIP-191
	hash := crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err := crypto.Sign(hash, userSK)
	if err != nil {
		panic(err.Error())
	}

	// Deactivate DID
	err = controller.DeactivateDID(DID, signature)
	if err != nil {
		panic(err.Error())
	}
}
```

## Interacting with the Mfile DID contract

In go-did, the `MfileDIDController` class is provided to control the Mfile DID document saved in the contract, thereby realizing the control of Mfile DID permissions. Currently the following chains are supported:

- dev：https://devchain.metamemo.one:8501
- megrez: https://chain.metamemo.one:8501

### Create DID

Create a new Mfile DID.

```go
package main

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/memoio/go-did/mfile"
	"github.com/memoio/go-did/types"
)

func main() {
	did := "did:mfile:bafkreic7emp2v6ofwkpiiqmrbjq2m6sgyws4eyq5jbphqiywkqyxzbags4"
	// Used for signing transactions and paying gas fees
	sk, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	// Used for signing user operations, the owner of the DID
	userSK, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	controller, err := mfile.NewMfileDIDController(sk, "dev", did)
	if err != nil {
		panic(err.Error())
	}

	mfileDID, _ := types.ParseMfileDID(did)
	memodid, _ := types.ParseMemoDID("did:memo:d687daa192ffa26373395872191e8502cc41fbfbf27dc07d3da3a35de57c2d96")

	// Get message to be signed
	message, err := controller.GetRegisterDIDMessage(mfileDID, "mid", 0, big.NewInt(50), []string{"memo", "example"}, *memodid)
	if err != nil {
		panic(err.Error())
	}

	// Sign the message using EIP-191
	hash := crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err := crypto.Sign(hash, userSK)
	if err != nil {
		panic(err.Error())
	}

	// Register DID
	err = controller.RegisterDID(mfileDID, "mid", 0, big.NewInt(50), []string{"memo", "example"}, *memodid, signature)
	if err != nil {
		panic(err.Error())
	}
}
```

### Change Owner

The owner of Mfile DID can transfer the Mfile DID by changing the owner.

```go
package main

import (
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/memoio/go-did/mfile"
	"github.com/memoio/go-did/types"
)

func main() {
	did := "did:mfile:bafkreic7emp2v6ofwkpiiqmrbjq2m6sgyws4eyq5jbphqiywkqyxzbags4"
	// Used for signing transactions and paying gas fees
	sk, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	// Used for signing user operations, the owner of the DID
	userSK, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	controller, err := mfile.NewMfileDIDController(sk, "dev", did)
	if err != nil {
		panic(err.Error())
	}

	mfileDID, _ := types.ParseMfileDID(did)
	memodid, _ := types.ParseMemoDID("did:memo:d687daa192ffa26373395872191e8502cc41fbfbf27dc07d3da3a35de57c2d96")

	// Get message to be signed
	message, err := controller.GetChangeControllerMessage(mfileDID, *memodid)
	if err != nil {
		panic(err.Error())
	}

	// Sign the message using EIP-191
	hash := crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err := crypto.Sign(hash, userSK)
	if err != nil {
		panic(err.Error())
	}

	// Change controller
	err = controller.ChangeController(mfileDID, *memodid, signature)
	if err != nil {
		panic(err.Error())
	}
}
```

### Change the file type

The files corresponding to Mfile DID include public files and private files. You can modify the file type through this method. 0 represents private files and 1 represents public files.

```go
package main

import (
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/memoio/go-did/mfile"
	"github.com/memoio/go-did/types"
)

func main() {
	did := "did:mfile:bafkreic7emp2v6ofwkpiiqmrbjq2m6sgyws4eyq5jbphqiywkqyxzbags4"
	// Used for signing transactions and paying gas fees
	sk, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	// Used for signing user operations, the owner of the DID
	userSK, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	controller, err := mfile.NewMfileDIDController(sk, "dev", did)
	if err != nil {
		panic(err.Error())
	}

	mfileDID, _ := types.ParseMfileDID(did)

	// Get message to be signed
	message, err := controller.GetChangeFileTypeMessage(mfileDID, 1)
	if err != nil {
		panic(err.Error())
	}

	// Sign the message using EIP-191
	hash := crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err := crypto.Sign(hash, userSK)
	if err != nil {
		panic(err.Error())
	}

	// Change file type
	err = controller.ChangeFileType(mfileDID, 1, signature)
	if err != nil {
		panic(err.Error())
	}
}
```

### Change the price of the file

```go
package main

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/memoio/go-did/mfile"
	"github.com/memoio/go-did/types"
)

func main() {
	did := "did:mfile:bafkreic7emp2v6ofwkpiiqmrbjq2m6sgyws4eyq5jbphqiywkqyxzbags4"
	// Used for signing transactions and paying gas fees
	sk, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	// Used for signing user operations, the owner of the DID
	userSK, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	controller, err := mfile.NewMfileDIDController(sk, "dev", did)
	if err != nil {
		panic(err.Error())
	}

	mfileDID, _ := types.ParseMfileDID(did)

	// Get message to be signed
	message, err := controller.GetChangePriceMessage(mfileDID, big.NewInt(25))
	if err != nil {
		panic(err.Error())
	}

	// Sign the message using EIP-191
	hash := crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err := crypto.Sign(hash, userSK)
	if err != nil {
		panic(err.Error())
	}

	// Change price
	err = controller.ChangePrice(mfileDID, big.NewInt(25), signature)
	if err != nil {
		panic(err.Error())
	}
}
```

### Change file keywords

The keywords of the file are used to search for the file. You can change the keywords of the file as needed.

```go
package main

import (
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/memoio/go-did/mfile"
	"github.com/memoio/go-did/types"
)

func main() {
	did := "did:mfile:bafkreic7emp2v6ofwkpiiqmrbjq2m6sgyws4eyq5jbphqiywkqyxzbags4"
	// Used for signing transactions and paying gas fees
	sk, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	// Used for signing user operations, the owner of the DID
	userSK, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	controller, err := mfile.NewMfileDIDController(sk, "dev", did)
	if err != nil {
		panic(err.Error())
	}

	mfileDID, _ := types.ParseMfileDID(did)

	// Get message to be signed
	message, err := controller.GetChangeKeywordsMessage(mfileDID, []string{"movie", "china"})
	if err != nil {
		panic(err.Error())
	}

	// Sign the message using EIP-191
	hash := crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err := crypto.Sign(hash, userSK)
	if err != nil {
		panic(err.Error())
	}

	// Change keywords
	err = controller.ChangeKeywords(mfileDID, []string{"movie", "china"}, signature)
	if err != nil {
		panic(err.Error())
	}
}
```

### Grant Read Permission

When the file displayed by the Mfile DID is a private file, other Memo DID owners can directly grant read permissions in addition to purchasing read permissions.

```go
package main

import (
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/memoio/go-did/mfile"
	"github.com/memoio/go-did/types"
)

func main() {
	did := "did:mfile:bafkreic7emp2v6ofwkpiiqmrbjq2m6sgyws4eyq5jbphqiywkqyxzbags4"
	// Used for signing transactions and paying gas fees
	sk, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	// Used for signing user operations, the owner of the DID
	userSK, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	controller, err := mfile.NewMfileDIDController(sk, "dev", did)
	if err != nil {
		panic(err.Error())
	}

	mfileDID, _ := types.ParseMfileDID(did)
	memodid, _ := types.ParseMemoDID("did:memo:d687daa192ffa26373395872191e8502cc41fbfbf27dc07d3da3a35de57c2d96")

	// Get message to be signed
	message, err := controller.GetAddRelationShipMessage(mfileDID, types.Read, *memodid)
	if err != nil {
		panic(err.Error())
	}

	// Sign the message using EIP-191
	hash := crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err := crypto.Sign(hash, userSK)
	if err != nil {
		panic(err.Error())
	}

	// Add relation ship
	err = controller.AddRelationShip(mfileDID, types.Read, *memodid, signature)
	if err != nil {
		panic(err.Error())
	}
}
```

### Revoke Read Permission

You can revoke previously granted read permissions, but you cannot revoke read permissions purchased by others.

```go
package main

import (
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/memoio/go-did/mfile"
	"github.com/memoio/go-did/types"
)

func main() {
	did := "did:mfile:bafkreic7emp2v6ofwkpiiqmrbjq2m6sgyws4eyq5jbphqiywkqyxzbags4"
	// Used for signing transactions and paying gas fees
	sk, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	// Used for signing user operations, the owner of the DID
	userSK, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	controller, err := mfile.NewMfileDIDController(sk, "dev", did)
	if err != nil {
		panic(err.Error())
	}

	mfileDID, _ := types.ParseMfileDID(did)
	memodid, _ := types.ParseMemoDID("did:memo:d687daa192ffa26373395872191e8502cc41fbfbf27dc07d3da3a35de57c2d96")

	// Get message to be signed
	message, err := controller.GetDeactivateRelationShipMessage(mfileDID, types.Read, *memodid)
	if err != nil {
		panic(err.Error())
	}

	// Sign the message using EIP-191
	hash := crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err := crypto.Sign(hash, userSK)
	if err != nil {
		panic(err.Error())
	}

	// Deactivate relation ship
	err = controller.DeactivateRelationShip(mfileDID, types.Read, *memodid, signature)
	if err != nil {
		panic(err.Error())
	}
}
```

### Delete DID

Delete Mfile DID.

```go
package main

import (
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/memoio/go-did/mfile"
	"github.com/memoio/go-did/types"
)

func main() {
	did := "did:mfile:bafkreic7emp2v6ofwkpiiqmrbjq2m6sgyws4eyq5jbphqiywkqyxzbags4"
	// Used for signing transactions and paying gas fees
	sk, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	// Used for signing user operations, the owner of the DID
	userSK, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	controller, err := mfile.NewMfileDIDController(sk, "dev", did)
	if err != nil {
		panic(err.Error())
	}

	mfileDID, _ := types.ParseMfileDID(did)

	// Get message to be signed
	message, err := controller.GetDeactivateDIDMessage(mfileDID)
	if err != nil {
		panic(err.Error())
	}

	// Sign the message using EIP-191
	hash := crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err := crypto.Sign(hash, userSK)
	if err != nil {
		panic(err.Error())
	}

	// Deactivate DID
	err = controller.DeactivateDID(mfileDID, signature)
	if err != nil {
		panic(err.Error())
	}
}
```
