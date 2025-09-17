# Mintburn Module: Ops Guide

This document explains how to configure and operate the mintburn module with CCV provider gating and a params‑based allow‑list for consumer chains.

## Overview
- Provider registry gate: Only IBC transfer channels whose local client ID is registered by the CCV Provider are authorized for mintburn.
- Params allow‑list: As a secondary gate (and for bootstrap), a list of full consumer chain‑ids can be authorized via params/governance.
- Single escrow per consumer: Mintburn uses a dedicated bank address derived from `GetEscrowAddress("transfer", "mintburn-" + base(<consumerChainID>))` for all mintburn flows from a consumer, regardless of transfer lane.
- Direction: Mintburn intercepts consumer → provider transfers (OnRecv on the provider) and releases from the dedicated escrow; provider → consumer transfers pass through as normal ICS‑20 sends (the consumer mints vouchers).

## Configure via genesis.json
Add allowed consumer chain‑ids under `app_state.mintburn.params.allowed_consumers`.

Example:

```
{
  "app_state": {
    "mintburn": {
      "params": {
        "allowed_consumers": ["maanydex-1"]
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
