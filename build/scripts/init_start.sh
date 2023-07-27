#!/bin/bash

# 本脚本用于在rpm安装完成后，启动dms服务

systemctl daemon-reload
systemctl start dms.service
for i in {1..10}; do
	systemctl status dms.service &>/dev/null
    if [  $? -eq 0 ]; then
        echo "Init and start dms success!"
        exit 0
    fi
    sleep 1
done

echo "init and start dms fail! Please check journalctl -u dms"