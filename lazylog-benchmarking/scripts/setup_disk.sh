#!/bin/bash

# mount /dev/sda4 and create fs
# sudo mkfs.ext4 /dev/sda4
sudo chown -R $USER /data
sudo rm -rf /data/*