# Restructuring

1. Configure blockrewards module -> Check
2.

# Before Live:

- Gov proposal times
- block speed
- val update block time
- test blockrewards with multiple vals

# DevOps

- loosing connection relayer

# Test IBC umaany tx

- SEND: maany-local-1 -> maanydex

  - FROM: val = maany1evvmj6yr7nv29tk7yp5z3qaru308rt5acw4mqp
  - TO: alm = maany-dex19k2p7rdqvjcm7yq57c6ntgsfna857pq79fgmpt
  - INITIAL BALANCE
    - val = 4500000.000000umaany
    - alm = 10000000.000000umaany
  - AFTER BALANCE

    - val = 4199996.000000 (sending: 300000000000umaany )
    - alm = 13000000.000000umaany

  - FROM: alm = maany-dex19k2p7rdqvjcm7yq57c6ntgsfna857pq79fgmpt
  - TO: val = maany1evvmj6yr7nv29tk7yp5z3qaru308rt5acw4mqp

`maanypd tx mintburn escrow-initial maanydex 10000000000umaany --from val --gas auto --gas-adjustment 1.8 --fees 200000umaany --chain-id maany-local-1 --keyring-backend test`
