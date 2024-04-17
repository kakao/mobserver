package exporter

import (
	"context"
	"errors"
	"fmt"
	"mobserver/internal/metric"
	"mobserver/internal/model"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type oplogCollector struct {
	ctx  context.Context
	base *baseCollector
}

func newOplogCollector(client *mongo.Client, logger *logrus.Logger) prometheus.Collector {
	return &oplogCollector{
		ctx:  context.Background(),
		base: newBaseCollector(client, logger),
	}
}

func (c *oplogCollector) Describe(ch chan<- *prometheus.Desc) {
	c.base.Describe(ch, c.collect)
}

func (c *oplogCollector) Collect(ch chan<- prometheus.Metric) {
	c.base.Collect(ch)
}

func (c *oplogCollector) collect(ch chan<- prometheus.Metric) {
	db := c.base.client.Database("local")
	coll := db.Collection("oplog.rs")

	var oplogSize model.CollSize
	if err := db.RunCommand(c.ctx, bson.D{{Key: "collStats", Value: "oplog.rs"}}).Decode(&oplogSize); err != nil {
		c.base.logger.Errorf("Failed to get oplog size: %v", err)
		return
	}

	// get first and last items in the oplog
	firstTs, err := getOpTimestamp(c.ctx, coll, bson.D{{Key: "$natural", Value: 1}})
	if err != nil {
		c.base.logger.Errorf("Failed to get first oplog timestamp: %v", err)
		return
	}

	lastTs, err := getOpTimestamp(c.ctx, coll, bson.D{{Key: "$natural", Value: -1}})
	if err != nil {
		c.base.logger.Errorf("Failed to get last oplog timestamp: %v", err)
		return
	}

	diff := lastTs - firstTs

	mt := metric.Oplog{
		LogSizeMB: float64(oplogSize.MaxSize),
		UsedMB:    float64(oplogSize.Size),
		FirstTS:   float64(firstTs),
		LastTS:    float64(lastTs),
		TimeDiff:  float64(diff),
	}

	for _, mt := range mt.ToPromMetrics() {
		ch <- mt
	}
}

func getOpTimestamp(ctx context.Context, collection *mongo.Collection, sort bson.D) (int64, error) {
	var limit int64 = 1

	opts := options.Find()
	opts.SetSort(sort)
	opts.SetLimit(limit)
	cursor, err := collection.Find(ctx, bson.D{}, opts)
	if err != nil {
		return 0, errors.New("objects not found in local.oplog.rs -- Is this a new and empty db instance?")
	}
	defer cursor.Close(ctx)

	var opTime model.OpTime
	var total int64 = 0

	for cursor.Next(ctx) {
		if err := cursor.Decode(&opTime); err == nil {
			total++
		}
	}
	if total != limit {
		return 0, fmt.Errorf("expected %d oplog entries, found %d", limit, total)
	}

	return int64(opTime.Ts.T), nil
}
