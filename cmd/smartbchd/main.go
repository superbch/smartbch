package main

import (
	"github.com/holiman/uint256"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/cli"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/superbch/superbch/app"
	"github.com/superbch/superbch/param"
)

type AppCreator func(logger log.Logger, chainId *uint256.Int, config *param.ChainConfig) abci.Application

func main() {
	rootCmd := createSuperbchdCmd()
	executor := cli.PrepareBaseCmd(rootCmd, "GA", DefaultNodeHome)
	err := executor.Execute()
	if err != nil {
		// handle with #870
		panic(err)
	}
}

func createSuperbchdCmd() *cobra.Command {
	cobra.EnableCommandSorting = false
	ctx := NewDefaultContext()
	rootCmd := &cobra.Command{
		Use:               "superbchd",
		Short:             "SuperBCH Chain Daemon (server)",
		PersistentPreRunE: PersistentPreRunEFn(ctx),
	}
	addInitCommands(ctx, rootCmd)
	rootCmd.AddCommand(StartCmd(ctx, newApp))
	rootCmd.AddCommand(ConfigCmd(DefaultNodeHome))
	rootCmd.AddCommand(GenerateConsensusKeyInfoCmd(ctx))
	rootCmd.AddCommand(GenerateGenesisValidatorCmd(ctx))
	rootCmd.AddCommand(AddGenesisValidatorCmd(ctx))
	rootCmd.AddCommand(StakingCmd(ctx))
	rootCmd.AddCommand(VersionCmd())
	return rootCmd
}

func addInitCommands(ctx *Context, rootCmd *cobra.Command) {
	initCmd := InitCmd(ctx, DefaultNodeHome)
	genTestKeysCmd := GenTestKeysCmd(ctx)
	rootCmd.AddCommand(initCmd, genTestKeysCmd)
}

func newApp(logger log.Logger, chainId *uint256.Int, config *param.ChainConfig) abci.Application {
	return app.NewApp(config, chainId, viper.GetInt64(flagGenesisMainnetHeight), viper.GetInt64(flagCCGenesisMainnetHeight), logger, viper.GetBool(flagSkipSanityCheck))
}
