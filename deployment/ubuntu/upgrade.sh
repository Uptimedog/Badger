#!/bin/bash

function badger {
    echo "Upgrade badger ..."

    cd /etc/badger
    mv config.prod.yml config.bk.yml

    LATEST_VERSION=$(curl --silent "https://api.github.com/repos/uptimedog/badger/releases/latest" | jq '.tag_name' | sed -E 's/.*"([^"]+)".*/\1/' | tr -d v)

    curl -sL https://github.com/uptimedog/badger/releases/download/v{$LATEST_VERSION}/badger_Linux_x86_64.tar.gz | tar xz

    rm config.prod.yml
    mv config.bk.yml config.prod.yml

    systemctl restart badger

    echo "Badger upgrade done!"
}

badger
