package metric

import (
	"mobserver/internal/model"

	"github.com/prometheus/client_golang/prometheus"
)

type Top struct {
	InsertCount float64 `prom:"insert_count"`
	InsertTime  float64 `prom:"insert_time"`

	QueriesCount float64 `prom:"queries_count"`
	QueriesTime  float64 `prom:"queries_time"`

	UpdateCount float64 `prom:"update_count"`
	UpdateTime  float64 `prom:"update_time"`

	RemoveCount float64 `prom:"remove_count"`
	RemoveTime  float64 `prom:"remove_time"`

	GetmoreCount float64 `prom:"getmore_count"`
	GetmoreTime  float64 `prom:"getmore_time"`

	CommandsCount float64 `prom:"commands_count"`
	CommandsTime  float64 `prom:"commands_time"`
}

func NewTop(src map[string]model.TopField) *Top {
	t := &Top{}
	for k, v := range src {
		switch k {
		case "insert":
			t.InsertCount = float64(v.Count)
			t.InsertTime = float64(v.Time)
		case "queries":
			t.QueriesCount = float64(v.Count)
			t.QueriesTime = float64(v.Time)
		case "update":
			t.UpdateCount = float64(v.Count)
			t.UpdateTime = float64(v.Time)
		case "remove":
			t.RemoveCount = float64(v.Count)
			t.RemoveTime = float64(v.Time)
		case "getmore":
			t.GetmoreCount = float64(v.Count)
			t.GetmoreTime = float64(v.Time)
		case "commands":
			t.CommandsCount = float64(v.Count)
			t.CommandsTime = float64(v.Time)
		}
	}
	return t
}

func (m *Top) ToPromMetrics(labelValues ...string) []prometheus.Metric {
	rawMetrics := structToMap(m)
	return buildPromMetrics(topMetricPrefix, rawMetrics, labelValues...)
}
