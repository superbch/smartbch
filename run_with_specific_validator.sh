# usage:
# ./run_with_specific_validator.sh validator_address home

set -eux

go build  -tags cppbtree -o build/superbchd ./cmd/superbchd

#todo: for test, not production
rm -rf /Users/matrix/.superbchd

./build/superbchd init freedomMan --chain-id 0x2711 \
  --init-balance=1000000000000000000000000000000 \
  --test-keys="37929f578acf92f58f14c5b9cd45ff28c2868c2ba194620238f25d354926a287" \

# shellcheck disable=SC2046
./build/superbchd add-genesis-validator $(./build/superbchd generate-genesis-validator \
  --validator-address="$1" \
  --consensus-pubkey=$(./build/superbchd generate-consensus-key-info) \
  --voting-power=1 \
  --staking-coin=1000000000000000000000 \
  --introduction="freeman") \

cp ./priv_validator_key.json /Users/matrix/.superbchd/config/

#./build/superbchd start --home="$HOME" --mainnet-url=http://135.181.219.10:8432/ --mainnet-genesis-height=20 --mainnet-user=user --mainnet-password=
#./build/superbchd start --mainnet-rpc-url=http://127.0.0.1:1234/