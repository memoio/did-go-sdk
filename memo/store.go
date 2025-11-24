package memo

import "github.com/memoio/go-did/types"

type DIDController interface {
	// Create
	GetRegisterMessage(did *types.MemoDID, methodType string, publicKeyBytes []byte) (string, error)
	RegisterDID(did *types.MemoDID, methodType string, publicKeyBytes []byte, signature []byte) error

	// Update
	GetAddVerificationMethodMessage(did *types.MemoDID, vtype string, controller types.MemoDID, publicKeyBytes []byte) (string, error)
	AddVerificationMethod(did *types.MemoDID, vtype string, controller types.MemoDID, publicKeyBytes []byte, signature []byte) error
	GetUpdateVerificationMethodMessage(didUrl types.MemoDIDUrl, vtype string, publicKeyBytes []byte) (string, error)
	UpdateVerificationMethod(didUrl types.MemoDIDUrl, vtype string, publicKeyBytes []byte, signature []byte) error
	GetDeactivateVerificationMethodMessage(didUrl types.MemoDIDUrl) (string, error)
	DeactivateVerificationMethod(didUrl types.MemoDIDUrl, signature []byte) error
	// Relation ship include: authentication; assertionMethod; capabilityDelegation; recovery
	GetAddRelationShipMessage(did *types.MemoDID, relationType int, didUrl types.MemoDIDUrl, expireTime int64) (string, error)
	AddRelationShip(did *types.MemoDID, relationType int, didUrl types.MemoDIDUrl, expireTime int64, signature []byte) error
	GetDeactivateRelationShipMessage(did *types.MemoDID, relationType int, didUrl types.MemoDIDUrl) (string, error)
	DeactivateRelationShip(did *types.MemoDID, relationType int, didUrl types.MemoDIDUrl, signature []byte) error

	// // Update mfile-did
	// BuyReadPermission(mfileDID types.MfileDID, memoDID *types.MemoDID) error

	// Delete
	GetDeactivateDIDMessage(did *types.MemoDID) (string, error)
	DeactivateDID(did *types.MemoDID, signature []byte) error
}

type DIDResolver interface {
	// Read
	Resolve(didString string) (*types.MemoDIDDocument, error)
	Dereference(didUrlString string) ([]types.PublicKey, error)
}
