package main

import (
	"fmt"
	"mobserver/exporter"
	"regexp"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

//nolint:gochecknoglobals
var (
	version   string
	commit    string
	buildDate string
)

type Flags struct {
	User             string `name:"mongodb.user" help:"monitor user, need clusterMonitor role in admin db and read role in local db" env:"MONGODB_USER" placeholder:"monitorUser"`
	Password         string `name:"mongodb.password" help:"monitor user password" env:"MONGODB_PASSWORD" placeholder:"monitorPassword"`
	URI              string `name:"mongodb.uri" help:"MongoDB connection URI" env:"MONGODB_URI" placeholder:"mongodb://user:pass@127.0.0.1:27017/admin?ssl=true"`
	WebListenAddress string `name:"web.listen-address" value-name:"<:port>" description:"Address on which to expose metrics and web interface." default:":9100"`
	WebTelemetryPath string `name:"web.telemetry-path" help:"Metrics expose path" default:"/metrics"`
	TLSConfigPath    string `name:"web.config" help:"Path to the file having Prometheus TLS config for basic auth"`
	TimeoutOffset    int    `name:"web.timeout-offset" help:"Offset to subtract from the request timeout in seconds" default:"1"`
	LogLevel         string `name:"log.level" help:"Only log messages with the given severity or above. Valid levels: [debug, info, warn, error, fatal]" enum:"debug,info,warn,error,fatal" default:"error"`
	GlobalConnPool   bool   `name:"mongodb.global-conn-pool" help:"Use global connection pool instead of creating new pool for each http request." negatable:""`
	DirectConnect    bool   `name:"mongodb.direct-connect" help:"Whether or not a direct connect should be made. Direct connections are not valid if multiple hosts are specified or an SRV URI is used." default:"true" negatable:""`
	ConnectTimeoutMS int    `name:"mongodb.connect-timeout-ms" help:"Connection timeout in milliseconds" default:"5000"`

	EnableReplicasetStatus bool `name:"collector.replicasetstatus" help:"Enable collecting metrics from replSetGetStatus"`
	EnableTopMetrics       bool `name:"collector.topmetrics" help:"Enable collecting metrics from top admin command"`
	EnableCurrentopMetrics bool `name:"collector.currentopmetrics" help:"Enable collecting metrics currentop admin command"`
	EnableOplogStats       bool `name:"collector.oplogstats" help:"Enable collecting metrics from oplog"`
	EnableShardingStats    bool `name:"collector.shardstats" help:"Enable collecting metrics from shard"`
	EnableLVMSnapshotStats bool `name:"collector.lvmsnapshotstats" help:"Enable collecting metrics from lvs"`
	EnableRollbackStats    bool `name:"collector.rollbackstats" help:"Enable collecting metrics from rollback"`

	CollectAll bool `name:"collect-all" help:"Enable all collectors. Same as specifying all --collector.<name>"`

	LVMSnapshotBackupDir string `name:"lvm-backup-dir" help:"Directory to store lvm snapshot backup" placeholder:"/data/lvm-snapshot-backup-dir"`

	Version bool `name:"version" help:"Show version and exit"`
}

func main() {
	var opts Flags
	ctx := kong.Parse(&opts,
		kong.Name("mobserver"),
		kong.Description("Advanced MongoDB Prometheus exporter"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
		kong.Vars{
			"version": version,
		})

	if opts.Version {
		fmt.Println("mobserver - Advanced MongoDB Prometheus exporter")
		fmt.Printf("Version: %s\n", version)
		fmt.Printf("Commit: %s\n", commit)
		fmt.Printf("Build date: %s\n", buildDate)
		return
	}

	log := logrus.New()
	logLevel, err := logrus.ParseLevel(opts.LogLevel)
	if err != nil {
		ctx.Fatalf("Invalid log level: %s", opts.LogLevel)
	}

	log.SetLevel(logLevel)

	if opts.WebTelemetryPath == "" {
		log.Warn("Web telemetry path \"\" invalid, falling back to \"/metrics\" instead")
		opts.WebTelemetryPath = "/metrics"
	}

	cs, err := connstring.ParseAndValidate(opts.URI)
	if err != nil {
		ctx.Fatalf("Invalid MongoDB URI format: %s: %v", opts.URI, err)
	}

	if opts.User != "" && cs.Username != opts.User {
		if cs.Username != "" {
			log.Warnf("Username in URI (%s) and --mongodb.user (%s) do not match, using --mongodb.user", cs.Username, opts.User)
		}
		cs.Username = opts.User
	}

	if opts.Password != "" && cs.Password != opts.Password {
		if cs.Password != "" {
			log.Warnf("Password in URI and --mongodb.password do not match, using --mongodb.password")
		}
		cs.Password = opts.Password
	}

	if opts.DirectConnect {
		cs.DirectConnection = true
	}

	opts.URI = buildURI(&cs)

	log.Debugln("URI:", opts.URI)

	if opts.TimeoutOffset <= 0 {
		log.Warn("Timeout offset needs to be greater than \"0\", falling back to \"1\". You can specify the timout offset with --web.timeout-offset command argument")
		opts.TimeoutOffset = 1
	}

	exporterOpts := &exporter.ServerOpts{
		Path:             opts.WebTelemetryPath,
		WebListenAddress: opts.WebListenAddress,
		TLSConfigPath:    opts.TLSConfigPath,
	}

	exp := buildExporter(&opts, log)
	exp.ValidateAndModifyOpts()

	exporter.RunWebServer(exporterOpts, exp, log)
}

func buildExporter(opts *Flags, log *logrus.Logger) *exporter.Exporter {
	log.Debugf("Connection URI: %s", opts.URI)

	exporterOpts := &exporter.Opts{
		Logger:           log,
		URI:              opts.URI,
		GlobalConnPool:   opts.GlobalConnPool,
		DirectConnect:    opts.DirectConnect,
		ConnectTimeoutMS: opts.ConnectTimeoutMS,
		TimeoutOffset:    opts.TimeoutOffset,

		CollectAll: opts.CollectAll,

		EnableReplicasetStatus: opts.EnableReplicasetStatus,
		EnableOplogStats:       opts.EnableOplogStats,
		EnableCurrentopMetrics: opts.EnableCurrentopMetrics,
		EnableTopMetrics:       opts.EnableTopMetrics,
		EnableShardingStats:    opts.EnableShardingStats,
		EnableLVMSnapshotStats: opts.EnableLVMSnapshotStats,
		EnableRollbackStats:    opts.EnableRollbackStats,

		LVMSnapshotBackupDir: opts.LVMSnapshotBackupDir,
	}

	e := exporter.New(exporterOpts)

	return e
}

func buildURI(cs *connstring.ConnString) string {
	prefix := "mongodb://" // default prefix
	matchRegexp := regexp.MustCompile(`^mongodb(\+srv)?://`)
	uri := cs.String()

	// Split the uri prefix if there is any
	if matchRegexp.MatchString(uri) {
		uriArray := strings.SplitN(uri, "://", 2)
		prefix = uriArray[0] + "://"
		uri = uriArray[1]
	}

	// IF user@pass not contained in uri AND custom user and pass supplied in arguments
	// DO concat a new uri with user and pass arguments value
	if !strings.Contains(uri, "@") && cs.Username != "" && cs.Password != "" {
		// add user and pass to the uri
		uri = fmt.Sprintf("%s:%s@%s", cs.Username, cs.Password, uri)
	}

	// add back prefix after adding the user and pass
	uri = prefix + uri

	return uri
}
