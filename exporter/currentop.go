package exporter

import (
	"context"
	"mobserver/internal/metric"
	"mobserver/internal/model"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type currentOpCollector struct {
	ctx  context.Context
	base *baseCollector

	minQueryTimeMs int
}

func newCurrentOpCollector(client *mongo.Client, logger *logrus.Logger, minQueryTimeMs int) prometheus.Collector {
	return &currentOpCollector{
		ctx:            context.Background(),
		base:           newBaseCollector(client, logger),
		minQueryTimeMs: minQueryTimeMs,
	}
}

func (c *currentOpCollector) Describe(ch chan<- *prometheus.Desc) {
	c.base.Describe(ch, c.collect)
}

func (c *currentOpCollector) Collect(ch chan<- prometheus.Metric) {
	c.base.Collect(ch)
}

func (c *currentOpCollector) collect(ch chan<- prometheus.Metric) {
	rawOps, err := c.getCurrentOp()
	if err != nil {
		return
	}

	filter := func(c model.CurrentOpBatchField) bool {
		if c.Command == nil {
			return true
		}

		if _, exist := c.Command["createIndexes"]; exist {
			return false
		}

		if tr, ok := c.Command["$truncated"].(string); ok {
			if strings.HasPrefix(tr, "{ createIndexes") {
				return false
			}
		}

		if strings.Contains(c.Msg, "Index Build") {
			return false
		}

		return true
	}

	filteredCurrentOp := []model.CurrentOpBatchField{}
	for _, op := range rawOps {
		if filter(op) {
			filteredCurrentOp = append(filteredCurrentOp, op)
		}
	}

	opWithTotal := metric.NewCurrentOp(filteredCurrentOp)

	for _, mt := range opWithTotal.ToPromMetrics() {
		ch <- mt
	}
}

func (c *currentOpCollector) getCurrentOp() ([]model.CurrentOpBatchField, error) {
	systemNsRegex := "^$|^admin.*|^local.*|^config.*|^.*system\\.buckets$|^.*system\\.profile$|^.*system\\.js$|^.*system\\.views$"

	currentOp := bson.D{{Key: "$currentOp", Value: bson.D{
		{Key: "allUsers", Value: true},
		{Key: "idleConnections", Value: false},
		{Key: "idleSessions", Value: true},
		{Key: "idleCursors", Value: false},
		{Key: "localOps", Value: true},
		{Key: "truncateOps", Value: true},
	}}}

	matchStage := bson.D{{Key: "$match", Value: bson.D{
		{Key: "microsecs_running", Value: bson.D{{Key: "$gt", Value: c.minQueryTimeMs * 1000}}},
		{Key: "ns", Value: bson.D{{Key: "$not", Value: bson.D{{Key: "$regex", Value: systemNsRegex}}}}},
		{Key: "desc", Value: bson.D{{Key: "$regex", Value: "^conn"}}},
		{Key: "op", Value: bson.D{{Key: "$nin", Value: bson.A{"", "none"}}}},
	}}}

	projectionStage := bson.D{{Key: "$project", Value: bson.D{
		{Key: "_id", Value: 0},
		{Key: "microsecs_running", Value: 1},
		{Key: "op", Value: 1},
		{Key: "ns", Value: 1},
		{Key: "waitingForLock", Value: 1},
		{Key: "planSummary", Value: 1},
		{Key: "waitingForFlowControl", Value: 1},
		{Key: "waitingForLatch", Value: 1},
		{Key: "transaction", Value: 1},
		{Key: "msg", Value: 1},
		{Key: "command", Value: 1},
	}}}

	cmd := bson.D{
		{Key: "aggregate", Value: 1},
		{Key: "pipeline", Value: bson.A{currentOp, matchStage, projectionStage}},
		{Key: "allowDiskUse", Value: true},
		{Key: "cursor", Value: bson.D{}},
	}

	result := model.CurrentOp{}

	if err := c.base.client.Database("admin").RunCommand(c.ctx, cmd).Decode(&result); err != nil {
		c.base.logger.Errorf("Failed to get currentOp command: %v", err)
		return nil, err
	}

	return result.Cursor.FirstBatch, nil
}
