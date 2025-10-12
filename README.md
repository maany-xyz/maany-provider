# Maany-Hub

![Maany-Banner](./.assets/banner.png)

[![Project Status: Active -- The project has reached a stable, usable state and is being actively
developed.](https://img.shields.io/badge/repo%20status-Active-green.svg)](https://www.repostatus.org/#active)
[![License: Apache-2.0](https://img.shields.io/github/license/cosmos/gaia.svg)](https://github.com/cosmos/gaia/blob/main/LICENSE)
[![Version](https://img.shields.io/github/v/release/cosmos/gaia.svg)](https://github.com/maany-xyz/maany-provider/latest)
<!-- [![Go Report Card](https://goreportcard.com/badge/github.com/cosmos/gaia)](https://goreportcard.com/report/github.com/cosmos/gaia) -->
<!-- [![GoDoc](https://img.shields.io/badge/godoc-reference-blue?logo=go)](https://pkg.go.dev/github.com/cosmos/gaia) -->
<!-- [![Lines of Code](https://sonarcloud.io/api/project_badges/measure?project=cosmos_gaia&metric=ncloc)](https://sonarcloud.io/summary/new_code?id=cosmos_gaia) -->
<!-- [![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=cosmos_gaia&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=cosmos_gaia) -->
<!-- [![Coverage](https://sonarcloud.io/api/project_badges/measure?project=cosmos_gaia&metric=coverage)](https://sonarcloud.io/summary/new_code?id=cosmos_gaia) -->
<!-- [![Twitter](https://badgen.net/badge/icon/Twitter?icon=twitter&label)](https://x.com/cosmoshub)
[![Discord](https://badgen.net/badge/icon/Discord?icon=discord&label)](https://discord.gg/interchain)
[![Cosmos Hub Forum](https://badgen.net/badge/icon/Cosmos%20Hub%20Forum?icon=atom&label)](https://forum.cosmos.network) -->

<h1 align="center">Maany Provider Chain</h1>

<p align="center">
  <b>The Core Layer of the Maany Blockchain Ecosystem</b><br>
  Securing and coordinating every Maany zone through delegated Proof of Stake and interchain security.
</p>

---

## 🌍 Overview

**Maany Provider** is the foundational blockchain of the **Maany ecosystem**, built with the [Cosmos SDK](https://docs.cosmos.network/).  
It serves as the **main hub** that secures, validates, and economically anchors all connected **Maany consumer chains**, including:

- 🪙 **Maany DEX** — the decentralized exchange consumer chain.  
- 🧩 **Maany App Chains** — individual app-specific subnets created via the Integration Framework.  

The Provider chain is responsible for:
- Staking, validator management, and governance.
- Escrow-based cross-chain minting (`x/mintburn`).
- Block reward distribution and ecosystem wallet logic.
- Chain registration, interchain security, and ICS management.

---

## ⚙️ Modules

| Module | Purpose |
|---------|----------|
| **x/mintburn** | Handles escrow and mint logic for mirrored tokens on consumer chains. |
| **x/blockrewards** | Manages validator and ecosystem reward emissions. |
| **x/autolp / x/buyback** *(consumer-side)* | Complementary modules interacting with provider policy and rewards. |

---

## 🧩 Architecture

```text
          ┌──────────────────────────┐
          │      Maany Provider      │
          │  (Delegated PoS Hub)     │
          └────────────┬─────────────┘
                       │  Interchain Security (ICS)
                       ▼
     ┌─────────────────────────────┬─────────────────────────────┐
     │        Maany DEX            │         Maany App Chains    │
     │ (Liquidity + Price Engine)  │  (Tokenized Web2 dApps)     │
     └─────────────────────────────┴─────────────────────────────┘
