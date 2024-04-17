package exporter

import (
	"context"
	"mobserver/internal/metric"
	"mobserver/internal/model"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type topCollector struct {
	ctx  context.Context
	base *baseCollector
}

func newTopCollector(client *mongo.Client, logger *logrus.Logger) prometheus.Collector {
	return &topCollector{
		ctx:  context.Background(),
		base: newBaseCollector(client, logger),
	}
}

func (c *topCollector) Describe(ch chan<- *prometheus.Desc) {
	c.base.Describe(ch, c.collect)
}

func (c *topCollector) Collect(ch chan<- prometheus.Metric) {
	c.base.Collect(ch)
}

func (c *topCollector) collect(ch chan<- prometheus.Metric) {
	var result struct {
		Totals bson.M `bson:"totals"`
	}

	if err := c.base.client.Database("admin").RunCommand(c.ctx, bson.D{{Key: "top", Value: 1}}).Decode(&result); err != nil {
		c.base.logger.Errorf("Failed to get top command: %v", err)
		return
	}

	delete(result.Totals, "note")

	tmp, _ := bson.Marshal(result.Totals)
	tops := make(map[string]map[string]model.TopField)

	if err := bson.Unmarshal(tmp, &tops); err != nil {
		c.base.logger.Errorf("Failed to parse top command: %v", err)
		return
	}

	for ns, top := range tops {
		if metric.IsSystemCollection(ns) {
			continue
		}

		db, coll := metric.ParseNamespace(ns)
		metricLabels := []string{db, coll}

		for _, mt := range metric.NewTop(top).ToPromMetrics(metricLabels...) {
			ch <- mt
		}
	}
}
