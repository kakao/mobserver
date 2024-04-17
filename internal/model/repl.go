package model

import (
	"time"
)

type ReplSetGetStatusDoc struct {
	Date    time.Time            `bson:"date"`
	MyState MongoReplRoleType    `bson:"myState"`
	Members []*RSStatusMemberDoc `bson:"members"`
	Set     string               `bson:"set"`
}

type RSStatusMemberDoc struct {
	Name string `bson:"name"`

	State MongoReplRoleType `bson:"state"`

	OpTimeDate        time.Time `bson:"optimeDate"`
	Self              bool      `bson:"self"`
	ElectionDate      time.Time `bson:"electionDate"`
	LastHeartbeatRecv time.Time `bson:"lastHeartbeatRecv"`

	PingMilliseconds int32 `bson:"pingMs"`
}

type ReplSetGetConfigDoc struct {
	Config *RSConfigDoc `bson:"config"`
}

type RSConfigDoc struct {
	Version         int32                `bson:"version"`
	Term            int32                `bson:"term"`
	Members         []*RSConfigMemberDoc `bson:"members"`
	ProtocolVersion int64                `bson:"protocolVersion"`
}

type RSConfigMemberDoc struct {
	Host         string `bson:"host"`
	ArbiterOnly  bool   `bson:"arbiterOnly"`
	BuildIndexes bool   `bson:"buildIndexes"`
	Hidden       bool   `bson:"hidden"`
	Priority     int32  `bson:"priority"`
	Votes        int32  `bson:"votes"`
}
