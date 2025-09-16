# Merkle Proof Tool:

1. Create an escrow tx: `maanypd tx mintburn escrow-initial maanydex 10000000000umaany --from val --gas auto --gas-adjustment 1.8 --fees 200000umaany --chain-id maany-local-1 --keyring-backend test -y`
2. Check it: `maanypd q tx <hash from tx above>`

- Take the `height`

3. (Optional) Check the escrow-proof: `maanypd q mintburn escrow-proof maanydex umaany`

- `maanypd q mintburn escrow maanydex umaany`

- Write down:
- `data-hex`
- Note:
  - this doesn't change and for testing just use the value in the command below, but double check for testnet/mainnet and provide right `--data-hex`.
  - down below we use the hermes account address `maany-dex1c7j33khjnqf2s44t2aykshthjkpf50sfxzu7af`, which is the address of `maanydex` chain-id. Make sure you provide the correct relayer account address for testnet/mainnet in the field ` --recipient`. The steps how to do that is described in `root/.sh/README.md`

5. Run script:

   - `tools/mintburn-proofutil/create-merkleproof-genesis.sh \
      --node http://localhost:26657 \
      --store-key x-mintburn \
      --data-hex 0131 \
      --height 5 \
      --provider-chain-id maany-local-1 \
      --provider-client-id 07-tendermint-0 \
      --allowed-provider-denom umaany \
      --mint-denom umaany \
      --amount-denom umaany \
      --amount-value 10000000000 \
      --escrow-id 1 \
      --recipient maany-dex1c7j33khjnqf2s44t2aykshthjkpf50sfxzu7af \
     > tools/mintburn-proofutil/bundle.json`

6. This generates the ./bundle.json -> copy paste it into maanydex config genesis.json under `genesismint`
