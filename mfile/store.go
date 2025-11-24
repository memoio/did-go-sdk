package mfile

import (
	"math/big"

	"github.com/memoio/go-did/types"
)

type MfileStore interface {
	GetRegisterDIDMessage(did *types.MfileDID, encode string, ftype uint8, price *big.Int, keywords []string, controller types.MemoDID) (string, error)
	RegisterDID(did *types.MfileDID, encode string, ftype uint8, price *big.Int, keywords []string, controller types.MemoDID, signature []byte) error

	GetChangeControllerMessage(did *types.MfileDID, controller types.MemoDID) (string, error)
	ChangeController(did *types.MfileDID, controller types.MemoDID, signature []byte) error
	GetChangeFileTypeMessage(did *types.MfileDID, ftype uint8) (string, error)
	ChangeFileType(did *types.MfileDID, ftype uint8, signature []byte) error
	GetChangePriceMessage(did *types.MfileDID, price *big.Int) (string, error)
	ChangePrice(did *types.MfileDID, price *big.Int, signature []byte) error
	GetChangeKeywordsMessage(did *types.MfileDID, keywords []string) (string, error)
	ChangeKeywords(did *types.MfileDID, keywords []string, signature []byte) error
	// Relation ship include: read
	GetAddRelationShipMessage(did *types.MfileDID, relationType int, memoDID types.MemoDID) (string, error)
	AddRelationShip(did *types.MfileDID, relationType int, memoDID types.MemoDID, signature []byte) error
	GetDeactivateRelationShipMessage(did *types.MfileDID, relationType int, memoDID types.MemoDID) (string, error)
	DeactivateRelationShip(did *types.MfileDID, relationType int, memoDID types.MemoDID, signature []byte) error

	GetDeactivateDIDMessage(did *types.MfileDID) (string, error)
	DeactivateDID(did *types.MfileDID, signature []byte) error
}

type MfileResolver interface {
	Resolve(didString string) error
	CanRead(didString string) error
}
