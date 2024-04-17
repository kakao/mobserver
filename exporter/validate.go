package exporter

import (
	"context"
	"errors"
	"fmt"
	"mobserver/internal/mongoutils"
	"os"
	"os/exec"
	"strings"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

func validateOpts(ctx context.Context, client *mongo.Client, opts *Opts) error {
	if err := validateGeneralOpts(opts); err != nil {
		return fmt.Errorf("failed to validate general options: %w", err)
	}

	if err := validateLocalhostOpts(opts); err != nil {
		return fmt.Errorf("failed to validate localhost options: %w", err)
	}

	if err := validateToplogyOpts(ctx, client, opts); err != nil {
		return fmt.Errorf("failed to validate topology options: %w", err)
	}

	if err := validateSnapshotStats(opts); err != nil {
		return fmt.Errorf("failed to validate snapshot status options: %w", err)
	}

	return nil
}

func validateGeneralOpts(opts *Opts) error {
	if !opts.EnableLVMSnapshotStats && !opts.EnableRollbackStats {
		return nil
	}

	checkCommands := []string{
		"ls",
		"awk",
		"grep",
	}

	for _, cmd := range checkCommands {
		if _, err := exec.LookPath(cmd); err != nil {
			if errors.Is(err, exec.ErrNotFound) {
				opts.Logger.Warnf("Cannot find %s in PATH. Disabling rollback, snapshot status metrics", cmd)
				opts.EnableRollbackStats = false
				opts.EnableLVMSnapshotStats = false
			} else {
				return fmt.Errorf("failed to check for %s: %w", cmd, err)
			}
		}
	}

	return nil
}

func validateSnapshotStats(opts *Opts) error {
	if !opts.EnableLVMSnapshotStats {
		return nil
	}

	checkCommands := []string{
		"df",
		"lvs",
	}

	for _, cmd := range checkCommands {
		if _, err := exec.LookPath(cmd); err != nil {
			if errors.Is(err, exec.ErrNotFound) {
				opts.Logger.Warnf("Cannot find %s in PATH. Disabling snapshot status metrics", cmd)
				opts.EnableLVMSnapshotStats = false
			} else {
				return fmt.Errorf("failed to check for %s: %w", cmd, err)
			}
		}
	}

	return nil
}

func validateToplogyOpts(ctx context.Context, client *mongo.Client, opts *Opts) error {
	hello, err := mongoutils.GetHello(ctx, client)
	if err != nil {
		return fmt.Errorf("failed to get hello result: %w", err)
	}

	if hello.ArbiterOnly {
		opts.Logger.Warnf("This is an arbiter, disabling all metrics except for replset status")
		opts.EnableTopMetrics = false
		opts.EnableCurrentopMetrics = false
		opts.EnableOplogStats = false
		opts.EnableShardingStats = false
		opts.EnableRollbackStats = false
		opts.EnableLVMSnapshotStats = false
	}

	if hello.Msg == "isdbgrid" {
		// mongos
		if opts.EnableLVMSnapshotStats {
			opts.Logger.Warnf("Disabling snapshot stats because this is a mongos")
			opts.EnableLVMSnapshotStats = false
		}
		if opts.EnableRollbackStats {
			opts.Logger.Warnf("Disabling rollback stats because this is a mongos")
			opts.EnableRollbackStats = false
		}
		if opts.EnableOplogStats {
			opts.Logger.Warnf("Disabling oplog stats because this is a mongos")
			opts.EnableOplogStats = false
		}
		if opts.EnableCurrentopMetrics {
			opts.Logger.Warnf("Disabling currentop metrics because this is a mongos")
			opts.EnableCurrentopMetrics = false
		}
		if opts.EnableTopMetrics {
			opts.Logger.Warnf("Disabling top stats because this is a mongos")
			opts.EnableTopMetrics = false
		}
	}

	cmdLineOpts, err := mongoutils.GetCmdLineOpts(ctx, client)
	if err != nil {
		return fmt.Errorf("failed to get command line options: %w", err)
	}

	if cmdLineOpts.Parsed.Sharding.ClusterRole != "configsvr" {
		if opts.EnableShardingStats {
			opts.Logger.Warnf("Disabling sharding stats because this is not a config server")
			opts.EnableShardingStats = false
		}
	}

	return nil
}

func validateLocalhostOpts(opts *Opts) error {
	cs, _ := connstring.Parse(opts.URI)
	isLocalhost := true

	for _, host := range cs.Hosts {
		if !(strings.HasPrefix(host, "localhost") || strings.HasPrefix(host, "127.0.0.1")) {
			opts.Logger.Errorf("Remote host %s detected in %s", host, opts.URI)
			isLocalhost = false
			break
		}
	}

	if !isLocalhost {
		if opts.EnableLVMSnapshotStats {
			opts.Logger.Warn("LVMSnapshotStats is not supported for remote MongoDB, disabling it")
			opts.EnableLVMSnapshotStats = false
		}

		if opts.EnableRollbackStats {
			opts.Logger.Warn("RollbackStats is not supported for remote MongoDB, disabling it")
			opts.EnableRollbackStats = false
		}
	} else {
		if opts.LVMSnapshotBackupDir == "" {
			opts.Logger.Errorf("LVMSnapshotBackupDir should be set for localhost")
			os.Exit(1)
		}
	}

	return nil
}
