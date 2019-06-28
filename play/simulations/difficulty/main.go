package main

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ellcrys/go-prompt"

	"github.com/dustin/go-humanize"

	"github.com/ellcrys/elld/play/simulations/difficulty/helpers"
	"github.com/ellcrys/elld/util/logger"
	funk "github.com/thoas/go-funk"
)

var log logger.Logger

func init() {
	log = logger.NewLogrus()
}

// DiffSim performs difficulty calculations simulation
type DiffSim struct {
	sync.Mutex
	MinMineTime int
	MaxMineTime int
	Blocks      []*helpers.Block
}

// mine mines a block
func (d *DiffSim) mine(dur time.Duration) *helpers.Block {
	now := time.Now()
	parentBlock := d.Blocks[len(d.Blocks)-1]
	diff := helpers.CalcDifficultyInception(uint64(now.Unix()), parentBlock)
	time.Sleep(dur)
	return &helpers.Block{
		Number:     parentBlock.Number + 1,
		Timestamp:  now,
		Difficulty: diff,
	}
}

func (d *DiffSim) getMinMax() (int, int) {
	d.Lock()
	defer d.Unlock()
	return d.MinMineTime, d.MaxMineTime
}

func (d *DiffSim) setMinMax(min, max int) {
	d.Lock()
	defer d.Unlock()
	d.MinMineTime = min
	d.MaxMineTime = max
}

// run runs the simulation
func (d *DiffSim) run() {
	for {
		min, max := d.getMinMax()
		mineTime := time.Duration(funk.RandomInt(min, max)) * time.Second
		parentBlock := d.Blocks[len(d.Blocks)-1]
		newBlock := d.mine(mineTime)

		big100 := new(big.Float).SetInt64(100)

		newBlockDiffFl := new(big.Float).SetInt(newBlock.Difficulty)
		parentBlockDiffFl := new(big.Float).SetInt(parentBlock.Difficulty)
		x := new(big.Float).Sub(newBlockDiffFl, parentBlockDiffFl)
		y := x.Quo(x, parentBlockDiffFl)
		diffChange := new(big.Float).Mul(y, big100)

		log.Info("Mined block: "+fmt.Sprintf("%d", newBlock.Number),
			"Difficulty", newBlock.Difficulty.Int64(),
			"DifficultyChange(%)", diffChange.String(),
			"BlockTime", newBlock.Timestamp.Unix(),
			"ParentBlockTime", parentBlock.Timestamp.Unix(),
			"TimeDiff", humanize.RelTime(parentBlock.Timestamp, newBlock.Timestamp, "ago", "from now"),
		)

		d.Blocks = append(d.Blocks, newBlock)
	}
}

func main() {
	diffSim := &DiffSim{
		MinMineTime: 1,
		MaxMineTime: 15,
	}

	// Add genesis block
	diffSim.Blocks = append(diffSim.Blocks, &helpers.Block{
		Number:     95,
		Timestamp:  time.Unix(1540742728, 0),
		Difficulty: new(big.Int).SetInt64(100000),
	})

	for {
		inp := prompt.String(">")
		if inp == "s" {
			go diffSim.run()
			continue
		}
		minMaxPars := strings.Split(inp, ",")
		min, _ := strconv.Atoi(minMaxPars[0])
		max, _ := strconv.Atoi(minMaxPars[1])
		diffSim.setMinMax(min, max)
		log.Info("Min/Max mine time changed")
	}
}
