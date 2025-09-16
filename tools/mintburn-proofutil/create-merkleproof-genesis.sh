#!/usr/bin/env bash
set -euo pipefail

# One-shot generator (camelCase JSON):
# - Fetch ABCI proof for a KV (store/key@height)
# - Transform to ICS-23 merkleProof via your Go tool
# - Auto-derive hexKey + value (base64) from the merkle proof
# - Fetch block app_hash at --height+1 and base64-encode it for genesisTrustedRoot.hash
# - Emit the final genesismint bundle JSON to stdout (camelCase)
#
# Requirements: curl, jq, base64, xxd, go (to run your proofutil)

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

# NEW: ICA params
ica_controller_connection_id=""
ica_owner=""
ica_tx_timeout_seconds=""
ica_max_claim_per_block=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --node)                           node="$2"; shift 2;;
    --store-key)                      store_key="$2"; shift 2;;
    --data-hex)                       data_hex="$2"; shift 2;;
    --height)                         height="$2"; shift 2;;
    --provider-chain-id)              provider_chain_id="$2"; shift 2;;
    --provider-client-id)             provider_client_id="$2"; shift 2;;
    --allowed-provider-denom)         allowed_provider_denom="$2"; shift 2;;
    --mint-denom)                     mint_denom="$2"; shift 2;;
    --amount-denom)                   amount_denom="$2"; shift 2;;
    --amount-value)                   amount_value="$2"; shift 2;;
    --escrow-id)                      escrow_id="$2"; shift 2;;
    --recipient)                      recipient="$2"; shift 2;;
    # NEW flags:
    --ica-controller-connection-id)   ica_controller_connection_id="$2"; shift 2;;
    --ica-owner)                      ica_owner="$2"; shift 2;;
    --ica-tx-timeout-seconds)         ica_tx_timeout_seconds="$2"; shift 2;;
    --ica-max-claim-per-block)        ica_max_claim_per_block="$2"; shift 2;;
    *) echo "unknown arg: $1" >&2; exit 1;;
  esac
done

# Required args
for v in \
  node store_key data_hex height provider_chain_id provider_client_id \
  allowed_provider_denom mint_denom amount_denom amount_value escrow_id recipient \
  ica_controller_connection_id ica_owner ica_tx_timeout_seconds ica_max_claim_per_block
do
  if [[ -z "${!v}" ]]; then
    echo "missing --${v//_/-}" >&2
    exit 1
  fi
done

# Dependencies check
for bin in curl jq base64 xxd go; do
  command -v "$bin" >/dev/null 2>&1 || { echo "missing dependency: $bin" >&2; exit 1; }
done

# 1) ABCI query -> proof_raw
abci_url="${node}/abci_query?path=\"/store/${store_key}/key\"&data=0x${data_hex}&prove=true&height=${height}"
proof_raw="$(curl -sS "$abci_url")"

# 2) Transform to ICS-23 merkle proof via Go tool
merkle_proof_json="$(go run ./tools/mintburn-proofutil <<< "$proof_raw")"

# 3) Derive value_b64 and hex_key from the NON-store proof
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

# Safety check
if [[ "$(echo "$hex_key" | tr '[:upper:]' '[:lower:]')" != "$(echo "$data_hex" | tr '[:upper:]' '[:lower:]')" ]]; then
  echo "warning: derived hexKey (${hex_key}) != --data-hex (${data_hex})" >&2
fi

# 4) Block app_hash at (height+1) -> base64
block_url="${node}/block?height=$((height+1))"
app_hash_hex="$(curl -sS "$block_url" | jq -r '.result.block.header.app_hash // empty')"
if [[ -z "$app_hash_hex" || "$app_hash_hex" == "null" ]]; then
  echo "error: could not fetch app_hash from ${block_url}" >&2
  exit 1
fi
genesis_trusted_root_hash="$(printf "%s" "$app_hash_hex" | xxd -r -p | base64 | tr -d '\n')"

revision_number_str="0"
revision_height_str="$height"

# 5) Emit final bundle JSON
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
  --arg root_hash "$genesis_trusted_root_hash" \
  --arg ica_conn_id "$ica_controller_connection_id" \
  --arg ica_owner "$ica_owner" \
  --arg ica_timeout "$ica_tx_timeout_seconds" \
  --arg ica_max_claim "$ica_max_claim_per_block" '
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
    useGenesisTrustedRoot: true,
    icaControllerConnectionId: $ica_conn_id,
    icaOwner:                  $ica_owner,
    icaTxTimeoutSeconds:       $ica_timeout,
    icaMaxClaimPerBlock:       $ica_max_claim
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
