package exporter

import (
	"context"
	"mobserver/internal/mongoutils"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
)

type Exporter struct {
	client   *mongo.Client
	clientMu sync.Mutex
	logger   *logrus.Logger
	opts     *Opts
	lock     *sync.Mutex

	isMongos bool
}

type Opts struct {
	User             string
	Password         string
	DirectConnect    bool
	ConnectTimeoutMS int
	GlobalConnPool   bool
	TimeoutOffset    int

	CollectAll             bool
	EnableReplicasetStatus bool
	EnableTopMetrics       bool
	EnableCurrentopMetrics bool
	EnableOplogStats       bool
	EnableShardingStats    bool
	EnableLVMSnapshotStats bool
	EnableRollbackStats    bool
	EnableInstanceMetrics  bool

	LVMSnapshotBackupDir string
	SlowQueryThresholdMS int

	Logger *logrus.Logger

	URI string
}

func New(opts *Opts) *Exporter {
	if opts == nil {
		opts = new(Opts)
	}

	if opts.Logger == nil {
		opts.Logger = logrus.New()
	}

	ctx := context.Background()

	exp := &Exporter{
		logger: opts.Logger,
		opts:   opts,
		lock:   &sync.Mutex{},
	}

	var cli *mongo.Client
	var err error

	for {
		cli, err = exp.getClient(ctx)
		if err != nil {
			exp.logger.Errorf("Cannot connect to MongoDB: %v", err)
			time.Sleep(5 * time.Second)
		} else {
			break
		}
	}

	if hello, err := mongoutils.GetHello(ctx, cli); err == nil {
		exp.isMongos = hello.Msg == "isdbgrid"
	}
	if cliOpts, err := mongoutils.GetCmdLineOpts(ctx, cli); err != nil {
		exp.logger.Errorf("Cannot get command line options using default slow query threshold(100ms): %v", err)
		exp.opts.SlowQueryThresholdMS = 100
	} else {
		exp.opts.SlowQueryThresholdMS = cliOpts.Parsed.OperationProfiling.SlowOpThresholdMs
	}

	return exp
}

func (e *Exporter) ValidateAndModifyOpts() {
	ctx := context.TODO()
	client, err := e.getClient(ctx)
	if err != nil {
		e.logger.Errorf("Cannot connect to MongoDB: %v", err)
		os.Exit(1)
	}

	if e.opts.CollectAll {
		e.opts.EnableReplicasetStatus = true
		e.opts.EnableTopMetrics = true
		e.opts.EnableCurrentopMetrics = true
		e.opts.EnableOplogStats = true
		e.opts.EnableShardingStats = true
		e.opts.EnableLVMSnapshotStats = true
		e.opts.EnableRollbackStats = true
		e.opts.EnableInstanceMetrics = true
	}

	if err := validateOpts(ctx, client, e.opts); err != nil {
		e.logger.Errorf("Failed to validate options: %v", err)
		os.Exit(1)
	}
}

func (e *Exporter) makeRegistry(client *mongo.Client) *prometheus.Registry {
	registry := prometheus.NewRegistry()
	if client == nil {
		return registry
	}

	if e.opts.EnableReplicasetStatus {
		registry.MustRegister(newReplicationStatusCollector(client, e.logger, e.isMongos))
	}

	if e.opts.EnableTopMetrics {
		registry.MustRegister(newTopCollector(client, e.logger))
	}

	if e.opts.EnableCurrentopMetrics {
		registry.MustRegister(newCurrentOpCollector(client, e.logger, e.opts.SlowQueryThresholdMS))
	}

	if e.opts.EnableOplogStats {
		registry.MustRegister(newOplogCollector(client, e.logger))
	}

	if e.opts.EnableLVMSnapshotStats {
		registry.MustRegister(newSnapshotCollector(client, e.logger, e.opts.LVMSnapshotBackupDir))
	}

	if e.opts.EnableRollbackStats {
		registry.MustRegister(newRollbackCollector(client, e.logger))
	}

	if e.opts.EnableShardingStats {
		registry.MustRegister(newShardingStatsCollector(client, e.logger))
	}

	if e.opts.EnableInstanceMetrics {
		registry.MustRegister(newInstanceCollector(client, e.logger))
	}

	return registry
}

func (e *Exporter) getClient(ctx context.Context) (*mongo.Client, error) {
	connOpts := mongoutils.ConnectionOpts{
		URI:              e.opts.URI,
		User:             e.opts.User,
		Password:         e.opts.Password,
		DirectConnect:    e.opts.DirectConnect,
		ConnectTimeoutMS: int64(e.opts.ConnectTimeoutMS),
	}

	if e.opts.GlobalConnPool {
		// Get global client. Maybe it must be initialized first.
		// Initialization is retried with every scrape until it succeeds once.
		e.clientMu.Lock()
		defer e.clientMu.Unlock()

		// If client is already initialized, return it.
		if e.client != nil {
			return e.client, nil
		}

		client, err := mongoutils.Connect(context.Background(), &connOpts)
		if err != nil {
			return nil, err
		}
		e.client = client

		return client, nil
	}

	// !e.opts.GlobalConnPool: create new client for every scrape.
	client, err := mongoutils.Connect(ctx, &connOpts)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// Handler returns an http.Handler that serves metrics. Can be used instead of
// run for hooking up custom HTTP servers.
func (e *Exporter) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seconds, err := strconv.Atoi(r.Header.Get("X-Prometheus-Scrape-Timeout-Seconds"))
		// To support also older ones vmagents.
		if err != nil {
			seconds = 10
		}
		seconds -= e.opts.TimeoutOffset

		var client *mongo.Client
		ctx, cancel := context.WithTimeout(r.Context(), time.Duration(seconds)*time.Second)
		defer cancel()

		filters := r.URL.Query()["collect[]"]

		requestOpts := Opts{}

		if len(filters) == 0 {
			requestOpts = *e.opts
		}

		for _, filter := range filters {
			switch filter {
			case "replicasetstatus":
				requestOpts.EnableReplicasetStatus = true
			case "oplogstatus":
				requestOpts.EnableOplogStats = true
			case "currentopmetrics":
				requestOpts.EnableCurrentopMetrics = true
			case "topmetrics":
				requestOpts.EnableTopMetrics = true
			case "lvmsnapshotstats":
				requestOpts.EnableLVMSnapshotStats = true
			case "rollbackstats":
				requestOpts.EnableRollbackStats = true
			}
		}

		client, err = e.getClient(ctx)
		if err != nil {
			e.logger.Errorf("Cannot connect to MongoDB: %v", err)
		}

		// Close client after usage.
		if !e.opts.GlobalConnPool {
			defer func() {
				if client != nil {
					err := client.Disconnect(ctx)
					if err != nil {
						e.logger.Errorf("Cannot disconnect client: %v", err)
					}
				}
			}()
		}

		var gatherers prometheus.Gatherers

		registry := e.makeRegistry(client)
		gatherers = append(gatherers, registry)

		// Delegate http serving to Prometheus client library, which will call collector.Collect.
		h := promhttp.HandlerFor(gatherers, promhttp.HandlerOpts{
			ErrorHandling: promhttp.ContinueOnError,
			ErrorLog:      e.logger,
		})

		h.ServeHTTP(w, r)
	})
}
