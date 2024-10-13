#!/bin/bash

# 可配置的并发进程数量
NUM_PROCESSES=5

# 你可以通过命令行参数指定进程数量，或者使用默认的值
if [ $# -eq 1 ]; then
  NUM_PROCESSES=$1
fi

# 创建日志文件夹
mkdir -p logs

# 启动指定数量的 throughput 进程，并将它们的输出重定向到不同的日志文件
for i in $(seq 1 $NUM_PROCESSES)
do
  echo "Starting throughput process $i..."
  ./single_client > "logs/throughput_$i.log" 2>&1 &
done

# 等待所有后台进程完成
wait

echo "All throughput processes have completed."

# 初始化统计变量
total_latency=0
total_logs=0

# 处理每个日志文件
for i in $(seq 1 $NUM_PROCESSES)
do
  logfile="logs/throughput_${i}.log"
  echo "Processing $logfile..."

  # 从日志文件中读取平均延迟和总写入日志条目数
  avg_latency=$(grep "Average latency" "$logfile" | awk '{print $3}')
  logs_written=$(grep "Total logs written" "$logfile" | awk '{print $4}')

  # 如果找到了相应的数据，累加到总和
  if [ -n "$avg_latency" ] && [ -n "$logs_written" ]; then
    total_latency=$(echo "$total_latency + $avg_latency" | bc)
    total_logs=$((total_logs + logs_written))
  else
    echo "Error: Could not parse $logfile"
  fi
done

# 计算并输出总的平均延迟
if [ $NUM_PROCESSES -gt 0 ]; then
  overall_avg_latency=$(echo "scale=2; $total_latency / $NUM_PROCESSES" | bc)
  echo "Overall Average Latency: $overall_avg_latency ms"
fi

# 输出总的写入日志条目数
echo "Total Logs Written: $total_logs"
