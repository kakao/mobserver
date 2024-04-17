package metric

import (
	"reflect"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

func structToMap(v interface{}) map[string]float64 {
	val := reflect.ValueOf(v).Elem()
	typ := val.Type()

	m := make(map[string]float64, val.NumField())
	for i := 0; i < val.NumField(); i++ {
		t := typ.Field(i).Tag.Get("prom")
		if t == "" || t == "-" {
			continue
		}
		m[t] = val.Field(i).Float()
	}

	return m
}

func buildPromMetrics(prefix string, rawMetrics map[string]float64, labelValues ...string) []prometheus.Metric {
	var metrics []prometheus.Metric
	for k, v := range rawMetrics {
		m := Schema[prefix][k]
		name := prefix
		if k != "" {
			name += "_" + k
		}
		if len(labelValues) > 0 {
			metrics = append(metrics, m.BuildMetric(name, v, labelValues...))
		} else {
			metrics = append(metrics, m.BuildMetric(name, v))
		}
	}

	return metrics
}

func ParseNamespace(ns string) (string, string) {
	parts := strings.Split(ns, ".")
	if len(parts) < 2 {
		return parts[0], ""
	}

	return parts[0], strings.Join(parts[1:], ".")
}

func IsSystemCollection(ns string) bool {
	return strings.HasPrefix(ns, "local.") || strings.HasPrefix(ns, "config.") || strings.HasPrefix(ns, "admin.")
}
