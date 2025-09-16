#!/usr/bin/env bash
set -euo pipefail

# Resolve paths relative to this script's directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROOF_SCRIPT="${SCRIPT_DIR}/create-merkleproof-genesis.sh"
BUNDLE_OUT="${SCRIPT_DIR}/bundle.json"

# Consumer chain config
CONSUMER_CONFIG="$HOME/.neutrond"
GENESIS_PATH="${CONSUMER_CONFIG}/config/genesis.json"

# Requirements
command -v maanypd >/dev/null 2>&1 || { echo "maanypd not found in PATH"; exit 1; }
command -v jq >/dev/null 2>&1 || { echo "jq not found in PATH"; exit 1; }
[[ -x "$PROOF_SCRIPT" ]] || { echo "Proof script not found or not executable at: $PROOF_SCRIPT"; exit 1; }
[[ -f "$GENESIS_PATH" ]] || { echo "genesis.json not found at $GENESIS_PATH"; exit 1; }

# --- Config (same as before) ---
CHAIN_ID="maany-local-1"
PROVIDER_CLIENT_ID="07-tendermint-0"
NODE="http://localhost:26657"
FROM="val"
KEYRING="test"
FEES="200000umaany"
GAS="auto"
GAS_ADJ="1.8"

# Escrow params
CONSUMER_CHAIN_ID="maanydex"
AMOUNT_VALUE="10000000000"
DENOM="umaany"

# Proof/bundle params
STORE_KEY="x-mintburn"
DATA_HEX="0131"
ALLOWED_PROVIDER_DENOM="umaany"
MINT_DENOM="umaany"
AMOUNT_DENOM="umaany"
ESCROW_ID="1"
RECIPIENT="maany-dex1c7j33khjnqf2s44t2aykshthjkpf50sfxzu7af"

# NEW: ICA params (adjust as needed)
ICA_CONTROLLER_CONNECTION_ID="connection-0"
ICA_OWNER="mintburn-claims"
ICA_TX_TIMEOUT_SECONDS="300"
ICA_MAX_CLAIM_PER_BLOCK="10"

echo "Submitting escrow-initial tx..."
TX_JSON=$(
  maanypd tx mintburn escrow-initial \
    "$CONSUMER_CHAIN_ID" \
    "${AMOUNT_VALUE}${DENOM}" \
    --from "$FROM" \
    --gas "$GAS" \
    --gas-adjustment "$GAS_ADJ" \
    --fees "$FEES" \
    --chain-id "$CHAIN_ID" \
    --keyring-backend "$KEYRING" \
    -y -o json
)

TX_HASH=$(echo "$TX_JSON" | jq -r '.txhash')
if [[ -z "${TX_HASH}" || "${TX_HASH}" == "null" ]]; then
  echo "Failed to obtain tx hash from submission response:"
  echo "$TX_JSON"
  exit 1
fi
echo "Tx hash: $TX_HASH"

# Poll until indexed and get height
echo "Waiting for transaction to be indexed..."
ATTEMPTS=0
MAX_ATTEMPTS=60
while :; do
  if TXQ_JSON=$(maanypd q tx "$TX_HASH" -o json 2>/dev/null); then
    HEIGHT=$(echo "$TXQ_JSON" | jq -r '.height // empty')
    if [[ -n "$HEIGHT" && "$HEIGHT" != "0" ]]; then
      echo "Tx found at height: $HEIGHT"
      break
    fi
  fi
  ((ATTEMPTS++))
  if (( ATTEMPTS >= MAX_ATTEMPTS )); then
    echo "Timed out waiting for tx $TX_HASH to be indexed."
    exit 1
  fi
  sleep 1
done
sleep 5
# Generate bundle.json in the same folder as this script
echo "Generating bundle.json at height=$HEIGHT ..."
"$PROOF_SCRIPT" \
  --node "$NODE" \
  --store-key "$STORE_KEY" \
  --data-hex "$DATA_HEX" \
  --height "$HEIGHT" \
  --provider-chain-id "$CHAIN_ID" \
  --provider-client-id "$PROVIDER_CLIENT_ID" \
  --allowed-provider-denom "$ALLOWED_PROVIDER_DENOM" \
  --mint-denom "$MINT_DENOM" \
  --amount-denom "$AMOUNT_DENOM" \
  --amount-value "$AMOUNT_VALUE" \
  --escrow-id "$ESCROW_ID" \
  --recipient "$RECIPIENT" \
  --ica-controller-connection-id "$ICA_CONTROLLER_CONNECTION_ID" \
  --ica-owner "$ICA_OWNER" \
  --ica-tx-timeout-seconds "$ICA_TX_TIMEOUT_SECONDS" \
  --ica-max-claim-per-block "$ICA_MAX_CLAIM_PER_BLOCK" \
  > "$BUNDLE_OUT"

echo "Bundle written to: $BUNDLE_OUT"

# Insert bundle.json into consumer genesis.json under app_state.genesismint
echo "Updating consumer genesis.json ..."
TMP_GENESIS="$(mktemp)"
jq --slurpfile bundle "$BUNDLE_OUT" \
   '.app_state.genesismint = $bundle[0]' \
   "$GENESIS_PATH" > "$TMP_GENESIS"

mv "$TMP_GENESIS" "$GENESIS_PATH"
echo "Updated $GENESIS_PATH with .app_state.genesismint"
