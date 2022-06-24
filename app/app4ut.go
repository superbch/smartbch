package app

import (
	"github.com/holiman/uint256"
	"github.com/tendermint/tendermint/libs/log"

	modbtypes "github.com/superbch/moeingdb/types"
	moevmtc "github.com/superbch/moeingevm/evmwrap/testcase"
	stakingtypes "github.com/superbch/superbch/staking/types"
)

func (app *App) Logger() log.Logger {
	return app.logger
}

func (app *App) HistoryStore() modbtypes.DB {
	return app.historyStore
}

func (app *App) BlockNum() int64 {
	return app.block.Number
}

//nolint
func (app *App) WaitLock() { // wait for postCommit to finish
	app.mtx.Lock()
	app.mtx.Unlock()
}

func (app *App) CloseTrunk() {
	app.trunk.Close(true)
}
func (app *App) CloseTxEngineContext() {
	app.txEngine.Context().Close(false)
}

func (app *App) AddEpochForTest(e *stakingtypes.Epoch) { // breaks normal function, only used in test
	app.watcher.EpochChan <- e
}

func (app *App) AddBlockFotTest(mdbBlock *modbtypes.Block) { // breaks normal function, only used in test
	app.historyStore.AddBlock(mdbBlock, -1, nil)
	app.historyStore.AddBlock(nil, -1, nil) // To Flush
	app.publishNewBlock(mdbBlock)
}

func (app *App) SumAllBalance() *uint256.Int {
	return moevmtc.GetWorldStateFromMads(app.mads).SumAllBalance()
}

func (app *App) GetWordState() *moevmtc.WorldState {
	return moevmtc.GetWorldStateFromMads(app.mads)
}
