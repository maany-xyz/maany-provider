#!/usr/bin/env bash
set -euo pipefail

# One-shot generator (camelCase JSON):
# - Fetch ABCI proof for a KV (store/key@height)
# - Transform to ICS-23 merkleProof via your Go tool
# - Auto-derive hexKey + value (base64) from the merkle proof
# - Fetch block app_hash at --height and base64-encode it for genesisTrustedRoot.hash
# - Emit the final genesismint bundle JSON to stdout (camelCase)
#
# Requirements: curl, jq, base64, xxd, go (to run your proofutil)
#
# Usage:
#   tools/mintburn-bundle-one-shot.sh \
#     --node http://localhost:26657 \
#     --store-key x-mintburn \
#     --data-hex 0131 \
#     --height 102 \
#     --provider-chain-id maany-provider \
#     --provider-client-id 07-tendermint-0 \
#     --allowed-provider-denom umaany \
#     --mint-denom umaany \
#     --amount-denom umaany \
#     --amount-value 10000000000 \
#     --escrow-id 1 \
#     --recipient maany1...consumerRecipient \
#     > bundle.json

node=""
store_key=""
data_hex=""         # e.g. 0131 (without 0x)
height=""
provider_chain_id=""
provider_client_id=""
allowed_provider_denom=""
mint_denom=""
amount_denom=""
amount_value=""
escrow_id=""
recipient=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --node)                     node="$2"; shift 2;;
    --store-key)                store_key="$2"; shift 2;;
    --data-hex)                 data_hex="$2"; shift 2;;
    --height)                   height="$2"; shift 2;;
    --provider-chain-id)        provider_chain_id="$2"; shift 2;;
    --provider-client-id)       provider_client_id="$2"; shift 2;;
    --allowed-provider-denom)   allowed_provider_denom="$2"; shift 2;;
    --mint-denom)               mint_denom="$2"; shift 2;;
    --amount-denom)             amount_denom="$2"; shift 2;;
    --amount-value)             amount_value="$2"; shift 2;;
    --escrow-id)                escrow_id="$2"; shift 2;;
    --recipient)                recipient="$2"; shift 2;;
    *) echo "unknown arg: $1" >&2; exit 1;;
  esac
done

for v in node store_key data_hex height provider_chain_id provider_client_id allowed_provider_denom mint_denom amount_denom amount_value escrow_id recipient; do
  if [[ -z "${!v}" ]]; then
    echo "missing --$v" >&2
    exit 1
  fi
done

# Dependencies check
for bin in curl jq base64 xxd go; do
  command -v "$bin" >/dev/null 2>&1 || { echo "missing dependency: $bin" >&2; exit 1; }
done

# 1) ABCI query -> proof_raw (kept in-memory)
abci_url="${node}/abci_query?path=\"/store/${store_key}/key\"&data=0x${data_hex}&prove=true&height=${height}"
proof_raw="$(curl -sS "$abci_url")"

# 2) Run your Go transformer -> merkle_proof_json (capture via here-string)
merkle_proof_json="$(go run ./tools/mintburn-proofutil <<< "$proof_raw")"

# 3) Auto-derive hexKey + value (base64) from merkle_proof_json
#    Pick the NON-store proof: its key (base64) != store_key
value_b64="$(jq -r --arg sk "$store_key" '
  .proofs[]
  | select((.exist.key | @base64d) != $sk)
  | .exist.value
' <<< "$merkle_proof_json")"

hex_key="$(jq -r --arg sk "$store_key" '
  .proofs[]
  | select((.exist.key | @base64d) != $sk)
  | .exist.key
' <<< "$merkle_proof_json" | base64 -d | xxd -p -c256)"

# Safety check: hex_key should match data_hex (case-insensitive)
if [[ "$(echo "$hex_key"   | tr '[:upper:]' '[:lower:]')" != \
      "$(echo "$data_hex"  | tr '[:upper:]' '[:lower:]')" ]]; then
  echo "warning: derived hexKey (${hex_key}) != --data-hex (${data_hex})" >&2
fi

# 4) Fetch block header app_hash at --height and base64-encode it
block_url="${node}/block?height=$((height+1))"
app_hash_hex="$(curl -sS "$block_url" | jq -r '.result.block.header.app_hash // empty')"
if [[ -z "$app_hash_hex" || "$app_hash_hex" == "null" ]]; then
  echo "error: could not fetch app_hash from ${block_url}" >&2
  exit 1
fi
# Convert hex -> raw bytes -> base64 (no newline)
genesis_trusted_root_hash="$(printf "%s" "$app_hash_hex" | xxd -r -p | base64 | tr -d '\n')"

# Fixed revisionNumber as string "0" and revisionHeight = height (string)
revision_number_str="0"
revision_height_str="$height"

# 5) Emit final bundle JSON (camelCase) to stdout
jq -n \
  --arg merkle_proof_json "$merkle_proof_json" \
  --arg height "$height" \
  --arg store_key "$store_key" \
  --arg hex_key "$hex_key" \
  --arg value_b64 "$value_b64" \
  --arg provider_chain_id "$provider_chain_id" \
  --arg provider_client_id "$provider_client_id" \
  --arg allowed_provider_denom "$allowed_provider_denom" \
  --arg mint_denom "$mint_denom" \
  --arg amount_denom "$amount_denom" \
  --arg amount_value "$amount_value" \
  --arg escrow_id "$escrow_id" \
  --arg recipient "$recipient" \
  --arg rev_num "$revision_number_str" \
  --arg rev_height "$revision_height_str" \
  --arg root_hash "$genesis_trusted_root_hash" '
{
  params: {
    providerClientId:     $provider_client_id,
    providerChainId:      $provider_chain_id,
    allowedProviderDenom: $allowed_provider_denom,
    mintDenom:            $mint_denom,
    genesisTrustedRoot: {
      revisionNumber: $rev_num,
      revisionHeight: $rev_height,
      hash:           $root_hash
    },
    useGenesisTrustedRoot: true
  },
  mints: [
    {
      merkleProof:                 ($merkle_proof_json | fromjson),
      proofHeightRevisionNumber:   0,
      proofHeightRevisionHeight:   ($height | tonumber),
      keyPath:                     [ $store_key, $hex_key ],
      value:                       $value_b64,
      providerChainId:             $provider_chain_id,
      amountDenom:                 $amount_denom,
      amountValue:                 $amount_value,
      escrowId:                    $escrow_id,
      recipient:                   $recipient
    }
  ],
  claimedEscrowIds: []
}'
