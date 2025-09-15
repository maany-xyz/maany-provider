package types

import "strconv"

const (
	ModuleName = "mintburn"
	StoreKey   = "x-mintburn"
	RouterKey  = ModuleName
)

// const EscrowPrefix = "escrows" // -> key: escrows/<consumerID>/<denom>

// func EscrowKey(consumerID, denom string) []byte {
// 	return []byte(fmt.Sprintf("%s/%s/%s", EscrowPrefix, consumerID, denom))
// }
var EscrowIDCounterKey = []byte{0x00}
func Uint64ToString(n uint64) string { return strconv.FormatUint(n, 10) }

// Primary store (by escrow_id)
// Primary storage prefix (by escrow_id)
var EscrowPrefix = []byte{0x01}

// Secondary index prefix: (consumer_chain_id, denom) -> escrow_id
var EscrowIndexPrefix = []byte{0x02}

func EscrowKeyByID(escrowID string) []byte {
	return append(EscrowPrefix, []byte(escrowID)...)
}

func EscrowIndexKey(consumerChainID, denom string) []byte {
	// format: EscrowIndexPrefix || consumerChainID || 0x00 || denom
	k := make([]byte, 0, len(EscrowIndexPrefix)+len(consumerChainID)+1+len(denom))
	k = append(k, EscrowIndexPrefix...)
	k = append(k, []byte(consumerChainID)...)
	k = append(k, 0x00)
	k = append(k, []byte(denom)...)
	return k
}