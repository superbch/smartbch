set -eux

go build --tags cppbtree -o ./superbchd  ./cmd/superbchd
rm -rf ~/.superbchd
./superbchd init man --chain-id=0x2711
cp ./genesis.json ~/.superbchd/config/
cp ./config.toml ~/.superbchd/config/
./superbchd start --mainnet-url=http://34.88.14.23:1234 --superbch-url=http://34.88.14.23:8545 --watcher-speedup=true