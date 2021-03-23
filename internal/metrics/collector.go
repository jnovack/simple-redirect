package metrics

import (
	"strings"
	"sync"
	"time"

	version "github.com/jnovack/release"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

func timeTrack(ch chan<- prometheus.Metric, start time.Time, name string) {
	elapsed := time.Since(start)
	log.Debugf("%s took %.3fs", name, float64(elapsed.Milliseconds())/1000)

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("go_task_time", "Go task elasped time", []string{}, prometheus.Labels{"task": name, "application": version.Application}),
		prometheus.GaugeValue,
		float64(elapsed.Milliseconds())/1000,
	)
}

// Collector TODO Comment
type Collector struct {
	desc string
}

// Metric TODO Comment
type Metric struct {
	name   string
	help   string
	value  float64
	labels map[string]string
}

// Describe sends the super-set of all possible descriptors of metrics collected by this Collector.
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {

	metrics := make(chan prometheus.Metric)
	go func() {
		c.Collect(metrics)
		close(metrics)
	}()
	for m := range metrics {
		ch <- m.Desc()
	}
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *Collector) Collect(ch chan<- prometheus.Metric) {

	wg := sync.WaitGroup{}

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(parseForPrometheus(version.Application), "github.com/jnovack/"+version.Application, []string{}, prometheus.Labels{"version": version.Version}),
		prometheus.GaugeValue,
		1,
	)

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer timeTrack(ch, time.Now(), "GetMetrics")
		cm := GetMetrics()
		for _, m := range cm {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(m.name, m.help, []string{}, m.labels),
				prometheus.GaugeValue,
				float64(m.value),
			)
		}

	}()

	wg.Wait()
}

// NewCollector TODO Comment
func NewCollector() *Collector {
	return &Collector{
		desc: version.Application + " Collector",
	}
}

func parseForPrometheus(incoming string) string {
	outgoing := incoming
	outgoing = strings.Replace(outgoing, "/", "_", -1)
	outgoing = strings.Replace(outgoing, " ", "_", -1)
	outgoing = strings.Replace(outgoing, "-", "_", -1)
	outgoing = strings.Replace(outgoing, ".", "_", -1)
	return outgoing
}
