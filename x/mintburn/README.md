# Mintburn Module: Ops Guide

This document explains how to configure and operate the mintburn module with CCV provider gating and a params‑based allow‑list for consumer chains.

## Overview

- Provider registry gate: Only IBC transfer channels whose local client ID is registered by the CCV Provider are authorized for mintburn.
- Params allow‑list: As a secondary gate (and for bootstrap), a list of full consumer chain‑ids can be authorized via params/governance.
- Single escrow per consumer: Mintburn uses a dedicated bank address derived from `GetEscrowAddress("transfer", "mintburn-" + base(<consumerChainID>))` for all mintburn flows from a consumer, regardless of transfer lane.
- Direction: Mintburn intercepts consumer → provider transfers (OnRecv on the provider) and releases from the dedicated escrow; provider → consumer transfers pass through as normal ICS‑20 sends (the consumer mints vouchers).

## Configure via genesis.json

Add mintburn‑enabled consumer chain‑ids under `app_state.mintburn.params.allowed_consumers`. Mintburn will intercept only for consumers that are both CCV‑registered and listed here. All other consumers use normal ICS‑20 voucher logic.

Example:

```
{
  "app_state": {
    "mintburn": {
      "params": {
        "allowed_consumers": ["maanydex-1"],
        "reject_unauthorized_channels": true,
        "allowed_client_ids": [],
        "allowed_connection_ids": []
      }
    }
  }
}
```

Notes:

- These are full chain‑ids (including revision suffix when present), used for authorization. You can add/remove entries later via governance (see below).
- The CCV Provider registry remains the primary authorization source; params are a secondary gate.

## Update via Governance (ParamChange)

Submit a ParamChange proposal to update `AllowedConsumers`.

Proposal JSON body (example):

```
{
  "title": "Authorize mintburn for maanydex-1",
  "description": "Allow consumer chain maanydex-1 in mintburn params",
  "changes": [
    {
      "subspace": "mintburn",
      "key": "AllowedConsumers",
      "value": "[\"maanydex-1\",\"other-consumer-1\"]"
    }
  ]
}
```

- `subspace`: must be `mintburn`.
- `key`: `AllowedConsumers` (exact case).
- `value`: JSON‑encoded array as a string.

CLI example (Cosmos SDK gov v1)

```
# 1) Write proposal body
cat > proposal.json <<'EOF'
{
  "title": "Authorize mintburn for maanydex-1",
  "description": "Allow consumer chain maanydex-1 in mintburn params",
  "changes": [
    {
      "subspace": "mintburn",
      "key": "AllowedConsumers",
      "value": "[\"maanydex-1\"]"
    }
  ]
}
EOF

# 2) Submit proposal
gaiad tx gov submit-proposal param-change proposal.json \
  --from <key> --deposit 10000000uatom --gas auto --gas-adjustment 1.5 --fees 5000uatom \
  --chain-id <chain-id> -y

# 3) Vote (example: yes)
gaiad tx gov vote <proposal-id> yes --from <key> \
  --gas auto --gas-adjustment 1.5 --fees 5000uatom --chain-id <chain-id> -y
```

Notes:

- Replace `<key>`, `<proposal-id>`, fee denom, and `<chain-id>` with your environment values.
- Some networks use different base denoms and deposits/fees; adjust accordingly.

## Logs and Diagnostics

Mintburn emits structured logs to help trace flows:

- Handshake gating: logs `local_client_id`, counterparty `chain_id`, and whether provider registry/params authorized the channel.
- Recv path (consumer → provider): logs packet ports/channels, sender/receiver, denom/amount, base‑denom resolution, and release operations.
- Ack path (provider → consumer): logs outbound transfer context (passthrough).

If releases fail, check:

- The channel used is allow‑listed and belongs to an authorized consumer.
- The denom resolves to the provider base denom (e.g., `umaany`).
- The dedicated escrow address has sufficient balance.

## Security Considerations

- Authorization: The provider registry is the primary source of truth. Params allow‑list is a secondary gate; remove entries you no longer trust.
- Chain‑id spoofing: Do not authorize by normalized base names. Authorization uses the full chain‑id and/or the CCV registry.
- Multiple channels: Mintburn supports multiple transfer lanes per consumer; all consumer → provider releases debit the same dedicated escrow address for that consumer.

## Notes / Future Options

- Base‑denom parameterization: The provider base denom (currently enforced in code) can be exposed as a module param if needed.
- Authorization remains on full chain‑ids (and/or CCV registry), while escrow accounting is normalized by base chain‑id.

## Query Current Params

You can query the current params (including `AllowedConsumers`) via the params module.

Examples:

```
# Query a single key in a subspace
gaiad q params subspace mintburn AllowedConsumers -o json

# Dump the whole mintburn subspace (if supported)
gaiad q params subspace mintburn -o json
```

Note: CLI flags and output may vary slightly by SDK version; use `--help` to see available options.

# TASKS

Security/Authorization

- Handshake hard‑fail mode: add a param to reject unauthorized transfer channel
  handshakes (return error) instead of silently ignoring in mintburn. Useful to prevent
  stray channels. -> IMPLEMENTED
- Provider registry check depth: optionally verify the transfer connection matches the
  CCV connection for that consumer (same local client/connection), not just “clientID
  belongs to consumer”.

Configuration/Params

- Base denom param: replace hardcoded “umaany” with a module param (and optional
  multiple base denoms if needed).
- AllowedConsumers ergonomics: keep full chain‑id for auth (safe), but add CLI helper
  to append the next revision quickly (quality‑of‑life).

Accounting/Normalization

- Normalization risks: now that escrow addresses normalize to base chain‑id, two
  unrelated chains sharing the same base string would share escrow. Consider namespacing
  the escrow id with a short hash of the provider local client/connection (e.g.,
  “mintburn--”) while keeping logs human‑friendly.
- Migration helper: add a one‑shot admin/CLI to sweep old dedicated escrow balances to
  the normalized address if you ever change the derivation (or moved from non‑normalized
  in older state).

APIs/Queries/CLI

- Query params: optional module-native query (e.g., q mintburn params) instead of
  using the generic params query.
- Query escrow state: add endpoints to return:
  - Dedicated escrow address per consumer (derived),
  - Current balance of that escrow address,
  - Allowed channels recorded for the consumer.
- CLI helpers: - “mintburn params set-allowed-consumers …” wrapper around param‑change draft
  generation. - “mintburn show-escrow ” to print derived escrow address + balance.

Logging/Events

- Event enrichment: include normalized base id, local client id, and whether provider
  registry vs params authorized the channel. Add a specific event on successful release
  (OnRecv) with amount, receiver, and escrow address.
- Log volume: add a param to toggle debug logs if you want to quiet the path once
  stable.

Testing

- Unit tests: already added for chain‑id normalization; add: - ReceiverChainIsSource vs destination identifiers check (guard against
  regressions), - MarkEscrowClaimed path: module→escrow movement and idempotency.
- Integration tests (when environment allows): - Channel open/close gating via provider registry vs params. - Consumer→provider sends across two channels; verify both release from the single
  escrow. - Genesis escrow bootstrap: CLAIMED funding then consumer fee flow.

Operations/Docs

- Document “do not change” guidance for escrow id derivation post‑launch (or describe
  a safe migration procedure).
- Provide a short runbook for onboarding a new consumer:
  - CCV setup,
  - Open transfer channel (provider registry authorizes),
  - Param update (if needed),
  - CLAIMED via ICA,
  - Verify dedicated escrow funded and releases working.

If you want, I can take the first two concrete steps next:

- Add a base‑denom param + CLI/query hook.
- Tighten CLAIMED to require ICA (no fallback), with a transitional param/flag if you
  want a grace period.

# Upgrading Logic for consumer chain - Mintburn

With registry-only gating and our normalized escrow IDs, a consumer revision
upgrade is straightforward. The main work is re-establishing CCV and transfer paths;
mintburn’s accounting stays stable.

What stays stable

- Escrow address: derives from base(chain-id), so maanydex-1 → maanydex-2 keeps the
  same mintburn escrow account. No fund migration.
- Release logic: still uses the same dedicated escrow; multiple transfer lanes to the
  new revision are fine once authorized.
- Authorization posture: mintburn continues to allow-list only channels whose local
  client is registered by the CCV Provider.

Upgrade runbook

- CCV first: complete the CCV consumer upgrade so the ProviderKeeper registers the new
  local clientID for the upgraded consumer (chain-id maanydex-2).
- Transfer channel: open a new ICS-20 transfer channel over the connection tied to
  that new client. With hard-fail enabled, handshakes are rejected until the provider
  registry shows the new client mapping; once present, the handshake is accepted and the
  channel is allow‑listed.
- ICA re-establish: the consumer’s ICA host channel will re‑handshake; our ICA
  middleware will record the new authorized ICA address for the new full chain-id.
- Resume flows: consumer→provider fees and transfers will be intercepted and
  released from the same dedicated escrow address as before (thanks to base(chain-id)
  normalization).

Operational notes

- No params change needed: with registry‑only gating you don’t need to update an
  allow‑list for new revisions.
- Pending escrows corner case: if you ever had a PENDING escrow keyed to the old full
  chain-id (uncommon in practice), MarkEscrowClaimed requires the ICA for that exact
  chain-id. Either clear those before upgrade or add a small compatibility rule (e.g.,
  accept ICA of the new revision when base IDs match). Most setups avoid leaving PENDING
  escrows across upgrades.
- Timing with hard‑fail: handshakes will fail until CCV has registered the new client.
  If you need to bring up the transfer channel before CCV finishes, you can temporarily
  disable “hard‑fail” (if you add that param) and re‑enable it once CCV is ready.

Risk profile

- Spoofing-resistant: a rogue chain can’t pass registry checks because it can’t
  produce valid proofs for the registered client. Removing chain-id params further
  minimizes config risk.
- Accounting clarity: a single escrow per consumer family (base ID) simplifies
  mirrored supply tracking across revisions.

# Testing

- Test:
  - attempt to open a transfer channel before CCV registry is ready:
  - With reject_unauthorized_channels = true, handshake should error. - After CCV registry shows the consumer’s client mapping, handshake should
    succeed.

`maanypd tx ibc-transfer transfer transfer channel-2 maany-dex19k2p7rdqvjcm7yq57c6ntgsfna857pq79fgmpt 50000000umaany --from val --gas auto --gas-adjustment 1.6 --fees 190000umaany --keyring-backend test --chain-id maany-local-1 -y`
