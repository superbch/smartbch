package watcher

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/tendermint/tendermint/libs/log"

	cctypes "github.com/superbch/superbch/crosschain/types"
	"github.com/superbch/superbch/param"
	stakingtypes "github.com/superbch/superbch/staking/types"
	"github.com/superbch/superbch/watcher/types"
)

const (
	NumBlocksToClearMemory = 1000
	WaitingBlockDelayTime  = 2
)

// A watcher watches the new blocks generated on bitcoin cash's mainnet, and
// outputs epoch information through a channel
type Watcher struct {
	logger log.Logger

	rpcClient         types.RpcClient
	superBchRpcClient types.RpcClient

	latestFinalizedHeight int64

	heightToFinalizedBlock map[int64]*types.BCHBlock

	EpochChan          chan *stakingtypes.Epoch
	epochList          []*stakingtypes.Epoch
	numBlocksInEpoch   int64
	lastEpochEndHeight int64
	lastKnownEpochNum  int64

	CCEpochChan          chan *cctypes.CCEpoch
	lastCCEpochEndHeight int64
	numBlocksInCCEpoch   int64
	ccEpochList          []*cctypes.CCEpoch
	lastKnownCCEpochNum  int64

	numBlocksToClearMemory int
	waitingBlockDelayTime  int
	parallelNum            int

	chainConfig *param.ChainConfig
}

func NewWatcher(logger log.Logger, lastHeight, lastCCEpochEndHeight int64, lastKnownEpochNum int64, chainConfig *param.ChainConfig) *Watcher {
	return &Watcher{
		logger: logger,

		rpcClient:         NewRpcClient(chainConfig.AppConfig.MainnetRPCUrl, chainConfig.AppConfig.MainnetRPCUsername, chainConfig.AppConfig.MainnetRPCPassword, "text/plain;", logger),
		superBchRpcClient: NewRpcClient(chainConfig.AppConfig.SuperBchRPCUrl, "", "", "application/json", logger),

		lastEpochEndHeight:    lastHeight,
		latestFinalizedHeight: lastHeight,
		lastKnownEpochNum:     lastKnownEpochNum,

		heightToFinalizedBlock: make(map[int64]*types.BCHBlock),
		epochList:              make([]*stakingtypes.Epoch, 0, 10),

		EpochChan: make(chan *stakingtypes.Epoch, 10000),

		numBlocksInEpoch:       param.StakingNumBlocksInEpoch,
		numBlocksToClearMemory: NumBlocksToClearMemory,
		waitingBlockDelayTime:  WaitingBlockDelayTime,

		CCEpochChan:          make(chan *cctypes.CCEpoch, 96*10000),
		ccEpochList:          make([]*cctypes.CCEpoch, 0, 40),
		lastCCEpochEndHeight: lastCCEpochEndHeight,
		numBlocksInCCEpoch:   param.BlocksInCCEpoch,

		parallelNum: 10,
		chainConfig: chainConfig,
	}
}

func (watcher *Watcher) SetNumBlocksInEpoch(n int64) {
	watcher.numBlocksInEpoch = n
}

func (watcher *Watcher) SetNumBlocksToClearMemory(n int) {
	watcher.numBlocksToClearMemory = n
}

func (watcher *Watcher) SetWaitingBlockDelayTime(n int) {
	watcher.waitingBlockDelayTime = n
}

// The main function to do a watcher's job. It must be run as a goroutine
func (watcher *Watcher) Run(catchupChan chan bool) {
	if watcher.rpcClient == (*RpcClient)(nil) {
		//for ut
		catchupChan <- true
		return
	}
	latestFinalizedHeight := watcher.latestFinalizedHeight
	latestMainnetHeight := watcher.rpcClient.GetLatestHeight(true)
	latestFinalizedHeight = watcher.epochSpeedup(latestFinalizedHeight, latestMainnetHeight)
	watcher.fetchBlocks(catchupChan, latestFinalizedHeight, latestMainnetHeight)
}

func (watcher *Watcher) fetchBlocks(catchupChan chan bool, latestFinalizedHeight, latestMainnetHeight int64) {
	catchup := false
	for {
		if !catchup && latestMainnetHeight <= latestFinalizedHeight+9 {
			latestMainnetHeight = watcher.rpcClient.GetLatestHeight(true)
			if latestMainnetHeight <= latestFinalizedHeight+9 {
				watcher.logger.Debug("Catchup")
				catchup = true
				catchupChan <- true
				close(catchupChan)
			}
		}
		latestFinalizedHeight++
		latestMainnetHeight = watcher.rpcClient.GetLatestHeight(true)
		//10 confirms
		if latestMainnetHeight < latestFinalizedHeight+9 {
			watcher.suspended(time.Duration(watcher.waitingBlockDelayTime) * time.Second) //delay half of bch mainnet block intervals
			latestFinalizedHeight--
			continue
		}
		for latestFinalizedHeight+9 <= latestMainnetHeight {
			fmt.Printf("latestFinalizedHeight:%d,latestMainnetHeight:%d\n", latestFinalizedHeight, latestMainnetHeight)
			if latestFinalizedHeight+9+int64(watcher.parallelNum) <= latestMainnetHeight {
				watcher.parallelFetchBlocks(latestFinalizedHeight)
				latestFinalizedHeight += int64(watcher.parallelNum)
			} else {
				blk := watcher.rpcClient.GetBlockByHeight(latestFinalizedHeight, true)
				if blk == nil {
					//todo: panic it
					fmt.Printf("get block:%d failed\n", latestFinalizedHeight)
					latestFinalizedHeight--
					continue
				}
				watcher.addFinalizedBlock(blk)
				latestFinalizedHeight++
			}
		}
		latestFinalizedHeight--
	}
}

func (watcher *Watcher) parallelFetchBlocks(latestFinalizedHeight int64) {
	fmt.Printf("begin paralell fetch blocks\n")
	var blockSet = make([]*types.BCHBlock, watcher.parallelNum)
	var w sync.WaitGroup
	w.Add(watcher.parallelNum)
	for i := 0; i < watcher.parallelNum; i++ {
		go func(index int) {
			blockSet[index] = watcher.rpcClient.GetBlockByHeight(latestFinalizedHeight+int64(index), true)
			w.Done()
		}(i)
	}
	w.Wait()
	fmt.Printf("after paralell fetch blocks\n")
	for _, blk := range blockSet {
		watcher.addFinalizedBlock(blk)
	}
	watcher.logger.Debug("Get bch mainnet block", "latestFinalizedHeight", latestFinalizedHeight)
}

func (watcher *Watcher) epochSpeedup(latestFinalizedHeight, latestMainnetHeight int64) int64 {
	if watcher.chainConfig.AppConfig.Speedup {
		start := uint64(watcher.lastKnownEpochNum) + 1
		for {
			if latestMainnetHeight < latestFinalizedHeight+watcher.numBlocksInEpoch {
				watcher.ccEpochSpeedup()
				break
			}
			epochs := watcher.superBchRpcClient.GetEpochs(start, start+100)
			if len(epochs) == 0 {
				watcher.ccEpochSpeedup()
				break
			}
			for _, e := range epochs {
				out, _ := json.Marshal(e)
				fmt.Println(string(out))
			}
			watcher.epochList = append(watcher.epochList, epochs...)
			for _, e := range epochs {
				if e.EndTime != 0 {
					watcher.EpochChan <- e
				}
			}
			latestFinalizedHeight += int64(len(epochs)) * watcher.numBlocksInEpoch
			start = start + uint64(len(epochs))
		}
		watcher.latestFinalizedHeight = latestFinalizedHeight
		watcher.lastEpochEndHeight = latestFinalizedHeight
		watcher.logger.Debug("After speedup", "latestFinalizedHeight", watcher.latestFinalizedHeight)
	}
	return latestFinalizedHeight
}

func (watcher *Watcher) ccEpochSpeedup() {
	if !param.ShaGateSwitch {
		return
	}
	start := uint64(watcher.lastKnownCCEpochNum) + 1
	for {
		epochs := watcher.superBchRpcClient.GetCCEpochs(start, start+100)
		if len(epochs) == 0 {
			break
		}
		for _, e := range epochs {
			out, _ := json.Marshal(e)
			fmt.Println(string(out))
		}
		watcher.ccEpochList = append(watcher.ccEpochList, epochs...)
		for _, e := range epochs {
			watcher.CCEpochChan <- e
		}
		start = start + uint64(len(epochs))
	}
}

func (watcher *Watcher) suspended(delayDuration time.Duration) {
	time.Sleep(delayDuration)
}

// Record new block and if the blocks for a new epoch is all ready, output the new epoch
func (watcher *Watcher) addFinalizedBlock(blk *types.BCHBlock) {
	watcher.heightToFinalizedBlock[blk.Height] = blk
	watcher.latestFinalizedHeight++

	if watcher.latestFinalizedHeight-watcher.lastEpochEndHeight == watcher.numBlocksInEpoch {
		watcher.generateNewEpoch()
	}
	//if watcher.latestFinalizedHeight-watcher.lastCCEpochEndHeight == watcher.numBlocksInCCEpoch {
	//	watcher.generateNewCCEpoch()
	//}
}

// Generate a new block's information
func (watcher *Watcher) generateNewEpoch() {
	epoch := watcher.buildNewEpoch()
	watcher.epochList = append(watcher.epochList, epoch)
	watcher.logger.Debug("Generate new epoch", "epochNumber", epoch.Number, "startHeight", epoch.StartHeight)
	watcher.EpochChan <- epoch
	watcher.lastEpochEndHeight = watcher.latestFinalizedHeight
	watcher.ClearOldData()
}

func (watcher *Watcher) buildNewEpoch() *stakingtypes.Epoch {
	epoch := &stakingtypes.Epoch{
		StartHeight: watcher.lastEpochEndHeight + 1,
		Nominations: make([]*stakingtypes.Nomination, 0, 10),
	}
	var valMapByPubkey = make(map[[32]byte]*stakingtypes.Nomination)
	for i := epoch.StartHeight; i <= watcher.latestFinalizedHeight; i++ {
		blk, ok := watcher.heightToFinalizedBlock[i]
		if !ok {
			panic("Missing Block")
		}
		//Please note that BCH's timestamp is not always linearly increasing
		if epoch.EndTime < blk.Timestamp {
			epoch.EndTime = blk.Timestamp
		}
		for _, nomination := range blk.Nominations {
			if _, ok := valMapByPubkey[nomination.Pubkey]; !ok {
				valMapByPubkey[nomination.Pubkey] = &nomination
			}
			valMapByPubkey[nomination.Pubkey].NominatedCount += nomination.NominatedCount
		}
	}
	for _, v := range valMapByPubkey {
		epoch.Nominations = append(epoch.Nominations, v)
	}
	sortEpochNominations(epoch)
	return epoch
}

func (watcher *Watcher) GetCurrEpoch() *stakingtypes.Epoch {
	return watcher.buildNewEpoch()
}

//func (watcher *Watcher) generateNewCCEpoch() {
//	if !watcher.chainConfig.ShaGateSwitch {
//		return
//	}
//	epoch := watcher.buildNewCCEpoch()
//	watcher.ccEpochList = append(watcher.ccEpochList, epoch)
//	watcher.logger.Debug("Generate new cc epoch", "epochNumber", epoch.Number, "startHeight", epoch.StartHeight)
//	watcher.CCEpochChan <- epoch
//	watcher.lastCCEpochEndHeight = watcher.latestFinalizedHeight
//}

//func (watcher *Watcher) buildNewCCEpoch() *cctypes.CCEpoch {
//	epoch := &cctypes.CCEpoch{
//		StartHeight:   watcher.lastCCEpochEndHeight + 1,
//		TransferInfos: make([]*cctypes.CCTransferInfo, 0, 10),
//	}
//	for i := epoch.StartHeight; i <= watcher.latestFinalizedHeight; i++ {
//		blk, ok := watcher.heightToFinalizedBlock[i]
//		if !ok {
//			panic("Missing Block")
//		}
//		if epoch.EndTime < blk.Timestamp {
//			epoch.EndTime = blk.Timestamp
//		}
//		//epoch.TransferInfos = append(epoch.TransferInfos, blk.CCTransferInfos...)
//	}
//	return epoch
//}

func (watcher *Watcher) CheckSanity(skipCheck bool) {
	if !skipCheck {
		latestHeight := watcher.rpcClient.GetLatestHeight(false)
		if latestHeight <= 0 {
			panic("Watcher GetLatestHeight failed in Sanity Check")
		}
		blk := watcher.rpcClient.GetBlockByHeight(latestHeight, false)
		if blk == nil {
			panic("Watcher GetBlockByHeight failed in Sanity Check")
		}
	}
}

//sort by pubkey (small to big) first; then sort by nominationCount;
//so nominations sort by NominationCount, if count is equal, smaller pubkey stand front
func sortEpochNominations(epoch *stakingtypes.Epoch) {
	sort.Slice(epoch.Nominations, func(i, j int) bool {
		return bytes.Compare(epoch.Nominations[i].Pubkey[:], epoch.Nominations[j].Pubkey[:]) < 0
	})
	sort.SliceStable(epoch.Nominations, func(i, j int) bool {
		return epoch.Nominations[i].NominatedCount > epoch.Nominations[j].NominatedCount
	})
}

func (watcher *Watcher) ClearOldData() {
	elLen := len(watcher.epochList)
	if elLen == 0 {
		return
	}
	height := watcher.epochList[elLen-1].StartHeight
	height -= 5 * watcher.numBlocksInEpoch
	for {
		_, ok := watcher.heightToFinalizedBlock[height]
		if !ok {
			break
		}
		delete(watcher.heightToFinalizedBlock, height)
		height--
	}
	if elLen > 5 /*param it*/ {
		watcher.epochList = watcher.epochList[elLen-5:]
	}
	ccEpochLen := len(watcher.ccEpochList)
	if ccEpochLen > 5*int(param.StakingNumBlocksInEpoch/param.BlocksInCCEpoch) {
		watcher.epochList = watcher.epochList[ccEpochLen-5:]
	}
}
