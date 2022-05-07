# Full node client of SuperBCH

This repository contains the code of the full node client of smartBCH, an EVM&amp;Web3 compatible sidechain for Bitcoin Cash.

You can get more information at [smartbch.org](https://smartbch.org).

We are actively developing smartBCH and a testnet will launch soon. Before that, you can [download the source code](https://github.com/smartbch/smartbch/releases/tag/v0.1.0) and start [a private single node testnet](https://docs.smartbch.org/smartbch/developers-guide/runsinglenode) to test your DApp.

[![Go version](https://img.shields.io/badge/go-1.16-blue.svg)](https://golang.org/)
[![API Reference](https://camo.githubusercontent.com/915b7be44ada53c290eb157634330494ebe3e30a/68747470733a2f2f676f646f632e6f72672f6769746875622e636f6d2f676f6c616e672f6764646f3f7374617475732e737667)](https://pkg.go.dev/github.com/smartbch/smartbch)
[![codecov](https://codecov.io/gh/smartbch/smartbch/branch/cover/graph/badge.svg)](https://codecov.io/gh/smartbch/smartbch)
![build workflow](https://github.com/smartbch/smartbch/actions/workflows/main.yml/badge.svg)

### Docker

To run smartBCH via `docker-compose` you can execute the commands below! Note, the first time you run docker-compose it will take a while, as it will need to build the docker image.

```
# Generate a set of 10 test keys.
docker-compose run smartbch gen-test-keys -n 10 > test-keys.txt

# Init the node, include the keys from the last step as a comma separated list.
docker-compose run smartbch init mynode --chain-id 0x2711 \
    --init-balance=10000000000000000000 \
    --test-keys=`paste -d, -s test-keys.txt` \
    --home=/root/.smartbchd --overwrite

# Generate consensus key info
CPK=$(docker-compose run -w /root/.smartbchd/ smartbch generate-consensus-key-info)
docker-compose run --entrypoint mv smartbch /root/.smartbchd/priv_validator_key.json /root/.smartbchd/config

# Generate genesis validator
K1=$(head -1 test-keys.txt)
VAL=$(docker-compose run smartbch generate-genesis-validator $K1 \
  --consensus-pubkey $CPK \
  --staking-coin 10000000000000000000000 \
  --voting-power 1 \
  --introduction "tester" \
  --home /root/.smartbchd
  )
docker-compose run smartbch add-genesis-validator --home=/root/.smartbchd $VAL

# Start it up, you are all set!
# Note that the above generated 10 accounts are not unlocked, you have to operate them through private keys
docker-compose up
```
