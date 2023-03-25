package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var missedBlocks = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "tendermint_missed_blocks",
	Help: "The total number of missed blocks by validator",
},
	[]string{"validator_address"},
)

func UpdateMissedBlocks(valAddr string, delta int64) {
	if delta <= 0 {
		return
	}
	missedBlocks.WithLabelValues(valAddr).Add(float64(delta))
}
