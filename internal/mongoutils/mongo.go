package mongoutils

import (
	"context"
	"fmt"
	"mobserver/internal/model"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ConnectionOpts struct {
	URI              string
	User             string
	Password         string
	DirectConnect    bool
	ConnectTimeoutMS int64
}

func Connect(ctx context.Context, opts *ConnectionOpts) (*mongo.Client, error) {
	clientOpts := options.Client().ApplyURI(opts.URI)
	if opts.User != "" || opts.Password != "" {
		clientOpts.SetAuth(options.Credential{
			Username: opts.User,
			Password: opts.Password,
		})
	}

	clientOpts.SetDirect(opts.DirectConnect)
	clientOpts.SetAppName("mobserver")

	if clientOpts.ConnectTimeout == nil && opts.ConnectTimeoutMS > 0 {
		connectTimeout := time.Duration(opts.ConnectTimeoutMS) * time.Millisecond
		clientOpts.SetConnectTimeout(connectTimeout)
		clientOpts.SetServerSelectionTimeout(connectTimeout)
	}

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, fmt.Errorf("invalid MongoDB options: %w", err)
	}

	if err = client.Ping(ctx, nil); err != nil {
		// Ping failed. Close background connections. Error is ignored since the ping error is more relevant.
		_ = client.Disconnect(ctx)

		return nil, fmt.Errorf("cannot connect to MongoDB: %w", err)
	}

	return client, nil
}

func GetHello(ctx context.Context, client *mongo.Client) (*model.HelloDoc, error) {
	var result model.HelloDoc
	cmd := bson.D{{Key: "hello", Value: 1}}

	if err := client.Database("admin").RunCommand(ctx, cmd).Decode(&result); err != nil {
		return nil, fmt.Errorf("cannot run hello command: %w", err)
	}

	return &result, nil
}

func GetCmdLineOpts(ctx context.Context, client *mongo.Client) (*model.CmdLineOptsDoc, error) {
	var result model.CmdLineOptsDoc
	cmd := bson.D{{Key: "getCmdLineOpts", Value: 1}}

	if err := client.Database("admin").RunCommand(ctx, cmd).Decode(&result); err != nil {
		return nil, fmt.Errorf("cannot run getCmdLineOpts command: %w", err)
	}

	return &result, nil
}

func GetReplStatus(ctx context.Context, client *mongo.Client) (*model.ReplSetGetStatusDoc, error) {
	var result model.ReplSetGetStatusDoc
	cmd := bson.D{{Key: "replSetGetStatus", Value: 1}, {Key: "initialSync", Value: 1}}

	if err := client.Database("admin").RunCommand(ctx, cmd).Decode(&result); err != nil {
		return nil, fmt.Errorf("cannot run replSetGetStatus command: %w", err)
	}

	return &result, nil
}

func GetReplConfig(ctx context.Context, client *mongo.Client) (*model.ReplSetGetConfigDoc, error) {
	var result model.ReplSetGetConfigDoc
	cmd := bson.D{{Key: "replSetGetConfig", Value: 1}}

	if err := client.Database("admin").RunCommand(ctx, cmd).Decode(&result); err != nil {
		return nil, fmt.Errorf("cannot run replSetGetConfig command: %w", err)
	}

	return &result, nil
}

func GetDatabases(ctx context.Context, client *mongo.Client) ([]string, error) {
	var result model.ListDatabasesDoc
	cmd := bson.D{{Key: "listDatabases", Value: 1}}

	if err := client.Database("admin").RunCommand(ctx, cmd).Decode(&result); err != nil {
		return nil, fmt.Errorf("cannot run listDatabases command: %w", err)
	}

	dbs := make([]string, 0, len(result.Databases))
	for _, db := range result.Databases {
		dbs = append(dbs, db.Name)
	}

	return dbs, nil
}

func GetCollections(ctx context.Context, client *mongo.Client, db string) ([]model.Collection, error) {
	var result model.ListCollectionsDoc
	cmd := bson.D{{Key: "listCollections", Value: 1}}

	if err := client.Database(db).RunCommand(ctx, cmd).Decode(&result); err != nil {
		return nil, fmt.Errorf("cannot run listCollections command: %w", err)
	}

	colls := make([]model.Collection, 0)
	for _, coll := range result.Cursor.FirstBatch {
		if coll.Type == "collection" {
			colls = append(colls, model.Collection{
				Name: coll.Name,
				Type: coll.Type,
				Info: coll.Info,
			})
		}
	}

	return colls, nil
}

func GetAllDatabasesAndCollections(ctx context.Context, client *mongo.Client) (map[string]string, error) {
	res := make(map[string]string)

	dbs, err := GetDatabases(ctx, client)
	if err != nil {
		return nil, err
	}

	for _, db := range dbs {
		colls, err := GetCollections(ctx, client, db)
		if err != nil {
			return nil, err
		}

		for _, coll := range colls {
			binUUID, ok := coll.Info.UUID.(primitive.Binary)
			if !ok {
				return nil, fmt.Errorf("expected response field _id to be type primitive.Binary, but is type %T", coll.Info.UUID)
			}
			uid, err := uuid.FromBytes(binUUID.Data)
			if err != nil {
				return nil, err
			}
			res[uid.String()] = fmt.Sprintf("%s.%s", db, coll.Name)
		}
	}

	return res, nil
}
