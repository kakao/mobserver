# Mobserver
Mobserver is a monitoring tool for MongoDB. It collects metrics from MongoDB and exports them to Prometheus. It provides metrics that mongodb_exporter does not provide such as oplog, rollback, and so on. See [mobserver](exporter/README.md) for more details.

Note: Supported MongoDB version is 4.4 or later.

## How to build
Build process needs go 1.16 or later.
```bash
$ make build
```

## How to run
```bash
$ ./bin/mobserver \
    --web.listen-address=:9100 \
    --collect-all \
    --mongodb.global-conn-pool \
    --mongodb.uri=$MONGO_URI \
    --mongodb.user=$MONGO_USER \
    --mongodb.password=$MONGO_PASSWORD \
    --lvm-backup-dir=/backup
```

## Usage
| Flag | Description | Default | Example |
| ---- | ----------- | ------- | ------- |
| web.listen-address | Address on which to expose metrics and web interface | :9100 | :9100 |
| web.telemetry-path | Path under which to expose metrics | /metrics | /metrics |
| web.config | Path to the file having Prometheus TLS config for basic auth | - | file path |
| web.timeout-offset | Offset to subtract from the request timeout in seconds | 1 | 1 |
| mongodb.uri | MongoDB connection URI | - | mongodb://localhost:27017/?directConnection=true |
| mongodb.user | monitor user, need clusterMonitor role in admin db and read role in local db | - | monitorUser |
| mongodb.password | monitor user password | - | monitorPassword |
| log.level | Only log messages with the given severity or above | debug | [debug, info, warn, error, fatal] |
| [no-]mongodb.global-conn-pool | Use global connection pool instead of creating new pool for each http request. | false | - |
| [no-]mongodb.direct-connect | Whether or not a direct connect should be made. Direct connections are not valid if multiple hosts are specified or an SRV URI is used. | true | - |
| mongodb.connect-timeout-ms | Connection timeout in milliseconds | 5000 | 1000 |
| collector.replicasetstatus | Enable collecting metrics from replSetGetStatus | false | - |
| collector.topmetrics | Enable collecting metrics from top admin command | false | - |
| collector.currentopmetrics | Enable collecting metrics currentop admin command | false | - |
| collector.oplogstats | Enable collecting metrics from oplog | false | - |
| collector.shardstats | Enable collecting metrics from shard | false | - |
| collector.lvmsnapshotstats | Enable collecting metrics from lvs | false | - |
| collector.rollbackstats | Enable collecting metrics from rollback | false | - |
| collect-all | Collect all metrics | false | true |
| lvm-backup-dir | Collect all metrics | - | /data/lvm-snapshot-backup-dir |
| enable-currentop-store | Enable storing currentop metrics | false | - |
| version | Show version and exit | - | - |

## License

This software is licensed under the [Apache 2 license](LICENSE), quoted below.

Copyright 2024 Kakao Corp. <http://www.kakaocorp.com>

Licensed under the Apache License, Version 2.0 (the "License"); you may not
use this project except in compliance with the License. You may obtain a copy
of the License at http://www.apache.org/licenses/LICENSE-2.0.

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
License for the specific language governing permissions and limitations under
the License.