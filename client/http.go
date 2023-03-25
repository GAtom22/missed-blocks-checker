package client

import (
	"context"

	"github.com/rs/zerolog"
	tmrpc "github.com/tendermint/tendermint/rpc/client/http"
	ctypes "github.com/tendermint/tendermint/types"
)

type TendermintHTTP struct {
	URL                 string
	BlocksDiffInThePast int64
	Logger              zerolog.Logger
}

func NewTendermintHTTP(url string, logger *zerolog.Logger) *TendermintHTTP {
	return &TendermintHTTP{
		URL:                 url,
		BlocksDiffInThePast: 100,
		Logger:              logger.With().Str("component", "rpc").Logger(),
	}
}

func (tHttp *TendermintHTTP) GetAvgBlockTime() float64 {
	latestBlock := tHttp.GetBlock(nil)
	latestHeight := latestBlock.Height
	beforeLatestBlockHeight := latestBlock.Height - tHttp.BlocksDiffInThePast
	beforeLatestBlock := tHttp.GetBlock(&beforeLatestBlockHeight)

	heightDiff := float64(latestHeight - beforeLatestBlockHeight)
	timeDiff := latestBlock.Time.Sub(beforeLatestBlock.Time).Seconds()

	return timeDiff / heightDiff
}

func (tHttp *TendermintHTTP) GetBlock(height *int64) *ctypes.Block {
	client, err := tmrpc.New(tHttp.URL, "/websocket")
	if err != nil {
		tHttp.Logger.Fatal().Err(err).Msg("Could not create Tendermint client")
	}

	block, err := client.Block(context.Background(), height)
	if err != nil {
		tHttp.Logger.Fatal().Err(err).Msg("Could not query Tendermint status")
	}

	return block.Block
}
