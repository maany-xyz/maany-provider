package config

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ----- Coins / display -----
const (
	HumanCoinUnit    = "MAANY"  // UI/display ticker
	BaseCoinUnit     = "umaany" // base denom (on-chain)
	DefaultBondDenom = BaseCoinUnit
	DefaultExponent  = 6
)

// ----- Bech32 prefixes -----
// choose your root; I’ll assume "maany"
const (
	Bech32MainPrefix = "maany"

	PrefixValidator = "val"
	PrefixConsensus = "cons"
	PrefixPublic    = "pub"
	PrefixOperator  = "oper"

	Bech32PrefixAccAddr  = Bech32MainPrefix
	Bech32PrefixAccPub   = Bech32MainPrefix + PrefixPublic
	Bech32PrefixValAddr  = Bech32MainPrefix + PrefixValidator + PrefixOperator     // maanyvaloper
	Bech32PrefixValPub   = Bech32MainPrefix + PrefixValidator + PrefixOperator + PrefixPublic
	Bech32PrefixConsAddr = Bech32MainPrefix + PrefixValidator + PrefixConsensus    // maanyvalcons
	Bech32PrefixConsPub  = Bech32MainPrefix + PrefixValidator + PrefixConsensus + PrefixPublic
)

// GetDefaultConfig sets prefixes on the global SDK config.
// Call Seal() later (in main) after all Set* are done.
func GetDefaultConfig() *sdk.Config {
	cfg := sdk.GetConfig()
	cfg.SetBech32PrefixForAccount(Bech32PrefixAccAddr, Bech32PrefixAccPub)
	cfg.SetBech32PrefixForValidator(Bech32PrefixValAddr, Bech32PrefixValPub)
	cfg.SetBech32PrefixForConsensusNode(Bech32PrefixConsAddr, Bech32PrefixConsPub)

	// Optional: keep standard BIP-44
	cfg.SetPurpose(44)
	cfg.SetCoinType(118)

	return cfg
}
