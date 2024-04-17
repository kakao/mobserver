package exporter

import (
	"context"
	"mobserver/internal/metric"
	"mobserver/internal/model"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type shardingStatsCollector struct {
	ctx  context.Context
	base *baseCollector
}

func newShardingStatsCollector(client *mongo.Client, logger *logrus.Logger) prometheus.Collector {
	return &shardingStatsCollector{
		ctx:  context.Background(),
		base: newBaseCollector(client, logger),
	}
}

func (c *shardingStatsCollector) Describe(ch chan<- *prometheus.Desc) {
	c.base.Describe(ch, c.collect)
}

func (c *shardingStatsCollector) Collect(ch chan<- prometheus.Metric) {
	c.base.Collect(ch)
}

func (c *shardingStatsCollector) collect(ch chan<- prometheus.Metric) {
	res := metric.ShardingStats{}

	total, draining, err := c.getShardStats()
	if err != nil {
		c.base.logger.Errorf("getShardStats() failed: %v", err)
	} else {
		res.Shards = float64(total)
		res.DrainingShards = float64(draining)
	}

	sharded, unsharded, err := c.getDatabases()
	if err != nil {
		c.base.logger.Errorf("getDatabases() failed: %v", err)
	} else {
		res.ShardedDatabases = float64(sharded)
		res.UnshardedDatabases = float64(unsharded)
	}

	balancerEnabled, err := c.getBalancerEnabled()
	if err != nil {
		c.base.logger.Errorf("getBalancerEnabled() failed: %v", err)
	} else {
		if balancerEnabled {
			res.BalancerEnabled = 1
		} else {
			res.BalancerEnabled = 0
		}
	}

	chunks, err := c.getChunkStats()
	if err != nil {
		c.base.logger.Errorf("getChunkStats() failed: %v", err)
	} else {
		res.Chunks = chunks
	}

	chunkMoves, err := c.getLastChunkMoveStats()
	if err != nil {
		c.base.logger.Errorf("getLastChunkMoveStats() failed: %v", err)
	} else {
		res.LastMovedChunks = chunkMoves
	}

	for _, mt := range res.ToPromMetrics() {
		ch <- mt
	}
}

func (c *shardingStatsCollector) getShardStats() (int, int, error) {
	res := []model.ConfigShard{}
	cursor, err := c.base.client.Database("config").Collection("shards").Find(c.ctx, bson.D{})
	if err != nil {
		return 0, 0, err
	}

	if err := cursor.All(c.ctx, &res); err != nil {
		return 0, 0, err
	}

	healthy := 0
	draining := 0
	for _, shard := range res {
		if shard.Draining {
			draining++
		} else {
			healthy++
		}
	}

	return healthy + draining, draining, nil
}

func (c *shardingStatsCollector) getDatabases() (int, int, error) {
	res := []model.ConfigDatabase{}
	cursor, err := c.base.client.Database("config").Collection("databases").Find(c.ctx, bson.D{})
	if err != nil {
		return 0, 0, err
	}

	if err := cursor.All(c.ctx, &res); err != nil {
		return 0, 0, err
	}

	sharded := 0
	unsharded := 0
	for _, db := range res {
		if db.Partitioned {
			sharded++
		} else {
			unsharded++
		}
	}

	return sharded, unsharded, nil
}

func (c *shardingStatsCollector) getBalancerEnabled() (bool, error) {
	res := model.ConfigBalancerSettings{}
	cmd := bson.D{{Key: "_id", Value: "balancer"}}
	if err := c.base.client.Database("config").Collection("settings").FindOne(c.ctx, cmd).Decode(&res); err != nil {
		return false, err
	}

	return !res.Stopped, nil
}

func (c *shardingStatsCollector) getChunkStats() ([]model.ConfigChunk, error) {
	systemNsRegex := "^$|^admin.*|^local.*|^config.*|^.*system\\.buckets$|^.*system\\.profile$|^.*system\\.js$|^.*system\\.views$"
	collInfo := []struct {
		Ns        string              `bson:"_id"`
		UUID      interface{}         `bson:"uuid"`
		Timestamp primitive.Timestamp `bson:"timestamp"`
	}{}

	cursor, err := c.base.client.Database("config").Collection("collections").Find(c.ctx, bson.D{
		{Key: "_id", Value: bson.D{{Key: "$not", Value: bson.D{{Key: "$regex", Value: systemNsRegex}}}}},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(c.ctx)

	if err := cursor.All(c.ctx, &collInfo); err != nil {
		return nil, err
	}

	res := []model.ConfigChunk{}

	for _, coll := range collInfo {
		tmpRes := []model.ConfigChunk{}

		hasTimestamp := !coll.Timestamp.IsZero()

		var matchCmd bson.D

		if hasTimestamp {
			matchCmd = bson.D{{Key: "uuid", Value: coll.UUID}}
		} else {
			matchCmd = bson.D{{Key: "ns", Value: coll.Ns}}
		}

		cmd := bson.A{
			bson.D{{Key: "$match", Value: matchCmd}},
			bson.D{{Key: "$group", Value: bson.D{
				{Key: "_id", Value: "$shard"},
				{Key: "nChunks", Value: bson.D{{Key: "$sum", Value: 1}}},
			}}},
		}

		cursor, err := c.base.client.Database("config").Collection("chunks").Aggregate(c.ctx, cmd)
		if err != nil {
			return nil, err
		}

		if err := cursor.All(c.ctx, &tmpRes); err != nil {
			return nil, err
		}

		for idx := range tmpRes {
			tmpRes[idx].Ns = coll.Ns
		}

		res = append(res, tmpRes...)
	}

	return res, nil
}

func (c *shardingStatsCollector) getLastChunkMoveStats() ([]model.ConfigChunkMoves, error) {
	t := time.Now().UTC().Add(-time.Hour * 24)

	cmd := bson.A{
		bson.D{{Key: "$match", Value: bson.D{
			{Key: "time", Value: bson.D{{Key: "$gt", Value: t}}},
			{Key: "what", Value: "moveChunk.from"},
			{Key: "details.errmsg", Value: bson.D{{Key: "$exists", Value: false}}},
			{Key: "details.note", Value: "success"},
		}}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$ns"},
			{Key: "nChunks", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "_id", Value: "$_id"},
			{Key: "nChunks", Value: "$nChunks"},
		}}},
	}

	res := []model.ConfigChunkMoves{}
	cursor, err := c.base.client.Database("config").Collection("changelog").Aggregate(c.ctx, cmd)
	if err != nil {
		return nil, err
	}

	if err := cursor.All(c.ctx, &res); err != nil {
		return nil, err
	}

	return res, nil
}
