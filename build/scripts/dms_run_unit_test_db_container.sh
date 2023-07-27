#!/bin/bash
set -e

MYSQL_CONTAINER_NAME=$1

if test ! -z "$(docker ps -a | grep $MYSQL_CONTAINER_NAME)";then
    docker rm -f $MYSQL_CONTAINER_NAME
fi

docker run --privileged --name $MYSQL_CONTAINER_NAME -p 33306:3306 -e MYSQL_ROOT_PASSWORD=123 -e MYSQL_DATABASE=dms_unittest -d mysql:8.0
while (! docker logs $MYSQL_CONTAINER_NAME  2>&1 | grep -q "mysqld: ready for connections.") || [[ "$SECONDS" -gt 180 ]] ; 
do
    sleep 1
    echo "time($SECONDS) wait mysql start..."
done
echo "mysql start success or wait timeout"