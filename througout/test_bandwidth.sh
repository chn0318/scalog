#!/bin/bash
NUM_PROCESSES=5

if [ $# -eq 1 ]; then
  NUM_PROCESSES=$1
fi

mkdir -p logs

for i in $(seq 1 $NUM_PROCESSES)
do
  echo "Starting throughput process $i..."
  ./single_client > "logs/throughput_$i.log" 2>&1 &
done

wait

total_latency=0
total_logs=0

for i in $(seq 1 $NUM_PROCESSES)
do
  logfile="logs/throughput_${i}.log"
  echo "Processing $logfile..."
  client_latency=$(grep "Total execution time" "$logfile" | awk '{print $4}')
  logs_written=$(grep "Total logs written" "$logfile" | awk '{print $4}')

  if [[ -n "$client_latency" && -n "$logs_written" ]]; then
    total_latency=$((total_latency + client_latency))
    total_logs=$((total_logs + logs_written))
  else
    echo "Error: Could not parse $logfile"
    continue  
  fi
done

# 计算总体结果
if [[ $NUM_PROCESSES -gt 0 ]]; then
  echo "Total Logs Written: $total_logs"

  if [[ $total_logs -gt 0 ]]; then
    overall_avg_latency=$(echo "scale=2; $total_latency / $total_logs" | bc)
    echo "Average Latency: $overall_avg_latency ms"

    avg_bandwidth_per_sec=$((total_logs / 10))
    echo "Total Bandwidth: $avg_bandwidth_per_sec writes/sec"
  else
    echo "Error: No logs were written."
  fi
else
  echo "Error: Number of processes is zero or undefined."
fi
