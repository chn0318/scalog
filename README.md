# Scalog

Scalog is a distributed shared log system that consist from order layer, data layer and discovery node.
This project provides a command-line tool `scalogctl` for managing and running a Scalog cluster.

## Quick Start

### 1. Prerequisites

- Go 1.22 or later
- Docker installed on every node listed in `.scalog.yaml`
- All nodes accessible via SSH

### 2. Configure the Cluster

Create a `.scalog.yaml` file on one node:

```yaml
order-port: 26733
raft-port: 27238
order-replication-factor: 3
order-batching-interval: 1ms
order-0-ip: "192.168.0.3"
order-1-ip: "192.168.0.4"
order-2-ip: "192.168.0.5"

data-port: 23282
data-replication-factor: 2
data-batching-interval: 1ms
data-0-0-ip:  "192.168.0.6"
data-0-1-ip:  "192.168.0.7"
data-1-0-ip:  "192.168.0.8"
data-1-1-ip:  "192.168.0.9"

disc-port: 23472
disc-ip: "127.0.0.1"
```

### 3. Enter the scalog project directory

```
cd /path/to/scalog
```

All the following steps are performed in the scalog directory.

------

## Using `scalogctl`

1. Install the CLI tool:

   ```bash
   go install -mod=vendor ./scalogctl
   ```

2. Start the cluster:

   ```bash
   scalogctl start --config {path to .scalog.yaml}
   ```

3. Stop the cluster:

   ```bash
   scalogctl stop --config {path to .scalog.yaml}
   ```

4. Run the client or performance tester:

   ```bash
    docker run -it --network=host -v /path/to/.scalog.yaml:/root/.scalog.yaml chn0318/scalog:v2.0 client
    
    docker run -it --network=host -v /path/to/.scalog.yaml:/root/.scalog.yaml chn0318/scalog:v2.0 perf -t {thread number}
   ```

------

## Building Docker Images

### Use Prebuilt Images

Docker images are available at docker hub

```bash
docker pull chn0318/scalog:v2.0
```

### Build Locally

```bash
docker build -t {image_name} .
```

------

## Build Scalog Locally

```bash
go build -mod=vendor .
```

This will generate a binary named `scalog`.



> *Note:* Original README starts from here
---

# scalog
reimplementing scalog from scratch

## Build Scalog

Run the `go build` command to build Scalog, and `go test -v ./...` to run
unit-tests.

## Run the code

To run the server side code, we have to install `goreman` by running
```go
go get github.com/mattn/goreman
```

Use `goreman start` to run the server side code. The goreman configuration is
in `Procfile` and the Scalog configuration file is in `.scalog.yaml`

To run the client, we should use
```go
./scalog client --config .scalog.yaml
```

After that, we can use command `append [record]` to append a record to the
log, and use `read [GlobalSequenceNumber] [ShardID]` to read a record.
