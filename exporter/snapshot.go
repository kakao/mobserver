package exporter

import (
	"bytes"
	"fmt"
	"mobserver/internal/metric"
	"os/exec"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
)

type snapshotCollector struct {
	base *baseCollector

	snapDir string
}

func newSnapshotCollector(client *mongo.Client, logger *logrus.Logger, snapDir string) prometheus.Collector {
	return &snapshotCollector{
		base: newBaseCollector(client, logger),

		snapDir: snapDir,
	}
}

func (c *snapshotCollector) Describe(ch chan<- *prometheus.Desc) {
	c.base.Describe(ch, c.collect)
}

func (c *snapshotCollector) Collect(ch chan<- prometheus.Metric) {
	c.base.Collect(ch)
}

func (c *snapshotCollector) collect(ch chan<- prometheus.Metric) {
	snapAlloc, err := c.getSnapshotAllocatedSize()
	if err != nil {
		c.base.logger.Errorf("Failed to get snapshot allocated size: %v", err)
		return
	}

	for _, mt := range metric.SnapshotStatusToPromMetrics(snapAlloc) {
		ch <- mt
	}
}

func (c *snapshotCollector) getSnapshotAllocatedSize() (float64, error) {
	var out bytes.Buffer
	var stderr bytes.Buffer

	snap := strings.ReplaceAll(c.snapDir, "/", "\\/")

	cmd := exec.Command("bash", "-c", fmt.Sprintf("df | awk '/%s$/'", snap))
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return 0, fmt.Errorf("failed to get snapshot allocated size: %v", stderr.String())
	}

	isBackupMounted := out.String()
	snapInfo := ""

	if string(isBackupMounted) != "" {
		cmd = exec.Command("bash", "-c", "sudo lvs | awk '$6!~/[^0-9.]/&&$6>0{print$6}'")
		out.Reset()
		stderr.Reset()
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			return 0, fmt.Errorf("failed to get snapshot allocated size: %v", stderr.String())
		}
		snapInfo = strings.TrimSuffix(out.String(), "\n")
	} else {
		snapInfo = "0"
	}

	if snapInfo == "" {
		snapInfo = "0"
	}

	snapshot, err := strconv.ParseFloat(snapInfo, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse snapshot allocated size %s: %v", snapInfo, err)
	}

	return snapshot, nil
}
