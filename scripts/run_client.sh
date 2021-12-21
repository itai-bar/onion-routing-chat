#!/bin/bash
NODES_CONTAINERS_IDS=$(docker ps | grep 'torbasedchat_node' | awk '{print $1;}')
NODES_CONTAINERS_IPS=$(echo "$NODES_CONTAINERS_IDS" | xargs docker inspect | grep '"IPAddress"' | grep -o '[0-9]\{1,3\}\.[0-9]\{1,3\}\.[0-9]\{1,3\}\.[0-9]\{1,3\}' | tr "\n" " ")

SERVER_CONTAINER_IP=$(docker ps | grep 'torbasedchat_chat_server' | awk '{print $1;}' | xargs docker inspect | grep '"IPAddress"' | grep -o '[0-9]\{1,3\}\.[0-9]\{1,3\}\.[0-9]\{1,3\}\.[0-9]\{1,3\}')

echo "$NODES_CONTAINERS_IPS$SERVER_CONTAINER_IP"
python3 client/client.py $NODES_CONTAINERS_IPS $SERVER_CONTAINER_IP
