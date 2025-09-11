package types

import "fmt"

const (
	ModuleName = "mintburn"
	StoreKey   = "x-mintburn"
	RouterKey  = ModuleName
)

const EscrowPrefix = "escrows" // -> key: escrows/<consumerID>/<denom>

func EscrowKey(consumerID, denom string) []byte {
	return []byte(fmt.Sprintf("%s/%s/%s", EscrowPrefix, consumerID, denom))
}