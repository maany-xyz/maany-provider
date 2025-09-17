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

// Authorized ICA address per consumer chain: consumer_chain_id -> bech32 addr bytes
var AuthorizedICAPrefix = []byte{0x03}

// NormalizeChainIDBase returns the base chain-id by stripping a single trailing
// "-<digits>" revision suffix if present. Examples:
//   "maanydex-1" -> "maanydex"
//   "chain-test" -> "chain-test" (unchanged)
//   "foo-abc" -> "foo-abc" (unchanged)
func NormalizeChainIDBase(chainID string) string {
    if len(chainID) == 0 {
        return chainID
    }
    // Find last '-'
    last := -1
    for i := len(chainID) - 1; i >= 0; i-- {
        if chainID[i] == '-' {
            last = i
            break
        }
    }
    if last <= 0 || last == len(chainID)-1 {
        return chainID
    }
    // Ensure suffix is all digits
    for j := last + 1; j < len(chainID); j++ {
        c := chainID[j]
        if c < '0' || c > '9' {
            return chainID
        }
    }
    return chainID[:last]
}

// MintburnEscrowChannelIDFor derives a synthetic "channel id" string used to
// compute a dedicated ICS-20 escrow address per consumer (normalized by base
// chain-id). It does not need to correspond to a real IBC channel.
func MintburnEscrowChannelIDFor(consumerChainID string) string {
    base := NormalizeChainIDBase(consumerChainID)
    return "mintburn-" + base
}

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

func AuthorizedICAKey(consumerChainID string) []byte {
    k := make([]byte, 0, len(AuthorizedICAPrefix)+len(consumerChainID))
    k = append(k, AuthorizedICAPrefix...)
    k = append(k, []byte(consumerChainID)...)
    return k
}
