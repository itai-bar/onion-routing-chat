#!/bin/bash

ARGC=$#
NODES_AMOUNT=$1
LOG_MODE=$2

if [ $ARGC != 2 ]; then
    echo "Usage: ./run_tor_network [nodes amount] [log mode - (0: no logs, 1: tor and chat, 2: only chat)]"
    exit 1
fi

if [ $LOG_MODE == 0 ]; then 
    NODE_LOG_ARG="0"
    ROUTER_LOG_ARG="0"
    CHAT_LOG_ARG="0"
fi

if [ $LOG_MODE == 1 ]; then 
    NODE_LOG_ARG="1"
    ROUTER_LOG_ARG="1"
    CHAT_LOG_ARG="1"
fi

if [ $LOG_MODE == 2 ]; then 
    NODE_LOG_ARG="0"
    ROUTER_LOG_ARG="0"
    CHAT_LOG_ARG="1"
fi

if [ $NODES_AMOUNT -lt 3 ]; then
    echo "The network needs at least 3 nodes to start"
    exit 1
fi

echo "Cleaning old build"
docker-compose down > /dev/null 2>&1
echo "Building the services"
docker-compose build \
    --build-arg NODE_LOG=$NODE_LOG_ARG \
    --build-arg ROUTER_LOG=$ROUTER_LOG_ARG \
    --build-arg CHAT_LOG=$CHAT_LOG_ARG \
    > /dev/null 2>&1

docker-compose up --scale node=$NODES_AMOUNT

# echo "Saving Database"
docker cp torbasedchat_chat_server_1:/app/db.sqlite . # saving DB to host for next executions

# cleanup
echo "Cleaning the build"
docker-compose down > /dev/null 2>&1
echo "Removing unused docker data"
echo "y" | docker system prune > /dev/null 2>&1 