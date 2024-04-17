package metric

import (
	"mobserver/internal/model"

	"github.com/prometheus/client_golang/prometheus"
)

type ShardingStats struct {
	ShardedDatabases   float64                  `prom:"sharded_databases"`
	UnshardedDatabases float64                  `prom:"unsharded_databases"`
	BalancerEnabled    float64                  `prom:"balancer_enabled"`
	Chunks             []model.ConfigChunk      `prom:"-"`
	LastMovedChunks    []model.ConfigChunkMoves `prom:"-"`
	Shards             float64                  `prom:"shards"`
	DrainingShards     float64                  `prom:"draining_shards"`
}

func (m *ShardingStats) ToPromMetrics(_ ...string) []prometheus.Metric {
	rawMetrics := structToMap(m)
	metrics := buildPromMetrics(shardingMetricPrefix, rawMetrics)
	chunkNs := make(map[string]struct{})
	movedChunksNs := make(map[string]struct{})

	for _, v := range m.Chunks {
		db, coll := ParseNamespace(v.Ns)
		shard := v.Shard

		labels := []string{db, coll, shard}

		mt := map[string]float64{
			"chunks": float64(v.NChunks),
		}
		metrics = append(metrics, buildPromMetrics(shardingMetricPrefix, mt, labels...)...)

		chunkNs[v.Ns] = struct{}{}
	}

	for _, v := range m.LastMovedChunks {
		mt := map[string]float64{
			"last_24h_chunk_moves": float64(v.NChunks),
		}
		db, coll := ParseNamespace(v.Ns)

		metrics = append(metrics, buildPromMetrics(shardingMetricPrefix, mt, db, coll)...)
		movedChunksNs[v.Ns] = struct{}{}
	}

	for ns := range chunkNs {
		db, coll := ParseNamespace(ns)
		if _, ok := movedChunksNs[ns]; ok {
			continue
		}
		mt := map[string]float64{
			"last_24h_chunk_moves": 0,
		}
		metrics = append(metrics, buildPromMetrics(shardingMetricPrefix, mt, db, coll)...)
	}

	return metrics
}
