#!/bin/bash
set -euo pipefail

# Load common envs (PASSLESS_ENTRY, username, ...)
# keep this so it can reuse your existing variables
source "$(dirname "$0")/common.sh"

usage() {
    cat <<EOF
Usage: $0 <local_script_path> <hosts_file>
  <local_script_path>  : local file to copy to remote hosts (required)
  <hosts_file>         : file with one host per line. Each line can be:
                           - IP (e.g. 10.0.0.5)
                           - user@IP (e.g. ubuntu@10.0.0.5)
                       (required)
Environment:
  REMOTE_DEST_DIR      : destination directory on remote hosts (default: /home/chn)
  SSH_OPTS             : extra ssh/scp options (optional)
Example:
  $0 ./my_script.sh hosts.txt
EOF
}

if [ $# -lt 2 ]; then
    usage
    exit 1
fi

local_script="$1"
hosts_file="$2"
REMOTE_DEST_DIR="${REMOTE_DEST_DIR:-/home/chn}"
SSH_KEY="${PASSLESS_ENTRY:-}"
SSH_OPTS="${SSH_OPTS:-"-o StrictHostKeyChecking=no"}"

# sanity checks
if [ ! -f "$local_script" ]; then
    echo "Error: local file '$local_script' does not exist." >&2
    exit 2
fi

if [ ! -f "$hosts_file" ]; then
    echo "Error: hosts file '$hosts_file' does not exist." >&2
    exit 3
fi

if [ -z "$SSH_KEY" ]; then
    echo "Warning: PASSLESS_ENTRY not set. Assuming agent or default key will be used."
fi

# read hosts into array, skip empty lines and comments
mapfile -t hosts < <(grep -Ev '^\s*(#|$)' "$hosts_file" || true)
if [ ${#hosts[@]} -eq 0 ]; then
    echo "Error: no hosts found in $hosts_file" >&2
    exit 4
fi

echo "Copying '$local_script' to ${#hosts[@]} hosts (dest: $REMOTE_DEST_DIR)..."

# helper to perform copy and ensure dest dir exists
copy_to_host() {
    local host="$1"              # can be 'user@ip' or 'ip'
    local scp_target_dir="$REMOTE_DEST_DIR"
    local ssh_target
    local scp_cmd
    local ssh_cmd

    # normalize ssh target: if user is not provided, use $username from common.sh if set
    if [[ "$host" == *@* ]]; then
        ssh_target="$host"
    else
        if [ -n "${username:-}" ]; then
            ssh_target="${username}@${host}"
        else
            ssh_target="${host}"
        fi
    fi

    # ensure dest dir exists (use ssh)
    if [ -n "$SSH_KEY" ]; then
        ssh_cmd=(ssh -i "$SSH_KEY" $SSH_OPTS "$ssh_target")
        scp_cmd=(scp -i "$SSH_KEY" $SSH_OPTS "$local_script" "${ssh_target}:${scp_target_dir}/")
    else
        ssh_cmd=(ssh $SSH_OPTS "$ssh_target")
        scp_cmd=(scp $SSH_OPTS "$local_script" "${ssh_target}:${scp_target_dir}/")
    fi

    # create dest dir (suppress output but capture exit code)
    if ! "${ssh_cmd[@]}" "sudo mkdir -p '$scp_target_dir' && sudo chown \$(whoami):\$(whoami) '$scp_target_dir' >/dev/null 2>&1"; then
        echo "[FAIL] $ssh_target: failed to create dest dir $scp_target_dir" >&2
        return 1
    fi

    # copy file
    if "${scp_cmd[@]}"; then
        echo "[OK]   $ssh_target: copied to $scp_target_dir/$(basename "$local_script")"
        return 0
    else
        echo "[FAIL] $ssh_target: scp failed" >&2
        return 2
    fi
}

# iterate and run copies in background
pids=()
hosts_failed=0
for host in "${hosts[@]}"; do
    echo "Starting copy -> $host"
    copy_to_host "$host" &
    pids+=("$!")
done

# wait for all background copies
for pid in "${pids[@]}"; do
    if wait "$pid"; then
        : # success printed by function
    else
        hosts_failed=$((hosts_failed + 1))
    fi
done

if [ "$hosts_failed" -eq 0 ]; then
    echo "All copies finished successfully."
    exit 0
else
    echo "$hosts_failed host(s) failed to receive the file." >&2
    exit 5
fi
