package metric

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	topMetricPrefix      = "mongodb_top"
	replStatsPrefix      = "mongodb_replstats"
	roleMetricPrefix     = "mongodb_repl"
	oplogMetricPrefix    = "mongodb_oplogstats"
	processMetricPrefix  = "mongodb_process"
	systemMetricPrefix   = "mongodb_system"
	shardingMetricPrefix = "mongodb_config"
	instanceMetricPrefix = "mongodb_instance"
)

type Metric struct {
	Help        string
	LabelNames  []string
	PmValueType prometheus.ValueType
}

func (m *Metric) BuildMetric(name string, value float64, labelValues ...string) prometheus.Metric {
	return prometheus.MustNewConstMetric(
		prometheus.NewDesc(name, m.Help, m.LabelNames, nil),
		m.PmValueType,
		value,
		labelValues...,
	)
}

var Schema = map[string]map[string]Metric{
	// Metadata for replstatus metrics
	replStatsPrefix: {
		"heartbeat_delay": {
			Help:        "Heartbeat delay among relica-set members",
			PmValueType: prometheus.GaugeValue,
		},
		"lag": {
			Help:        "Replication optime lag in second",
			PmValueType: prometheus.GaugeValue,
		},
		"odd_state": {
			Help:        "State of abnormal odd-state member (abnormal means one of the memeber in primary, secondary, startup2 or arbiter)",
			PmValueType: prometheus.GaugeValue,
		},
		"elected_before_secs": {
			Help:        "Elapsed seconds after selected as primary",
			PmValueType: prometheus.GaugeValue,
		},
		"status": {
			Help:        "Replication status",
			PmValueType: prometheus.GaugeValue,
		},
		"version": {
			Help:        "Version in repl config",
			PmValueType: prometheus.GaugeValue,
		},
		"term": {
			Help:        "Term in repl config",
			PmValueType: prometheus.GaugeValue,
		},
		"protocolVersion": {
			Help:        "ProtocolVersion in repl config",
			PmValueType: prometheus.GaugeValue,
		},
		"arbiterOnly": {
			Help:        "ArbiterOnly in repl config",
			PmValueType: prometheus.GaugeValue,
		},
		"buildIndexes": {
			Help:        "BuildIndexes in repl config",
			PmValueType: prometheus.GaugeValue,
		},
		"hidden": {
			Help:        "Hidden in repl config",
			PmValueType: prometheus.GaugeValue,
		},
		"priority": {
			Help:        "Priority in repl config",
			PmValueType: prometheus.GaugeValue,
		},
		"slaveDelay": {
			Help:        "SlaveDelay in repl config",
			PmValueType: prometheus.GaugeValue,
		},
		"votes": {
			Help:        "Votes in repl config",
			PmValueType: prometheus.GaugeValue,
		},
	},

	// Metadata for role metrics
	roleMetricPrefix: {
		"role": {
			Help:        "Replication role",
			LabelNames:  []string{"role"},
			PmValueType: prometheus.GaugeValue,
		},
	},

	// Metadata for oplog metrics
	oplogMetricPrefix: {
		"logSizeMB": {
			Help:        "Allocated size for oplog",
			PmValueType: prometheus.GaugeValue,
		},
		"usedMB": {
			Help:        "Used size for oplog",
			PmValueType: prometheus.GaugeValue,
		},
		"timeDiff": {
			Help:        "Term of retainment",
			PmValueType: prometheus.GaugeValue,
		},
		"firstTs": {
			Help:        "The first time of oplog",
			PmValueType: prometheus.GaugeValue,
		},
		"lastTs": {
			Help:        "The last time of oplog",
			PmValueType: prometheus.GaugeValue,
		},
	},

	// Metadata for top metrics
	topMetricPrefix: {
		"insert_count": {
			Help:        "Usage statistics for insert count",
			LabelNames:  []string{"database", "collection"},
			PmValueType: prometheus.CounterValue,
		},
		"insert_time": {
			Help:        "Usage statistics for insert time",
			LabelNames:  []string{"database", "collection"},
			PmValueType: prometheus.CounterValue,
		},
		"queries_count": {
			Help:        "Usage statistics for queries count",
			LabelNames:  []string{"database", "collection"},
			PmValueType: prometheus.CounterValue,
		},
		"queries_time": {
			Help:        "Usage statistics for queries time",
			LabelNames:  []string{"database", "collection"},
			PmValueType: prometheus.CounterValue,
		},
		"update_count": {
			Help:        "Usage statistics for update count",
			LabelNames:  []string{"database", "collection"},
			PmValueType: prometheus.CounterValue,
		},
		"update_time": {
			Help:        "Usage statistics for update time",
			LabelNames:  []string{"database", "collection"},
			PmValueType: prometheus.CounterValue,
		},
		"remove_count": {
			Help:        "Usage statistics for remove count",
			LabelNames:  []string{"database", "collection"},
			PmValueType: prometheus.CounterValue,
		},
		"remove_time": {
			Help:        "Usage statistics for remove time",
			LabelNames:  []string{"database", "collection"},
			PmValueType: prometheus.CounterValue,
		},
		"getmore_count": {
			Help:        "Usage statistics for getmore count",
			LabelNames:  []string{"database", "collection"},
			PmValueType: prometheus.CounterValue,
		},
		"getmore_time": {
			Help:        "Usage statistics for getmore time",
			LabelNames:  []string{"database", "collection"},
			PmValueType: prometheus.CounterValue,
		},
		"commands_count": {
			Help:        "Usage statistics for commands count",
			LabelNames:  []string{"database", "collection"},
			PmValueType: prometheus.CounterValue,
		},
		"commands_time": {
			Help:        "Usage statistics for commands time",
			LabelNames:  []string{"database", "collection"},
			PmValueType: prometheus.CounterValue,
		},
	},

	// Metadata for process metrics
	processMetricPrefix: {
		"slow_query_count_total": {
			Help:        "Total long query counter (over 1 second)",
			PmValueType: prometheus.GaugeValue,
		},
		"slow_query_count": {
			Help:        "Long query counter (over 1 second)",
			LabelNames:  []string{"database", "collection"},
			PmValueType: prometheus.GaugeValue,
		},
		"longest_running_query_secs_total": {
			Help:        "Total longest running seconds",
			PmValueType: prometheus.GaugeValue,
		},
		"longest_running_query_secs": {
			Help:        "Longest running seconds",
			LabelNames:  []string{"database", "collection"},
			PmValueType: prometheus.GaugeValue,
		},
		"collscan_count_total": {
			Help:        "Total number of collscan",
			PmValueType: prometheus.GaugeValue,
		},
		"collscan_count": {
			Help:        "The number of collscan",
			LabelNames:  []string{"database", "collection"},
			PmValueType: prometheus.GaugeValue,
		},
		"waiting_for_lock_count_total": {
			Help:        "Total number of waiting lock",
			PmValueType: prometheus.GaugeValue,
		},
		"waiting_for_lock_count": {
			Help:        "The number of waiting lock",
			LabelNames:  []string{"database", "collection"},
			PmValueType: prometheus.GaugeValue,
		},
		"waiting_for_latch_count_total": {
			Help:        "The number of waiting latch",
			PmValueType: prometheus.GaugeValue,
		},
		"waiting_for_latch_count": {
			Help:        "The number of waiting latch",
			LabelNames:  []string{"database", "collection"},
			PmValueType: prometheus.GaugeValue,
		},
		"waiting_for_flow_control_count_total": {
			Help:        "Total number of waiting flow control",
			PmValueType: prometheus.GaugeValue,
		},
		"waiting_for_flow_control_count": {
			Help:        "The number of waiting flow control",
			LabelNames:  []string{"database", "collection"},
			PmValueType: prometheus.GaugeValue,
		},
		"transaction_count_total": {
			Help:        "Total number of transactions",
			PmValueType: prometheus.GaugeValue,
		},
		"transaction_count": {
			Help:        "The number of transactions",
			LabelNames:  []string{"database", "collection"},
			PmValueType: prometheus.GaugeValue,
		},
	},

	// Metadata for sharding metrics
	shardingMetricPrefix: {
		"sharded_databases": {
			Help:        "Number of sharded databases",
			PmValueType: prometheus.GaugeValue,
		},
		"unsharded_databases": {
			Help:        "Number of unsharded databases",
			PmValueType: prometheus.GaugeValue,
		},
		"balancer_enabled": {
			Help:        "Balancer is enabled",
			PmValueType: prometheus.GaugeValue,
		},
		"draining_shards": {
			Help:        "Number of draining shards",
			PmValueType: prometheus.GaugeValue,
		},
		"shards": {
			Help:        "Number of shards",
			PmValueType: prometheus.GaugeValue,
		},
		"chunks": {
			Help:        "Number of chunks for each collection",
			LabelNames:  []string{"database", "collection", "shard"},
			PmValueType: prometheus.GaugeValue,
		},
		"last_24h_chunk_moves": {
			Help:        "Number of chunk moves in last 24 hours",
			LabelNames:  []string{"database", "collection"},
			PmValueType: prometheus.GaugeValue,
		},
	},

	// Metadata for system metrics
	systemMetricPrefix: {
		"snapshot_allocation": {
			Help:        "The snapshot_allocation value is the current snap_allocation.",
			PmValueType: prometheus.GaugeValue,
		},
		"rollback_directory": {
			Help:        "The rollback files exist or not",
			LabelNames:  []string{"database", "collection"},
			PmValueType: prometheus.GaugeValue,
		},
	},

	// Metadata for instance metrics
	instanceMetricPrefix: {
		"version": {
			Help:        "The version of MongoDB",
			PmValueType: prometheus.GaugeValue,
		},
	},
}
