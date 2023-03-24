package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var missedBlocks = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "missed_blocks",
	Help: "The total number of missed blocks by validator",
},
	[]string{"validator_address"},
)

func UpdateMissedBlocks(valAddr string, count int64) {
	if count <= 0 {
		return
	}
	missedBlocks.WithLabelValues(valAddr).Add(float64(count))
}
