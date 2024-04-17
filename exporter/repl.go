package exporter

import (
	"context"
	"mobserver/internal/metric"
	"mobserver/internal/model"
	"mobserver/internal/mongoutils"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
)

type replicationStatusCollector struct {
	ctx      context.Context
	base     *baseCollector
	isMongos bool
}

func newReplicationStatusCollector(client *mongo.Client, logger *logrus.Logger, isMongos bool) prometheus.Collector {
	return &replicationStatusCollector{
		ctx:      context.Background(),
		base:     newBaseCollector(client, logger),
		isMongos: isMongos,
	}
}

func (c *replicationStatusCollector) Describe(ch chan<- *prometheus.Desc) {
	c.base.Describe(ch, c.collect)
}

func (c *replicationStatusCollector) Collect(ch chan<- prometheus.Metric) {
	c.base.Collect(ch)
}

func (c *replicationStatusCollector) collect(ch chan<- prometheus.Metric) {
	if c.isMongos {
		// mongos
		for _, mt := range metric.NewRoleMetrics(model.MongoReplRoleType(-1)) {
			ch <- mt
		}
		return
	}

	replStatus, err := mongoutils.GetReplStatus(c.ctx, c.base.client)
	if err != nil {
		c.base.logger.Errorf("Failed to get replication status: %v", err)
		return
	}
	for _, mt := range metric.NewRoleMetrics(replStatus.MyState) {
		ch <- mt
	}

	replConfig, err := mongoutils.GetReplConfig(c.ctx, c.base.client)
	if err != nil {
		c.base.logger.Errorf("Failed to get replication config: %v", err)
		return
	}

	if replStatus == nil || replConfig == nil {
		return
	}

	rsMt := metric.NewReplStatus(replStatus, replConfig)
	if rsMt == nil {
		c.base.logger.Errorf("Failed to find self member in replication status")
		return
	}

	for _, mt := range rsMt.ToPromMetrics() {
		ch <- mt
	}
}
