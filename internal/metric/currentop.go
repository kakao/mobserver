package metric

import (
	"mobserver/internal/model"

	"github.com/prometheus/client_golang/prometheus"
)

type CurrentOpWithTotal struct {
	CurrentOps map[string]*CurrentOp
	Total      *CurrentOpTotal
}

type CurrentOp struct {
	SlowQueryCount             float64 `prom:"slow_query_count"`
	LongestRunningQuerySecs    float64 `prom:"longest_running_query_secs"`
	CollscanCount              float64 `prom:"collscan_count"`
	WaitingForLockCount        float64 `prom:"waiting_for_lock_count"`
	WaitingForLatchCount       float64 `prom:"waiting_for_latch_count"`
	WaitingForFlowControlCount float64 `prom:"waiting_for_flow_control_count"`
	TransactionCount           float64 `prom:"transaction_count"`
}

type CurrentOpTotal struct {
	SlowQueryCountTotal             float64 `prom:"slow_query_count_total"`
	LongestRunningSecondsTotal      float64 `prom:"longest_running_query_secs_total"`
	CollscanCountTotal              float64 `prom:"collscan_count_total"`
	WaitingForLockCountTotal        float64 `prom:"waiting_for_lock_count_total"`
	WaitingForLatchCountTotal       float64 `prom:"waiting_for_latch_count_total"`
	WaitingForFlowControlCountTotal float64 `prom:"waiting_for_flow_control_count_total"`
	TransactionCountTotal           float64 `prom:"transaction_count_total"`
}

func NewCurrentOp(ops []model.CurrentOpBatchField) *CurrentOpWithTotal {
	result := make(map[string]*CurrentOp)
	total := &CurrentOpTotal{}

	for _, op := range ops {
		if op.Ns == "" {
			op.Ns = "unknown.unknown"
		}
		if result[op.Ns] == nil {
			result[op.Ns] = &CurrentOp{}
		}

		result[op.Ns].SlowQueryCount++
		total.SlowQueryCountTotal++
		if float64(op.MicrosecsRunning) > result[op.Ns].LongestRunningQuerySecs {
			result[op.Ns].LongestRunningQuerySecs = float64(op.MicrosecsRunning) / 1000000
			if result[op.Ns].LongestRunningQuerySecs > total.LongestRunningSecondsTotal {
				total.LongestRunningSecondsTotal = result[op.Ns].LongestRunningQuerySecs
			}
		}

		if op.PlanSummary == "COLLSCAN" {
			result[op.Ns].CollscanCount++
			total.CollscanCountTotal++
		}

		if op.WaitingForLock {
			result[op.Ns].WaitingForLockCount++
			total.WaitingForLockCountTotal++
		}

		if op.WaitingForLatch != nil {
			result[op.Ns].WaitingForLatchCount++
			total.WaitingForLatchCountTotal++
		}

		if op.Transaction != nil {
			result[op.Ns].TransactionCount++
			total.TransactionCountTotal++
		}

		if op.WaitingForFlowControl {
			result[op.Ns].WaitingForFlowControlCount++
			total.WaitingForFlowControlCountTotal++
		}
	}

	return &CurrentOpWithTotal{
		CurrentOps: result,
		Total:      total,
	}
}

func (m *CurrentOpWithTotal) ToPromMetrics() []prometheus.Metric {
	var metrics []prometheus.Metric

	for ns, op := range m.CurrentOps {
		db, coll := ParseNamespace(ns)
		rawMetrics := structToMap(op)
		metrics = append(metrics, buildPromMetrics(processMetricPrefix, rawMetrics, db, coll)...)
	}

	rawMetrics := structToMap(m.Total)
	metrics = append(metrics, buildPromMetrics(processMetricPrefix, rawMetrics)...)

	return metrics
}
