// tools/mintburn-proofutil/main.go
package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	commitmenttypes "github.com/cosmos/ibc-go/v8/modules/core/23-commitment/types"

	// CometBFT proto types for proofs
	tmcryptopb "github.com/cometbft/cometbft/proto/tendermint/crypto"

	// gogo proto/json for ibc-go types

	"github.com/gogo/protobuf/jsonpb"
)

type rpcResp struct {
	Result struct {
		Response struct {
			ProofOps struct {
				Ops []struct {
					Type string `json:"type"`
					Key  string `json:"key"`  // base64
					Data string `json:"data"` // base64
				} `json:"ops"`
			} `json:"proofOps"`
		} `json:"response"`
	} `json:"result"`
}

func main() {
	var r rpcResp
	if err := json.NewDecoder(os.Stdin).Decode(&r); err != nil {
		panic(err)
	}

	// Build []tmcryptopb.ProofOp (VALUE slice, not []*ProofOp)
	ops := make([]tmcryptopb.ProofOp, 0, len(r.Result.Response.ProofOps.Ops))
	for _, o := range r.Result.Response.ProofOps.Ops {
		key, err := base64.StdEncoding.DecodeString(o.Key)
		if err != nil { panic(err) }
		data, err := base64.StdEncoding.DecodeString(o.Data)
		if err != nil { panic(err) }
		ops = append(ops, tmcryptopb.ProofOp{
			Type: o.Type,
			Key:  key,
			Data: data,
		})
	}
	tm := &tmcryptopb.ProofOps{Ops: ops}

	// Convert to IBC MerkleProof
	mp, err := commitmenttypes.ConvertProofs(tm)
	if err != nil { panic(err) }

	// ---- Option A: print gogo JSON (nice for embedding in genesis) ----
	m := jsonpb.Marshaler{}
	if err := m.Marshal(os.Stdout, &mp); err != nil { panic(err) }
	fmt.Println()

	// // ---- Option B: print base64 of wire bytes (if your consumer expects bytes) ----
	// raw, err := gogoproto.Marshal(&mp)
	// if err != nil { panic(err) }
	// fmt.Println(base64.StdEncoding.EncodeToString(raw))
}
