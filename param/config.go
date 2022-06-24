package param

import (
	"os"
	"path/filepath"

	"github.com/tendermint/tendermint/config"
)

const (
	DefaultRpcEthGetLogsMaxResults = 10000
	DefaultRetainBlocks            = -1
	DefaultNumKeptBlocks           = 10000
	DefaultNumKeptBlocksInMoDB     = -1
	DefaultSignatureCache          = 20000
	DefaultRecheckThreshold        = 1000
	DefaultTrunkCacheSize          = 200
	DefaultChangeRetainEveryN      = 100
	DefaultPruneEveryN             = 10

	AppDataPath    = "app"
	ModbDataPath   = "modb"
	SyncdbDataPath = "syncdb"
)

type AppConfig struct {
	//app config:
	AppDataPath    string `mapstructure:"app_data_path"`
	ModbDataPath   string `mapstructure:"modb_data_path"`
	SyncdbDataPath string `mapstructure:"syncdb_data_path"`
	// rpc config
	RpcEthGetLogsMaxResults int `mapstructure:"get_logs_max_results"`
	// tm db config
	RetainBlocks       int64 `mapstructure:"retain-blocks"`
	ChangeRetainEveryN int64 `mapstructure:"retain_interval_blocks"`
	// Use LiteDB instead of MoDB
	UseLiteDB bool `mapstructure:"use_litedb"`
	// the number of kept recent blocks for moeingads
	NumKeptBlocks int64 `mapstructure:"blocks_kept_ads"`
	// the number of kept recent blocks for moeingdb
	NumKeptBlocksInMoDB int64 `mapstructure:"blocks_kept_modb"`
	// the entry count of the signature cache
	SigCacheSize   int   `mapstructure:"sig_cache_size"`
	TrunkCacheSize int   `mapstructure:"trunk_cache_size"`
	PruneEveryN    int64 `mapstructure:"prune_every_n"`
	// How many transactions are allowed to left in the mempool
	// If more than this threshold, no further transactions can go in mempool
	RecheckThreshold int `mapstructure:"recheck_threshold"`
	//watcher config
	MainnetRPCUrl      string `mapstructure:"mainnet-rpc-url"`
	MainnetRPCUsername string `mapstructure:"mainnet-rpc-username"`
	MainnetRPCPassword string `mapstructure:"mainnet-rpc-password"`
	SuperBchRPCUrl     string `mapstructure:"superbch-rpc-url"`
	Speedup            bool   `mapstructure:"watcher-speedup"`

	FrontierGasLimit uint64 `mapstructure:"frontier-gaslimit"`

	ArchiveMode bool `mapstructure:"archive-mode"`

	WithSyncDB bool `mapstructure:"with-syncdb"`
}

type ChainConfig struct {
	NodeConfig *config.Config `mapstructure:"node_config"`
	AppConfig  *AppConfig     `mapstructure:"app_config"`
}

var (
	defaultHome = os.ExpandEnv("$HOME/.superbchd")
)

func DefaultAppConfig() *AppConfig {
	return DefaultAppConfigWithHome(defaultHome)
}
func DefaultAppConfigWithHome(home string) *AppConfig {
	if home == "" {
		home = defaultHome
	}
	return &AppConfig{
		AppDataPath:             filepath.Join(home, "data", AppDataPath),
		ModbDataPath:            filepath.Join(home, "data", ModbDataPath),
		SyncdbDataPath:          filepath.Join(home, "data", SyncdbDataPath),
		RpcEthGetLogsMaxResults: DefaultRpcEthGetLogsMaxResults,
		RetainBlocks:            DefaultRetainBlocks,
		NumKeptBlocks:           DefaultNumKeptBlocks,
		NumKeptBlocksInMoDB:     DefaultNumKeptBlocksInMoDB,
		SigCacheSize:            DefaultSignatureCache,
		RecheckThreshold:        DefaultRecheckThreshold,
		TrunkCacheSize:          DefaultTrunkCacheSize,
		ChangeRetainEveryN:      DefaultChangeRetainEveryN,
		PruneEveryN:             DefaultPruneEveryN,
		MainnetRPCPassword:      "123456",
		FrontierGasLimit:        uint64(BlockMaxGas / 200), //5Million gas
	}
}

func DefaultConfig() *ChainConfig {
	c := &ChainConfig{
		NodeConfig: config.DefaultConfig(),
		AppConfig:  DefaultAppConfig(),
	}
	c.NodeConfig.TxIndex.Indexer = "null"
	return c
}
