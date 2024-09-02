# Scalog Benchmarking for LazyLog

## Requirements
To run experiments, we need a 16 node machine (xl170/c6525-25g machine) with the data folder mounted at `/data`.

## Setup
Install dependencies and setup go
```
cd lazylog-benchmarking/scripts
./run_script_on_all.sh ./install_go.sh
./run_script_on_all.sh ./init_disk.sh
```
The `run_script_on_all.sh` script runs the provided script on all nodes (`node0` - `node16`); `install_go.sh` installs go and additional dependencies, `init_disk.sh` ensures that the `/data` folder is owned by the current user. 

## Obtaining data for Figure 7
Run the following 
```
cd lazylog-benchmarking/scripts
./run.sh 0
```

This starts an append-only benchmark on scalog that runs for two sets of parameters sequentially, (1) a single shard setup where we have 90 append clients appending data continuously for 3 mins, (2) a 5 shard setup with 90*5 clients appending data continuously for 3 mins. The number of clients were picked after running a latency-throughput exploration for a single shard which showed that beyond 100 clients the append latency drastically increases with increasing number of clients. The number of clients for the 5 shard experiments is chosen so that each shard roughly gets the same load as it does in the single shard experiments. After 6 minutes, this experiment terminates after creating per-client latency dumps in `lazylog-benchmarking/results/0.1ms/`. To analyze the results and print them in a readable format, run the following

```
cd lazylog-benchmarking/scripts

# Note this step assumes pip installation with numpy already exists, (if not run `sudo apt-get install pip; pip install numpy`)
python3 analyze.py 
```
**Ensure that the `lazylog-benchmarking/results` folder is cleared before any re-runs of the above experiment**

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
