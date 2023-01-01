#!/bin/bash

function deps {
    echo "Installing dependencies ..."

    apt-get update
    apt-get upgrade -y

    echo "Installing dependencies done!"
}

function badger {
    echo "Installing badger ..."

    mkdir -p /etc/badger
    cd /etc/badger
    LATEST_VERSION=$(curl --silent "https://api.github.com/repos/uptimedog/badger/releases/latest" | jq '.tag_name' | sed -E 's/.*"([^"]+)".*/\1/' | tr -d v)
    curl -sL https://github.com/uptimedog/badger/releases/download/v{$LATEST_VERSION}/badger_Linux_x86_64.tar.gz | tar xz

    echo "[Unit]
Description=Badger
Documentation=https://github.com/Uptimedog/Badger

[Service]
ExecStart=/etc/badger/badger server -c /etc/badger/config.prod.yml
Restart=on-failure
RestartSec=2

[Install]
WantedBy=multi-user.target" > /etc/systemd/system/badger.service

    systemctl daemon-reload
    systemctl enable badger.service
    systemctl start badger.service

    echo "Badger installation done!"
}

deps
badger
