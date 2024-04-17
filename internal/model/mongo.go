package model

// HelloDoc is a response model from hello command
type HelloDoc struct {
	// replica set name
	SetName string `bson:"setName"`

	// msg will be "isdbgrid" if the server is a mongos
	// https://www.mongodb.com/docs/manual/core/sharded-cluster-query-router/#confirm-connection-to-mongos-instances
	Msg string `bson:"msg"`

	Primary string `bson:"primary"`
	Me      string `bson:"me"`

	// hosts is a list of the hostnames and ports of the members of the replica set
	Hosts []string `bson:"hosts"`

	// arbiter flag
	ArbiterOnly bool `bson:"arbiterOnly"`
}

// CmdLineOptsDoc is a response model from getCmdLineOpts command
type CmdLineOptsDoc struct {
	Parsed struct {
		OperationProfiling struct {
			SlowOpThresholdMs int `bson:"slowOpThresholdMs"`
		}
		Storage struct {
			DBPath string `bson:"dbPath"`
		} `bson:"storage"`

		Replication struct {
			ReplSetName string `bson:"replSetName"`
		} `bson:"replication"`

		Sharding struct {
			ClusterRole string `bson:"clusterRole"`
		} `bson:"sharding"`
	} `bson:"parsed"`
}
