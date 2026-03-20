package metrics

import (
	"strconv"
	"sync/atomic"

	version "github.com/jnovack/release"
)

var (
	Target         = ""
	Status         = 0
	HTTPRedirects  int64
	HTTPSRedirects int64
)

func GetMetrics() []Metric {
	var metrics []Metric

	metrics = append(metrics, Metric{
		name:   parseForPrometheus(version.Application) + "_redirects",
		help:   "Total number of redirects performed",
		value:  float64(atomic.LoadInt64(&HTTPRedirects)),
		labels: map[string]string{"protocol": "http", "target": Target, "status": strconv.Itoa(Status)},
	})
	metrics = append(metrics, Metric{
		name:   parseForPrometheus(version.Application) + "_redirects",
		help:   "Total number of redirects performed",
		value:  float64(atomic.LoadInt64(&HTTPSRedirects)),
		labels: map[string]string{"protocol": "https", "target": Target, "status": strconv.Itoa(Status)},
	})

	return metrics
}
