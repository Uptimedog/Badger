sq add badger.db --handle @badger

sq inspect @badger

sq ping --all

sq '@badger | .options | .[0:]' -j

sq '@badger | .sessions | .[0:]' -j

sq '@badger | .users | .[0:]' -j

sq '@badger | .tasks | .[0:]' -j

sq '@badger | .logs | .[0:]' -j

sq '@badger | .hosts | .[0:]' -j

sq '@badger | .host_groups | .[0:]' -j

sq '@badger | .deployments | .[0:]' -j

sq '@badger | .teams | .[0:]' -j

rhino serve -c api.mock.json

