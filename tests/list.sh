#!/usr/bin/env bash 
# Lists host:port for every running sshbox container, for SSH from outside.

for cid in $(docker compose ps -q sshbox); do
    name=$(docker inspect -f '{{.Name}}' "$cid" | sed 's#^/##')
    hostport=$(docker port "$cid" 22/tcp)
    echo "$name    $hostport"
done


