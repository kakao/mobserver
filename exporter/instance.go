package exporter

import (
	"context"
	"mobserver/internal/metric"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type instanceCollector struct {
	ctx  context.Context
	base *baseCollector
}

func newInstanceCollector(client *mongo.Client, logger *logrus.Logger) prometheus.Collector {
	return &instanceCollector{
		ctx:  context.Background(),
		base: newBaseCollector(client, logger),
	}
}

func (c *instanceCollector) Describe(ch chan<- *prometheus.Desc) {
	c.base.Describe(ch, c.collect)
}

func (c *instanceCollector) Collect(ch chan<- prometheus.Metric) {
	c.base.Collect(ch)
}

func (c *instanceCollector) collect(ch chan<- prometheus.Metric) {
	doc := struct {
		Version string `bson:"version"`
	}{}
	cmd := bson.D{{Key: "buildinfo", Value: 1}}
	err := c.base.client.Database("admin").RunCommand(context.TODO(), cmd).Decode(&doc)
	if err != nil {
		return
	}

	v := strings.Split(doc.Version, ".")
	majorVersion, err := strconv.ParseFloat(v[0]+"."+v[1], 64)
	if err != nil {
		logrus.Errorf("Failed to parse major version: %v", err)
		return
	}

	ist := &metric.Instance{
		Version: majorVersion,
	}

	for _, m := range ist.ToPromMetrics() {
		ch <- m
	}
}
