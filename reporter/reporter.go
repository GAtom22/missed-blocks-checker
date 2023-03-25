package reporter

import (
	"github/GAtom22/missedblocks/config"
	"time"
)

type Reporter interface {
	Serialize(Report) string
	Init()
	Enabled() bool
	SendReport(Report) error
	Name() string
}

type Direction int

const (
	INCREASING = iota
	DECREASING
	JAILED
	UNJAILED
	TOMBSTONED
)

type ReportEntry struct {
	ValidatorAddress string
	ValidatorMoniker string
	Emoji            string
	Description      string
	MissingBlocks    int64
	Delta            int64
	Direction        Direction
}

func (r ReportEntry) GetTimeToJail(params *config.Params) time.Duration {
	blocksLeftToJail := params.MissedBlocksToJail - r.MissingBlocks
	secondsLeftToJail := params.AvgBlockTime * float64(blocksLeftToJail)

	return time.Duration(secondsLeftToJail) * time.Second
}

type Report struct {
	Entries []ReportEntry
}
