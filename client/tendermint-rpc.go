package client

import (
	"context"

	"github.com/rs/zerolog"
	tmrpc "github.com/tendermint/tendermint/rpc/client/http"
	ctypes "github.com/tendermint/tendermint/types"
)

type TendermintRPC struct {
	URL                 string
	BlocksDiffInThePast int64
	Logger              zerolog.Logger
}

func NewTendermintRPC(url string, logger *zerolog.Logger) *TendermintRPC {
	return &TendermintRPC{
		URL:                 url,
		BlocksDiffInThePast: 10,
		Logger:              logger.With().Str("component", "rpc").Logger(),
	}
}

func (tHttp *TendermintRPC) GetAvgBlockTime() float64 {
	latestBlock := tHttp.GetBlock(nil)
	latestHeight := latestBlock.Height
	beforeLatestBlockHeight := latestBlock.Height - tHttp.BlocksDiffInThePast
	beforeLatestBlock := tHttp.GetBlock(&beforeLatestBlockHeight)

	heightDiff := float64(latestHeight - beforeLatestBlockHeight)
	timeDiff := latestBlock.Time.Sub(beforeLatestBlock.Time).Seconds()

	return timeDiff / heightDiff
}

func (tHttp *TendermintRPC) GetBlock(height *int64) *ctypes.Block {
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
