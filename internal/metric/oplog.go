package metric

import "github.com/prometheus/client_golang/prometheus"

type Oplog struct {
	LogSizeMB float64 `prom:"logSizeMB"`
	UsedMB    float64 `prom:"usedMB"`
	FirstTS   float64 `prom:"firstTs"`
	LastTS    float64 `prom:"lastTs"`
	TimeDiff  float64 `prom:"timeDiff"`
}

func (m *Oplog) ToPromMetrics(labelValues ...string) []prometheus.Metric {
	rawMetrics := structToMap(m)
	return buildPromMetrics(oplogMetricPrefix, rawMetrics, labelValues...)
}
