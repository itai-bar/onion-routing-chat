#!/bin/bash
SERVER_CONTAINER_IP=$(docker ps | grep 'torbasedchat_chat_server' | awk '{print $1;}' | xargs docker inspect | grep '"IPAddress"' | grep -o '[0-9]\{1,3\}\.[0-9]\{1,3\}\.[0-9]\{1,3\}\.[0-9]\{1,3\}')

echo "172.20.0.2 $SERVER_CONTAINER_IP"
python3 client/tester.py 172.20.0.2 $SERVER_CONTAINER_IP