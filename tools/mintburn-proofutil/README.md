# Merkle Proof Tool:

1. Create an escrow tx: `maanypd tx mintburn escrow-initial maanydex 10000000000umaany --from val --gas auto --gas-adjustment 1.8 --fees 200000umaany --chain-id maany-local-1 --keyring-backend test -y`
2. Check it: `maanypd q mintburn escrows`
3. Check the escrow-proof: `maanypd q mintburn escrow-proof maanydex umaany`
4. Write down
   - `data-hex`
   - `height`
5. Run script:

   - `tools/mintburn-proofutil/create-merkleproof-genesis.sh \
      --node http://localhost:26657 \
      --store-key x-mintburn \
      --data-hex 0131 \
      --height 4 \
      --provider-chain-id maany-local-1 \
      --provider-client-id 07-tendermint-0 \
      --allowed-provider-denom umaany \
      --mint-denom umaany \
      --amount-denom umaany \
      --amount-value 10000000000 \
      --escrow-id 1 \
      --recipient maany-dex19k2p7rdqvjcm7yq57c6ntgsfna857pq79fgmpt \
     > tools/mintburn-proofutil/bundle.json`

# Gen Hash

NODE="http://localhost:26657"
APP_HASH_HEX=$(curl -s "$NODE/block?height=2" | jq -r '.result.block.header.app_hash')
echo $APP_HASH_HEX
APP_HASH_B64=$(echo "$APP_HASH_HEX" | xxd -r -p | base64 | tr -d '\n')
echo "$APP_HASH_B64"
