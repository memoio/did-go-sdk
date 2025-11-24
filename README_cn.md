# Memo DID And Mfile DID Golang SDK

## 介绍

该文档主要介绍如何使用go-did，从而和Memo DID以及Mfile DID的合约进行交互。关于DID的详细信息请查阅[DID文档](https://www.w3.org/TR/did-core/)，关于Memo DID的详细信息请查阅[Memo DID文档](https://github.com/memoio/did-docs/blob/master/memo-did-design.md)，关于Mfile DID的详细信息请查阅[Mfile DID文档](https://github.com/memoio/did-docs/blob/master/mfile-did-design.md)。

## 安装

-

## 与Memo DID合约进行交互

在go-did中，提供了`MemoDIDController`类，用于控制合约中保存的Memo DID文档，从而实现对Memo DID权限的控制。目前支持如下链：

- dev：https://devchain.metamemo.one:8501
- megrez: https://chain.metamemo.one:8501

### 创建DID

创建一个全新的DID。

```go
package main

import (
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/memoio/go-did/memo"
)

func main() {
	// 用于签名交易和支付gas费用
	sk, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	// 用于签名用户操作，DID的所有者
	userSK, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	controller, err := memo.NewMemoDIDController(sk, "dev")
	if err != nil {
		panic(err.Error())
	}

	// 创建未注册的DID
	did, err := controller.CreateUnregisteredDID("EcdsaSecp256k1VerificationKey2019", crypto.CompressPubkey(&userSK.PublicKey))
	if err != nil {
		panic(err.Error())
	}

	// 获取需要签名的消息
	message, err := controller.GetRegisterMessage(did, "EcdsaSecp256k1VerificationKey2019", crypto.CompressPubkey(&userSK.PublicKey))
	if err != nil {
		panic(err.Error())
	}

	// 使用EIP-191格式签名消息
	hash := crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err := crypto.Sign(hash, userSK)
	if err != nil {
		panic(err.Error())
	}

	// 注册DID
	err = controller.RegisterDID(did, "EcdsaSecp256k1VerificationKey2019", crypto.CompressPubkey(&userSK.PublicKey), signature)
	if err != nil {
		panic(err.Error())
	}

	log.Println(did.String())
}
```

### 查看DID文档详细信息

如果DID已经创建，可以查看完整的DID文档。

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

### 添加新的验证方法

可以为一个已有的Memo DID添加新的验证方法。

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
	// 用于签名交易和支付gas费用
	sk, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	// 用于签名用户操作，DID的所有者
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
	publicKeyBytes, err := hex.DecodeString(publicKeyHex[2:]) // 移除0x前缀
	if err != nil {
		panic(err.Error())
	}

	// 获取需要签名的消息
	message, err := controller.GetAddVerificationMethodMessage(DID, "EcdsaSecp256k1VerificationKey2019", *DID, publicKeyBytes)
	if err != nil {
		panic(err.Error())
	}

	// 使用EIP-191格式签名消息
	hash := crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err := crypto.Sign(hash, userSK)
	if err != nil {
		panic(err.Error())
	}

	// 添加验证方法
	err = controller.AddVerificationMethod(DID, "EcdsaSecp256k1VerificationKey2019", *DID, publicKeyBytes, signature)
	if err != nil {
		panic(err.Error())
	}
}
```

### 修改验证方法

可以修改已有的验证方法

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
	// 用于签名交易和支付gas费用
	sk, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	// 用于签名用户操作，DID的所有者
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
	publicKeyBytes, err := hex.DecodeString(publicKeyHex[2:]) // 移除0x前缀
	if err != nil {
		panic(err.Error())
	}

	// 获取需要签名的消息
	message, err := controller.GetUpdateVerificationMethodMessage(didUrl, "EcdsaSecp256k1VerificationKey2019", publicKeyBytes)
	if err != nil {
		panic(err.Error())
	}

	// 使用EIP-191格式签名消息
	hash := crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err := crypto.Sign(hash, userSK)
	if err != nil {
		panic(err.Error())
	}

	// 更新验证方法
	err = controller.UpdateVerificationMethod(didUrl, "EcdsaSecp256k1VerificationKey2019", publicKeyBytes, signature)
	if err != nil {
		panic(err.Error())
	}
}
```

### 删除验证方法

可以删除已有的验证方法

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
	// 用于签名交易和支付gas费用
	sk, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	// 用于签名用户操作，DID的所有者
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

	// 获取需要签名的消息
	message, err := controller.GetDeactivateVerificationMethodMessage(didUrl)
	if err != nil {
		panic(err.Error())
	}

	// 使用EIP-191格式签名消息
	hash := crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err := crypto.Sign(hash, userSK)
	if err != nil {
		panic(err.Error())
	}

	// 停用验证方法
	err = controller.DeactivateVerificationMethod(didUrl, signature)
	if err != nil {
		panic(err.Error())
	}
}
```

### 添加登录验证方法

在创建Memo DID后，可以添加新的登录验证方法，验证方法包括公钥信息等。成功添加后，可以使用对应私钥的签名，以该DID的身份线下登录第三方应用，例如memo中间件。

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
	// 用于签名交易和支付gas费用
	sk, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	// 用于签名用户操作，DID的所有者
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

	// 获取需要签名的消息
	message, err := controller.GetAddRelationShipMessage(DID, types.Authentication, didUrl, 0)
	if err != nil {
		panic(err.Error())
	}

	// 使用EIP-191格式签名消息
	hash := crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err := crypto.Sign(hash, userSK)
	if err != nil {
		panic(err.Error())
	}

	// 添加关系
	err = controller.AddRelationShip(DID, types.Authentication, didUrl, 0, signature)
	if err != nil {
		panic(err.Error())
	}
}
```

### 删除登录验证方法

可以删除已有的登录方法。删除成功后，使用对应私钥的签名，将不能以该DID的身份线下登录第三方应用。

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
	// 用于签名交易和支付gas费用
	sk, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	// 用于签名用户操作，DID的所有者
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

	// 获取需要签名的消息
	message, err := controller.GetDeactivateRelationShipMessage(DID, types.Authentication, didUrl)
	if err != nil {
		panic(err.Error())
	}

	// 使用EIP-191格式签名消息
	hash := crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err := crypto.Sign(hash, userSK)
	if err != nil {
		panic(err.Error())
	}

	// 停用关系
	err = controller.DeactivateRelationShip(DID, types.Authentication, didUrl, signature)
	if err != nil {
		panic(err.Error())
	}
}
```

### 添加代理访问验证方法

在创建Memo DID后，可以添加新的代理访问验证方法，验证方法包括公钥信息等。成功添加后，可以使用对应私钥的签名，以该DID的身份访问需要权限的资源，例如，Memo中间件中用户的私有文件。

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
	// 用于签名交易和支付gas费用
	sk, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	// 用于签名用户操作，DID的所有者
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

	// 获取需要签名的消息
	message, err := controller.GetAddRelationShipMessage(DID, types.CapabilityDelegation, didUrl, 0)
	if err != nil {
		panic(err.Error())
	}

	// 使用EIP-191格式签名消息
	hash := crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err := crypto.Sign(hash, userSK)
	if err != nil {
		panic(err.Error())
	}

	// 添加关系
	err = controller.AddRelationShip(DID, types.CapabilityDelegation, didUrl, 0, signature)
	if err != nil {
		panic(err.Error())
	}
}
```

### 删除代理访问验证方法

可以删除原有的代理访问验证方法删除成功后，使用对应私钥的签名，将不能以该DID的身份访问需要权限的资源。

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
	// 用于签名交易和支付gas费用
	sk, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	// 用于签名用户操作，DID的所有者
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

	// 获取需要签名的消息
	message, err := controller.GetDeactivateRelationShipMessage(DID, types.CapabilityDelegation, didUrl)
	if err != nil {
		panic(err.Error())
	}

	// 使用EIP-191格式签名消息
	hash := crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err := crypto.Sign(hash, userSK)
	if err != nil {
		panic(err.Error())
	}

	// 停用关系
	err = controller.DeactivateRelationShip(DID, types.CapabilityDelegation, didUrl, signature)
	if err != nil {
		panic(err.Error())
	}
}
```

### 购买读权限

能够通过付费的方式购买私有文件的读权限。在购买读权限后，会将memo did添加到mfile did的read字段中，从而能够线下请求mfile did对应的文件。在购买读权限之前，需要调用approve方法。

**注意：** `ApproveOfMfileContract` 和 `BuyReadPermission` 方法目前在 memo controller 中已被注释。请查看实现以获取最新的 API。

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

	// 注意：这些方法目前已被注释
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

### 删除DID

可以删除已创建的DID。删除后，DID将不可用且该DID将无法重新创建。

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
	// 用于签名交易和支付gas费用
	sk, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	// 用于签名用户操作，DID的所有者
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

	// 获取需要签名的消息
	message, err := controller.GetDeactivateDIDMessage(DID)
	if err != nil {
		panic(err.Error())
	}

	// 使用EIP-191格式签名消息
	hash := crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err := crypto.Sign(hash, userSK)
	if err != nil {
		panic(err.Error())
	}

	// 停用DID
	err = controller.DeactivateDID(DID, signature)
	if err != nil {
		panic(err.Error())
	}
}
```

## 与Mfile DID合约进行交互

在go-did中，提供了`MfileDIDController`类，用于控制合约中保存的Mfile DID文档，从而实现对Mfile DID权限的控制。目前支持如下链：

- dev：https://devchain.metamemo.one:8501

### 创建DID

创建一个全新的Mfile DID。

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
	// 用于签名交易和支付gas费用
	sk, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	// 用于签名用户操作，DID的所有者
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

	// 获取需要签名的消息
	message, err := controller.GetRegisterDIDMessage(mfileDID, "mid", 0, big.NewInt(50), []string{"memo", "example"}, *memodid)
	if err != nil {
		panic(err.Error())
	}

	// 使用EIP-191格式签名消息
	hash := crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err := crypto.Sign(hash, userSK)
	if err != nil {
		panic(err.Error())
	}

	// 注册DID
	err = controller.RegisterDID(mfileDID, "mid", 0, big.NewInt(50), []string{"memo", "example"}, *memodid, signature)
	if err != nil {
		panic(err.Error())
	}
}
```

### 更改所有者

Mfile DID的所有者可以通过更改所有者的方式，将Mfile DID实现转让的功能。

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
	// 用于签名交易和支付gas费用
	sk, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	// 用于签名用户操作，DID的所有者
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

	// 获取需要签名的消息
	message, err := controller.GetChangeControllerMessage(mfileDID, *memodid)
	if err != nil {
		panic(err.Error())
	}

	// 使用EIP-191格式签名消息
	hash := crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err := crypto.Sign(hash, userSK)
	if err != nil {
		panic(err.Error())
	}

	// 更改控制器
	err = controller.ChangeController(mfileDID, *memodid, signature)
	if err != nil {
		panic(err.Error())
	}
}
```

### 更改文件类型

Mfile DID对应的文件包括公开文件以及私有文件，可以通过该方法修改文件的类型。其中，0表示private文件，1表示public文件

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
	// 用于签名交易和支付gas费用
	sk, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	// 用于签名用户操作，DID的所有者
	userSK, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	controller, err := mfile.NewMfileDIDController(sk, "dev", did)
	if err != nil {
		panic(err.Error())
	}

	mfileDID, _ := types.ParseMfileDID(did)

	// 获取需要签名的消息
	message, err := controller.GetChangeFileTypeMessage(mfileDID, 1)
	if err != nil {
		panic(err.Error())
	}

	// 使用EIP-191格式签名消息
	hash := crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err := crypto.Sign(hash, userSK)
	if err != nil {
		panic(err.Error())
	}

	// 更改文件类型
	err = controller.ChangeFileType(mfileDID, 1, signature)
	if err != nil {
		panic(err.Error())
	}
}
```

### 更改文件的价格

可以修改文件的价格。

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
	// 用于签名交易和支付gas费用
	sk, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	// 用于签名用户操作，DID的所有者
	userSK, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	controller, err := mfile.NewMfileDIDController(sk, "dev", did)
	if err != nil {
		panic(err.Error())
	}

	mfileDID, _ := types.ParseMfileDID(did)

	// 获取需要签名的消息
	message, err := controller.GetChangePriceMessage(mfileDID, big.NewInt(25))
	if err != nil {
		panic(err.Error())
	}

	// 使用EIP-191格式签名消息
	hash := crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err := crypto.Sign(hash, userSK)
	if err != nil {
		panic(err.Error())
	}

	// 更改价格
	err = controller.ChangePrice(mfileDID, big.NewInt(25), signature)
	if err != nil {
		panic(err.Error())
	}
}
```

### 更改文件关键词

文件的关键词用于搜索文件，可以根据需要更改文件的关键词

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
	// 用于签名交易和支付gas费用
	sk, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	// 用于签名用户操作，DID的所有者
	userSK, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	controller, err := mfile.NewMfileDIDController(sk, "dev", did)
	if err != nil {
		panic(err.Error())
	}

	mfileDID, _ := types.ParseMfileDID(did)

	// 获取需要签名的消息
	message, err := controller.GetChangeKeywordsMessage(mfileDID, []string{"movie", "china"})
	if err != nil {
		panic(err.Error())
	}

	// 使用EIP-191格式签名消息
	hash := crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err := crypto.Sign(hash, userSK)
	if err != nil {
		panic(err.Error())
	}

	// 更改关键词
	err = controller.ChangeKeywords(mfileDID, []string{"movie", "china"}, signature)
	if err != nil {
		panic(err.Error())
	}
}
```

### 授予读取权限

当Mfile DID显示的文件为私有文件时，其他Memo DID的所有者除了购买读权限外，还可以直接授予读取权限。

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
	// 用于签名交易和支付gas费用
	sk, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	// 用于签名用户操作，DID的所有者
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

	// 获取需要签名的消息
	message, err := controller.GetAddRelationShipMessage(mfileDID, types.Read, *memodid)
	if err != nil {
		panic(err.Error())
	}

	// 使用EIP-191格式签名消息
	hash := crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err := crypto.Sign(hash, userSK)
	if err != nil {
		panic(err.Error())
	}

	// 添加关系
	err = controller.AddRelationShip(mfileDID, types.Read, *memodid, signature)
	if err != nil {
		panic(err.Error())
	}
}
```

### 撤销读取权限

可以撤销之前授予的读取权限，但是无法撤销其他人购买的读取权限。

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
	// 用于签名交易和支付gas费用
	sk, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	// 用于签名用户操作，DID的所有者
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

	// 获取需要签名的消息
	message, err := controller.GetDeactivateRelationShipMessage(mfileDID, types.Read, *memodid)
	if err != nil {
		panic(err.Error())
	}

	// 使用EIP-191格式签名消息
	hash := crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err := crypto.Sign(hash, userSK)
	if err != nil {
		panic(err.Error())
	}

	// 停用关系
	err = controller.DeactivateRelationShip(mfileDID, types.Read, *memodid, signature)
	if err != nil {
		panic(err.Error())
	}
}
```

### 删除DID

可以删除Mfile DID

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
	// 用于签名交易和支付gas费用
	sk, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	// 用于签名用户操作，DID的所有者
	userSK, err := crypto.GenerateKey()
	if err != nil {
		panic(err.Error())
	}

	controller, err := mfile.NewMfileDIDController(sk, "dev", did)
	if err != nil {
		panic(err.Error())
	}

	mfileDID, _ := types.ParseMfileDID(did)

	// 获取需要签名的消息
	message, err := controller.GetDeactivateDIDMessage(mfileDID)
	if err != nil {
		panic(err.Error())
	}

	// 使用EIP-191格式签名消息
	hash := crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err := crypto.Sign(hash, userSK)
	if err != nil {
		panic(err.Error())
	}

	// 停用DID
	err = controller.DeactivateDID(mfileDID, signature)
	if err != nil {
		panic(err.Error())
	}
}
```
