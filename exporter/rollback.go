package exporter

import (
	"bytes"
	"context"
	"fmt"
	"mobserver/internal/metric"
	"mobserver/internal/mongoutils"
	"os/exec"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
)

type rollbackCollector struct {
	ctx  context.Context
	base *baseCollector
}

func newRollbackCollector(client *mongo.Client, logger *logrus.Logger) prometheus.Collector {
	return &rollbackCollector{
		ctx:  context.Background(),
		base: newBaseCollector(client, logger),
	}
}

func (c *rollbackCollector) Describe(ch chan<- *prometheus.Desc) {
	c.base.Describe(ch, c.collect)
}

func (c *rollbackCollector) Collect(ch chan<- prometheus.Metric) {
	c.base.Collect(ch)
}

func (c *rollbackCollector) collect(ch chan<- prometheus.Metric) {
	// Rollback info represents a map of collection UUIDs
	// to a boolean value indicating whether a rollback is in progress.
	rollbackInfo, err := c.getRollbackStatus()
	if err != nil {
		c.base.logger.Errorf("Failed to get rollback status: %v", err)
		return
	}

	for _, mt := range metric.RollbackStatusToPromMetrics(rollbackInfo) {
		ch <- mt
	}
}

func (c *rollbackCollector) getRollbackStatus() (map[string]bool, error) {
	var out bytes.Buffer
	var stderr bytes.Buffer

	cmdLineOpts, err := mongoutils.GetCmdLineOpts(c.ctx, c.base.client)
	if err != nil {
		return nil, fmt.Errorf("failed to get command line options: %v", err)
	}

	rollbackDir := cmdLineOpts.Parsed.Storage.DBPath + "/rollback"

	res := make(map[string]bool)

	cmd := exec.Command("bash", "-c", fmt.Sprintf("[ -d %s ]", rollbackDir))
	if err := cmd.Run(); err != nil {
		// If the rollback directory does not exist, there is no rollback in progress.
		return res, nil
	}

	cmd = exec.Command("bash", "-c", fmt.Sprintf("ls -l %s | grep -v total | awk '{print $NF}'", rollbackDir))
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to get rollback status: %v", stderr.String())
	}

	collUUIDs := strings.Split(out.String(), "\n")
	if len(collUUIDs) == 0 {
		return res, nil
	}

	collInfo, err := mongoutils.GetAllDatabasesAndCollections(c.ctx, c.base.client)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection UUIDs: %v", err)
	}

	for _, uid := range collUUIDs {
		if uid == "" {
			continue
		}

		if ns, ok := collInfo[uid]; !ok {
			c.base.logger.Warnf("Unknown collection UUID: %s", uid)
			res["unknown.unknown"] = true
		} else {
			res[ns] = true
		}
	}

	return res, nil
}
