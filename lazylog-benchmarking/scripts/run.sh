#!/bin/bash

source $(dirname $0)/common.sh

benchmark_dir=$(realpath $(dirname $0)/..)
LOGDIR="/data"

# index into remote_nodes/ips for order nodes
order=("node0" "node1" "node2")

# index into remote_nodes/ips for data shards
data_primary=("node3" "node5" "node7" "node9" "node11")
data_secondary=("node4" "node6" "node8" "node10" "node12")

client_nodes=("node13" "node14" "node15")

batching_intervals=("0.1ms")

modify_batching_intervals() {
    sed -i "s|order-batching-interval: .*|order-batching-interval: $1|" "${benchmark_dir}/../.scalog.yaml"
    sed -i "s|data-batching-interval: .*|data-batching-interval: $1|" "${benchmark_dir}/../.scalog.yaml"
}

clear_server_logs() {
    # mount storage and clear existing logs if any
    sudo ./run_script_on_servers.sh ./setup_disk.sh
}

clear_client_logs() {
    # mount storage and clear existing logs if any
    sudo ./run_script_on_clients.sh ./setup_disk.sh
}

cleanup_servers() {
    # kill existing servers
    sudo ./run_script_on_servers.sh ./kill_all_goreman.sh
}

cleanup_clients() {
    # kill existing clients
    sudo ./run_script_on_clients.sh ./kill_all_benchmark.sh
}

drop_server_caches() {
    sudo ./run_script_on_servers.sh ./drop_caches.sh
}

start_order_nodes() {
    # start order nodes
    for ((i=0; i<=2; i++))
    do
        echo "Starting order-${i} on ${order[$i]}"
        ssh -o StrictHostKeyChecking=no -i $PASSLESS_ENTRY $username@${order[$i]} "sh -c \"cd $benchmark_dir/order-$i; nohup sudo ./run_goreman.sh > ${LOGDIR}/order-$i.log 2>&1 &\""
    done
}


# args: num shards
start_data_nodes() {
    # start data nodes
    for ((i=0; i<$1; i++))
    do
        echo "Starting data-${i}-0 on ${data_primary[$i]}"
        ssh -o StrictHostKeyChecking=no -i $PASSLESS_ENTRY $username@${data_primary[$i]} "sh -c \"cd $benchmark_dir/data-$i-0; nohup sudo ./run_goreman.sh > ${LOGDIR}/data-$i-0.log 2>&1 &\""
    done

    for ((i=0; i<$1; i++))
    do
        echo "Starting data-${i}-1 on ${data_secondary[$i]}"
        ssh -o StrictHostKeyChecking=no -i $PASSLESS_ENTRY $username@${data_secondary[$i]} "sh -c \"cd $benchmark_dir/data-$i-1; nohup sudo ./run_goreman.sh > ${LOGDIR}/data-$i-1.log 2>&1 &\""
    done
}

start_discovery() {
    # start discovery
    echo "Starting discovery on ${data_primary[0]}"
    ssh -o StrictHostKeyChecking=no -i $PASSLESS_ENTRY $username@${data_primary[0]} "sh -c \"cd $benchmark_dir/disc; nohup sudo ./run_goreman.sh > ${LOGDIR}/disc.log 2>&1 &\""
}

check_data_log() {
    for ((i=0; i<=4; i++))
    do
        echo "Checking data node data-$i-0..."
        ssh -o StrictHostKeyChecking=no -i $PASSLESS_ENTRY $username@${data_primary[$i]} "grep error ${LOGDIR}/data-$i-0.log"
    done

    for ((i=0; i<=4; i++))
    do
        echo "Checking data node data-$i-1..."
        ssh -o StrictHostKeyChecking=no -i $PASSLESS_ENTRY $username@${data_secondary[$i]} "grep error ${LOGDIR}/data-$i-1.log"
    done
}

start_append_clients() {
    ssh -o StrictHostKeyChecking=no -i $PASSLESS_ENTRY $username@$1 "cd $benchmark_dir/scripts; sudo ./run_append_client.sh $2 $3 $1 $4 $5 > ${LOGDIR}/client_$1.log 2>&1" &
}

start_random_read_clients() {
    ssh -o StrictHostKeyChecking=no -i $PASSLESS_ENTRY $username@$1 "cd $benchmark_dir/scripts; sudo ./run_random_read_client.sh $2 $3 $1 $4 $5 $6 > ${LOGDIR}/client_$1.log 2>&1" &
}

start_sequential_read_clients() {
    ssh -o StrictHostKeyChecking=no -i $PASSLESS_ENTRY $username@$1 "cd $benchmark_dir/scripts; sudo ./run_sequential_read_client.sh $2 $3 $1 $4 $5 $6 > ${LOGDIR}/client_$1.log 2>&1" &
}

load_phase() {
    sudo /usr/local/go/bin/go run load.go $1 $2 $3 $4 
}


# mode 
#   0 -> append experiment mode
#   1 -> read experiment mode
#   2 -> setup servers
#   3 -> kill server and client, read server logs for errors
#   4 -> clear server and client logs

mode="$1"
if [ "$mode" -eq 0 ]; then # append experiment mode
    clients_per_shard=90
    num_shards=("1" "5")
    for interval in "${batching_intervals[@]}";
    do
        # modify intervals
        modify_batching_intervals $interval

        for s in "${num_shards[@]}"; 
        do
            cleanup_clients
            cleanup_servers
            clear_server_logs
            clear_client_logs

            start_order_nodes
            start_data_nodes $s
            start_discovery

            # wait for 10 secs
            sleep 10

            c=$(($s * $clients_per_shard))
            num_client_nodes=${#client_nodes[@]}
            high_num=$((($c + $num_client_nodes - 1)/$num_client_nodes))
            low_num=$(($c / $num_client_nodes))
            mod=$(($c % $num_client_nodes))

            for (( i=0; i<num_client_nodes; i++))
            do
                if [ "$i" -lt "$mod" ]; then
                    # If there's a remainder, assign one additional job to the first 'mod' clients
                    num_jobs_for_client=$((low_num + 1))
                else
                    num_jobs_for_client=$low_num
                fi
                
                # start_append_clients <client_id> <num_of_clients_to_run> <num_appends_per_client> <total_clients> <interval>
                start_append_clients "${client_nodes[$i]}" $num_jobs_for_client "3m" $c $interval
            done

            echo "Waiting for clients to terminate"
            wait

            cleanup_clients
            cleanup_servers

            # check for errors in log files
            check_data_log
        done
    done
elif [ "$mode" -eq 1 ]; then # read experiment mode
    clients=("1")
    for interval in "${batching_intervals[@]}";
    do
        # modify intervals
        modify_batching_intervals $interval

        cleanup_clients
        cleanup_servers
        clear_server_logs
        clear_client_logs

        start_order_nodes
        start_data_nodes 
        start_discovery

        # wait for 10 secs
        sleep 10

        load_phase "4096" "2000000" "10" "gsnToShardMap.txt"

        echo "Done with loading"

        for c in "${clients[@]}"; 
        do 
            drop_server_caches 

            num_client_nodes=${#client_nodes[@]}
            high_num=$((($c + $num_client_nodes - 1)/$num_client_nodes))
            low_num=$(($c / $num_client_nodes))
            mod=$(($c % $num_client_nodes))

            for (( i=0; i<num_client_nodes; i++))
            do
                if [ "$i" -lt "$mod" ]; then
                    # If there's a remainder, assign one additional job to the first 'mod' clients
                    num_jobs_for_client=$((low_num + 1))
                else
                    num_jobs_for_client=$low_num
                fi

                # start_append_clients <client_id> <num_of_clients_to_run> <num_appends_per_client> <total_clients> <input_filename>
                start_sequential_read_clients "${client_nodes[$i]}" $num_jobs_for_client "3m" $c $interval "gsnToShardMap.txt"
            done

            echo "Waiting for clients to terminate"
            wait

            cleanup_clients
        done
        
        cleanup_servers
        # check for errors in log files
        check_data_log
    done
elif [ "$mode" -eq 2 ]; then # setup servers mode
    cleanup_clients
    cleanup_servers
    clear_server_logs
    clear_client_logs

    start_order_nodes
    start_data_nodes 
    start_discovery
elif [ "$mode" -eq 3 ]; then # kill servers and clients 
    cleanup_clients
    cleanup_servers

    # check for errors in log files
    check_data_log
elif [ "$mode" -eq 4 ]; then 
    cleanup_clients
    cleanup_servers
    # collect_logs
else # cleanup logs
    clear_server_logs
    clear_client_logs
fi
