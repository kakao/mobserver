# Collectors
This directory contains the collectors that are used to collect data from given MongoDB.

## Available Collectors
- CurrentOp Collector
- Oplog status Collector
- Replication status Collector
- Sharding status Collector
- Top command Collector
- Rollback status Collector
- LVM snapshot status Collector

## Explanation
### 1. CurrentOp Collector
CurrentOp collector collects the current operation details from the given MongoDB from the [currentOp](https://www.mongodb.com/docs/manual/reference/operator/aggregation/currentOp/#-currentop--aggregation-) command.
It can be used to collect slow queries, long-running queries, and other operations that are currently running on the MongoDB.
Metrics ends with _total such as `slow_query_count_total` are total values for a server. Otherwise, it will export with the label `database` and `collection`.

The collector collects below metrics:
- slow_query_count(_total): The number of running slow queries.
- longest_running_query_secs(_total): The longest running query in seconds.
- collscan_count(_total): The number of running collscan queries.
- waiting_for_lock_count(_total): The number of queries that are waiting for a lock.
- waiting_for_latch_count(_total): The number of queries that are waiting for a latch.
- waiting_for_flow_control_count(_total): The number of queries that are waiting for flow control.
- transaction_count(_total): The number of running transactions.


Query example:
```javascript
db.adminCommand({aggregate: "currentOp", pipeline: [{$match: ...}]})
```
Result example:
```javascript
{
    cursor: {
        firstBatch: [
            {
                op: "query",
                microsecs_running: 1000000,
                secs_running: 1,
                ns: "test.test",
                command: {
                    find: "test",
                    filter: {
                        _id: 1
                    }
                },
                msg: "some message",
                planSummary: "IDHACK",
                waitingForLock: false,
                waitingForLatch: null,
                waitingForFlowControl: false,
                transaction: null,
            },
            ...
        ]
    }
}
```
source code: [currentop.go #L38](currentop.go#L38)


### 2. Oplog status Collector
Oplog status collector collects the oplog status from [local.oplog.rs](https://www.mongodb.com/docs/manual/reference/local-database/#mongodb-data-local.oplog.rs). It will be automatically disabled if the given MongoDB is mongos.

The collector collects below metrics:
- logSizeMB: The size of the oplog in megabytes.
- usedMB: The used size of the oplog in bytes.
- firstTs: The timestamp of the first entry in the oplog.
- lastTs: The timestamp of the last entry in the oplog.
- timeDiff: The time difference between the first and last entry in the oplog.

`timeDiff` metric is useful when you should know about minimum time range for replication. It can be used to check available PITR (Point-in-Time Recovery) or minimum replication lag.

Query example:
```javascript
use config
var firstElem = db.oplog.rs.find().sort({$natural: 1}).limit(1)
var lastElem = db.oplog.rs.find().sort({$natural: -1}).limit(1)

var firstTs = firstElem[0].ts
var lastTs = lastElem[0].ts
var timeDiff = lastTs.t - firstTs.t
```

source code: [oplog.go #L37](oplog.go#L37)

### 3. Replication status Collector
Replication status collector collects the replication status from the given MongoDB.

The collector collects below metrics:
- [heartbeat_delay](https://www.mongodb.com/docs/v7.0/reference/command/replSetGetStatus/#mongodb-data-replSetGetStatus.members-n-.lastHeartbeatRecv): The delay of the heartbeat in seconds.
- [lag](https://www.mongodb.com/docs/v7.0/reference/command/replSetGetStatus/#mongodb-data-replSetGetStatus.members-n-.optimeDate): The replication lag in seconds.
- odd_state: Check the server's state is odd or not. Odd state means the server is not in PRIMARY, SECONDARY or STARTUP2 state.
- [elected_before_secs](https://www.mongodb.com/docs/v7.0/reference/command/replSetGetStatus/#mongodb-data-replSetGetStatus.members-n-.electionDate): The time in seconds since the last election.
- [version](https://www.mongodb.com/docs/v7.0/reference/replica-configuration/#mongodb-rsconf-rsconf.version): The version of the replica set.
- [term](https://www.mongodb.com/docs/v7.0/reference/replica-configuration/#mongodb-rsconf-rsconf.term): The term of the replica set.
- [protocolVersion](https://www.mongodb.com/docs/v7.0/reference/replica-configuration/#mongodb-rsconf-rsconf.protocolVersion): The protocol version of the replica set.
- [arbiterOnly](https://www.mongodb.com/docs/v7.0/reference/replica-configuration/#mongodb-rsconf-rsconf.members-n-.arbiterOnly): The arbiterOnly status of the replica set.
- [buildIndexes](https://www.mongodb.com/docs/v7.0/reference/replica-configuration/#mongodb-rsconf-rsconf.members-n-.buildIndexes): The buildIndexes status of the replica set.
- [hidden](https://www.mongodb.com/docs/v7.0/reference/replica-configuration/#mongodb-rsconf-rsconf.members-n-.hidden): The hidden status of the replica set.
- [priority](https://www.mongodb.com/docs/v7.0/reference/replica-configuration/#mongodb-rsconf-rsconf.members-n-.priority): The priority of the replica set.
- [votes](https://www.mongodb.com/docs/v7.0/reference/replica-configuration/#mongodb-rsconf-rsconf.members-n-.votes): The votes of the replica set.
- role: The server's replica role of the replica set. It will be exported with the label `role`, which value can be `primary`, `secondary`, `other`.

Query example:
```javascript
var replStatus = db.adminCommand({replSetGetStatus: 1}, {initialSync: 1})
var replConfig = db.adminCommand({replSetGetConfig: 1})
```

Query outputs are shown in [replSetGetStatus](https://www.mongodb.com/docs/v7.0/reference/command/replSetGetStatus/#replsetgetstatus) and [replSetGetConfig](https://www.mongodb.com/docs/v7.0/reference/command/replSetGetConfig/#replsetgetconfig) documents.

source code: [repl.go #L35](repl.go#L35)

### 4. Sharding status Collector
Sharding status collector collects the sharding status from [configsvr](https://www.mongodb.com/docs/manual/reference/config-database/#config-database). Sharding status can be collected from config servers. This gives an overview of the sharding status

The collector collects below metrics:
- sharded_databases: The number of sharded databases.
- unsharded_databases: The number of unsharded databases.
- balancer_enabled: The balancer status of the sharding.
- shards: The number of shards.
- draining_shards: The number of draining shards.
- chunks: The number of chunks. Chunks will be exported with the label `database`, `collection` and `shard`.
- last_24h_chunk_moves: The number of chunk moves in the last 24 hours. Chunk moves will be exported with the label `database` and `collection`.

source_codes: [sharding.go #L32](sharding.go#L32)

### 5. Top command Collector
Top command collector collects the [top command](https://www.mongodb.com/docs/manual/reference/command/top/#top) details from the given MongoDB. It can be used to collect the top command details such as `query`, `insert`, `update`, `remove`, `getmore` and `command`.

The collector collects below metrics:
- insert_count: The number of insert commands.
- insert_time: The time of insert commands.
- queries_count: The number of query commands.
- queries_time: The time of query commands.
- update_count: The number of update commands.
- update_time: The time of update commands.
- remove_count: The number of remove commands.
- remove_time: The time of remove commands.
- getmore_count: The number of getmore commands.
- getmore_time: The time of getmore commands.
- commands_count: The number of command commands.
- commands_time: The time of command commands.

Query example:
```javascript
db.adminCommand({top: 1})
```

source code: [top.go #L34](top.go#L34)

### 6. Rollback status Collector
Rollback status collector collects the rollback status from the given MongoDB. It observers the rollback files of each collections.
[Rollback](https://www.mongodb.com/docs/manual/core/replica-set-rollbacks/#rollback-data) is a process that restores the data to a previous state.
MongoDB will automatically generate rollback files in {{dbpath}}/rollback directory when the rollback occurs. The collector will check the rollback files of each collection. Collector will export with the label `database` and `collection`.

The collector collects below metrics:
- rollback_directory: The directory of the rollback files for each collection.

source code: [rollback.go #L36](rollback.go#L36)

### 7. LVM snapshot status Collector
LVM snapshot status collector collects the percentage of the used LVM snapshot space. If the LVM snapshot space reaches 100%, your backup will fail.

The collector collects below metrics:
- snapshot_allocation: The percentage of the used LVM snapshot space.

Used command:
```bash
df | awk '/<your snapshot area>$/'
sudo lvs | awk '$6!~/[^0-9.]/&&$6>0{print$6}'
```

source code: [lvm.go #L38](lvm.go#L38)


### 8. Instance status Collector
Instance status collector collects the binary version of the MongoDB instance.

The collector collects below metrics:
- version: The binary version of the MongoDB instance.

Query example:
```javascript
db.adminCommand({buildInfo: 1})
```
source code: [instance.go #L39](instance.go#L39)