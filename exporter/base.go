package exporter

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
)

const defaultCacheSize = 1000

type baseCollector struct {
	client *mongo.Client
	logger *logrus.Logger

	lock         sync.Mutex
	metricsCache []prometheus.Metric
}

func newBaseCollector(client *mongo.Client, logger *logrus.Logger) *baseCollector {
	return &baseCollector{
		client: client,
		logger: logger,
	}
}

func (d *baseCollector) Describe(ch chan<- *prometheus.Desc, collect func(mCh chan<- prometheus.Metric)) {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.metricsCache = make([]prometheus.Metric, 0, defaultCacheSize)

	metrics := make(chan prometheus.Metric)
	go func() {
		collect(metrics)
		close(metrics)
	}()

	for m := range metrics {
		d.metricsCache = append(d.metricsCache, m)
		ch <- m.Desc()
	}
}

func (d *baseCollector) Collect(ch chan<- prometheus.Metric) {
	d.lock.Lock()
	defer d.lock.Unlock()

	for _, metric := range d.metricsCache {
		ch <- metric
	}
}
