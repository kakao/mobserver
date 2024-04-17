package metric

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

func SnapshotStatusToPromMetrics(alloced float64) []prometheus.Metric {
	return buildPromMetrics(systemMetricPrefix, map[string]float64{"snapshot_allocation": alloced})
}

func RollbackStatusToPromMetrics(ri map[string]bool) []prometheus.Metric {
	res := []prometheus.Metric{}
	for ns, v := range ri {
		if !v {
			continue
		}

		nsLabels := strings.Split(ns, ".")

		raw := map[string]float64{"rollback_directory": 1}
		res = append(res, buildPromMetrics(systemMetricPrefix, raw, nsLabels...)...)
	}
	return res
}

type SystemStatus struct {
	Snapshot    float64 `prom:"snapshot_allocation"`
	RollbackDir float64 `prom:"rollback_directory"`
}

func (m *SystemStatus) ToPromMetrics(labelValues ...string) []prometheus.Metric {
	rawMetrics := structToMap(m)
	return buildPromMetrics(systemMetricPrefix, rawMetrics, labelValues...)
}
