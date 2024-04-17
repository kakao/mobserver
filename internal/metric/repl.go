package metric

import (
	"mobserver/internal/model"

	"github.com/prometheus/client_golang/prometheus"
)

type Repl struct {
	HeartbeatDelay    float64 `prom:"heartbeat_delay"`
	Lag               float64 `prom:"lag"`
	OddState          float64 `prom:"odd_state"`
	ElectedBeforeSecs float64 `prom:"elected_before_secs"`
	Status            float64 `prom:"status"`
	Version           float64 `prom:"version"`
	Term              float64 `prom:"term"`
	ProtocolVersion   float64 `prom:"protocolVersion"`
	ArbiterOnly       float64 `prom:"arbiterOnly"`
	BuildIndexes      float64 `prom:"buildIndexes"`
	Hidden            float64 `prom:"hidden"`
	Priority          float64 `prom:"priority"`
	Votes             float64 `prom:"votes"`
}

func NewReplStatus(rs *model.ReplSetGetStatusDoc, rc *model.ReplSetGetConfigDoc) *Repl {
	var me *model.RSStatusMemberDoc

	lastPrimaryOpTime := rs.Date
	var result = &Repl{
		Term:            float64(rc.Config.Term),
		Version:         float64(rc.Config.Version),
		ProtocolVersion: float64(rc.Config.ProtocolVersion),
		Status:          float64(rs.MyState),
		OddState:        float64(model.REPL_PRIMARY),
	}

	for _, member := range rs.Members {
		if member.Self {
			me = member
			break
		}
	}

	if me == nil {
		return nil
	}

	for _, member := range rs.Members {
		if !member.Self {
			heartbeatRecvSeconds := float64(rs.Date.Unix() - member.LastHeartbeatRecv.Unix())
			if heartbeatRecvSeconds > result.HeartbeatDelay {
				result.HeartbeatDelay = heartbeatRecvSeconds
			}
		}

		if member.State == model.REPL_PRIMARY {
			lastPrimaryOpTime = member.OpTimeDate
			result.ElectedBeforeSecs = float64(rs.Date.Unix() - member.ElectionDate.Unix())
		}
		if member.State.IsOddState() {
			result.OddState = float64(member.State)
		}
	}

	result.Lag = float64(lastPrimaryOpTime.Unix() - me.OpTimeDate.Unix())
	if result.Lag < 0 {
		result.Lag = 0
	}

	for _, member := range rc.Config.Members {
		if member.Host == me.Name {
			if member.ArbiterOnly {
				result.ArbiterOnly = 1.0
			}
			if member.BuildIndexes {
				result.BuildIndexes = 1.0
			}
			if member.Hidden {
				result.Hidden = 1.0
			}
			result.Priority = float64(member.Priority)
			result.Votes = float64(member.Votes)
			break
		}
	}

	return result
}

func (m *Repl) ToPromMetrics(labelValues ...string) []prometheus.Metric {
	rawMetrics := structToMap(m)
	return buildPromMetrics(replStatsPrefix, rawMetrics, labelValues...)
}

type Role struct {
	Role model.MongoReplRoleType `prom:"role"`
}

func NewRoleMetrics(role model.MongoReplRoleType) []prometheus.Metric {
	rawMetrics := map[string]float64{
		"primary":   0,
		"secondary": 0,
		"other":     0,
	}

	switch role {
	case model.REPL_PRIMARY:
		rawMetrics["primary"] = 1.0
	case model.REPL_SECONDARY:
		rawMetrics["secondary"] = 1.0
	default:
		rawMetrics["other"] = 1.0
	}

	res := make([]prometheus.Metric, 0)

	for k, v := range rawMetrics {
		m := Schema[roleMetricPrefix]["role"]
		name := roleMetricPrefix + "_role"
		res = append(res, m.BuildMetric(name, v, k))
	}

	return res
}
