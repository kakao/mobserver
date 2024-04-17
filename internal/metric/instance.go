package metric

import "github.com/prometheus/client_golang/prometheus"

type Instance struct {
	Version float64 `prom:"version"`
}

func (m *Instance) ToPromMetrics() []prometheus.Metric {
	rawMetrics := structToMap(m)
	return buildPromMetrics(instanceMetricPrefix, rawMetrics)
}
